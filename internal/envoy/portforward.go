package envoy

import (
	"context"
	"fmt"
	"net"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const envoyAdminPort = 19000

// PortForwardResult holds the local address and a stop function.
type PortForwardResult struct {
	LocalAddr string // e.g. "http://localhost:12345"
	Stop      func()
}

// PortForwardToGateway finds a ready gateway-proxy pod and opens a port-forward
// to the Envoy admin port. Caller must call Stop() when done.
func PortForwardToGateway(ctx context.Context, gatewayName, namespace, kubeContext string) (*PortForwardResult, error) {
	cfg, err := buildRestConfig(kubeContext)
	if err != nil {
		return nil, err
	}

	kc, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %w", err)
	}

	podName, err := findGatewayProxyPod(ctx, kc, gatewayName, namespace)
	if err != nil {
		return nil, err
	}

	localPort, err := freePort()
	if err != nil {
		return nil, fmt.Errorf("finding free local port: %w", err)
	}

	stopCh := make(chan struct{})
	readyCh := make(chan struct{})
	errCh := make(chan error, 1)

	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating SPDY transport: %w", err)
	}

	url := kc.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward").URL()

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)
	ports := []string{fmt.Sprintf("%d:%d", localPort, envoyAdminPort)}

	fw, err := portforward.New(dialer, ports, stopCh, readyCh, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("creating port-forwarder: %w", err)
	}

	go func() {
		errCh <- fw.ForwardPorts()
	}()

	select {
	case <-readyCh:
	case err := <-errCh:
		return nil, fmt.Errorf("port-forward failed: %w", err)
	case <-ctx.Done():
		close(stopCh)
		return nil, ctx.Err()
	}

	return &PortForwardResult{
		LocalAddr: fmt.Sprintf("http://localhost:%d", localPort),
		Stop:      func() { close(stopCh) },
	}, nil
}

func buildRestConfig(kubeContext string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	if kubeContext != "" {
		overrides.CurrentContext = kubeContext
	}

	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, overrides,
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig: %w", err)
	}
	return cfg, nil
}

// findGatewayProxyPod finds the first ready pod for a kgateway Gateway.
func findGatewayProxyPod(ctx context.Context, kc kubernetes.Interface, gatewayName, namespace string) (string, error) {
	selector := fmt.Sprintf("gateway.networking.k8s.io/gateway-name=%s", gatewayName)
	pods, err := kc.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return "", fmt.Errorf("listing gateway-proxy pods: %w", err)
	}

	for _, pod := range pods.Items {
		if isPodReady(&pod) {
			return pod.Name, nil
		}
	}

	return "", fmt.Errorf(
		"no ready gateway-proxy pod found for Gateway %s/%s (selector: %s)\n"+
			"Hint: ensure the gateway pod is running and you have RBAC permissions for 'get pods' and 'create pods/portforward' in namespace %s",
		namespace, gatewayName, selector, namespace,
	)
}

func isPodReady(pod *corev1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

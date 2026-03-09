package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "kfp",
		Short: "Kgateway filter chain printer — visualize Envoy config",
	}

	dump := &cobra.Command{
		Use:   "dump",
		Short: "Dump and visualize the Envoy filter chain configuration",
		RunE:  runDump,
	}

	// Input source flags (mutually exclusive)
	dump.Flags().String("file", "", "Path to an Envoy config_dump JSON file")
	dump.Flags().String("gateway", "", "Gateway name (fetches config via port-forward to gateway-proxy pod)")
	dump.Flags().StringP("namespace", "n", "default", "Namespace of the Gateway (used with --gateway)")
	dump.Flags().String("context", "", "Kubeconfig context (used with --gateway, default: current context)")

	root.AddCommand(dump)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runDump(cmd *cobra.Command, args []string) error {
	file, _ := cmd.Flags().GetString("file")
	gateway, _ := cmd.Flags().GetString("gateway")

	if file == "" && gateway == "" {
		return fmt.Errorf("specify either --file <path> or --gateway <name>")
	}
	if file != "" && gateway != "" {
		return fmt.Errorf("--file and --gateway are mutually exclusive")
	}

	fmt.Println("kfp dump — not yet implemented")
	return nil
}

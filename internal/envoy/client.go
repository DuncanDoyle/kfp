package envoy

import (
	"fmt"
	"io"
	"net/http"
)

// FetchConfigDump fetches /config_dump from the given Envoy admin base URL
// and returns the raw JSON bytes.
func FetchConfigDump(baseURL string) ([]byte, error) {
	resp, err := http.Get(baseURL + "/config_dump")
	if err != nil {
		return nil, fmt.Errorf("GET /config_dump failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Envoy admin returned HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading config_dump response: %w", err)
	}
	return data, nil
}

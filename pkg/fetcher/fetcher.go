package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	Timeout   = 60 * time.Second
	UserAgent = "russia-blocked-ips/1.0"
	MaxRetry  = 3
)

// Get fetches URL body with retry and timeout
func Get(url string) ([]byte, error) {
	client := &http.Client{Timeout: Timeout}

	var lastErr error
	for i := range MaxRetry {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", UserAgent)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("fetcher: %s returned %d", url, resp.StatusCode)
			continue
		}

		return body, nil
	}

	return nil, fmt.Errorf("fetcher: %s failed after %d retries: %w", url, MaxRetry, lastErr)
}

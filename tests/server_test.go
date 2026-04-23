package tests

import (
	"net/http"
	"testing"
)

func TestStartAndStopTestServer(t *testing.T) {
	// No t.Parallel(): the helper binds a fixed process-wide port.
	server := startTestServer()
	defer stopTestServer(server)

	resp, err := http.Get("http://127.0.0.1:1234/")
	if err != nil {
		t.Fatalf("http.Get() error = %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("resp.Body.Close() error = %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

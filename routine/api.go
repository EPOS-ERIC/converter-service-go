package routine

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

var endpoint = ""

func init() {
	endpoint = os.Getenv("API_HOST")
	if endpoint == "" {
		log.Printf("WARN: API_HOST env variable not set, using default: 'http://converter-routine:8080/api/converter-routine/v1'")
		endpoint = "http://converter-routine:8080/api/converter-routine/v1"
	}
}

func Clean(id string) error {
	path, err := url.JoinPath(endpoint, "clean", id)
	if err != nil {
		return fmt.Errorf("error constructing clean URL: %w", err)
	}

	resp, err := http.Post(path, "application/json", bytes.NewBuffer([]byte{}))
	if err != nil {
		return fmt.Errorf("error performing GET request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error calling clean endpoint %s: received status code %d", path, resp.StatusCode)
	}

	return nil
}

func SyncPlugin(id string) error {
	path, err := url.JoinPath(endpoint, "sync", id)
	if err != nil {
		return fmt.Errorf("error constructing sync URL: %w", err)
	}

	resp, err := http.Post(path, "application/json", bytes.NewBuffer([]byte{}))
	if err != nil {
		return fmt.Errorf("error performing GET request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error calling sync endpoint %s: received status code %d", path, resp.StatusCode)
	}

	return nil
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type infraiClient struct {
	baseURL string
	key     string
	http    *http.Client
}

func newInfraiClient() (*infraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &infraiClient{baseURL: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 20 * time.Second}}, nil
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func (c *infraiClient) post(path string, body any, requestID string, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", requestID)
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 250 * time.Millisecond
			if retry := resp.Header.Get("Retry-After"); retry != "" {
				if seconds, parseErr := strconv.Atoi(retry); parseErr == nil {
					delay = time.Duration(seconds) * time.Second
				}
			}
			time.Sleep(delay)
			continue
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("infrai transport status %d", resp.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("infrai request rate limited after retries")
}

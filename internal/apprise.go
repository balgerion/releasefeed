package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type appriseClient struct {
	baseURL string
	client  *http.Client
}

func NewApprise(baseURL string) *appriseClient {
	return &appriseClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type apprisePayload struct {
	Title string   `json:"title"`
	Body  string   `json:"body"`
	Tag   []string `json:"tag,omitempty"`
}

func (a *appriseClient) Notify(r Release, tags []string) error {
	p := apprisePayload{
		Title: fmt.Sprintf("[%s] %s", r.FeedName, r.Title),
		Body:  r.URL,
		Tag:   tags,
	}
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	resp, err := a.client.Post(a.baseURL+"/notify", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("apprise returned %d", resp.StatusCode)
	}
	return nil
}

func HasKeyword(content string, keywords []string) bool {
	lower := strings.ToLower(content)
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

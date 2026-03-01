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

func (a *appriseClient) Notify(r Release, tags []string, matched []string) error {
	body := r.URL
	if len(matched) > 0 {
		body += "\nkeywords: " + strings.Join(matched, ", ")
	}
	p := apprisePayload{
		Title: fmt.Sprintf("[%s] %s", r.FeedName, r.Title),
		Body:  body,
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

func MatchedKeywords(content string, keywords []string) []string {
	lower := strings.ToLower(content)
	var matched []string
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			matched = append(matched, kw)
		}
	}
	return matched
}
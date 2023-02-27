package morningpost

import (
	"fmt"
	"io"
	"net/http"
)

type HackerNewsClient struct {
	HTTPClient *http.Client
	HTTPHost   string
	URI        string
}

func NewHackerNewsClient() *HackerNewsClient {
	return &HackerNewsClient{
		HTTPClient: &http.Client{},
		HTTPHost:   "https://news.ycombinator.com",
		URI:        "rss",
	}
}

func (h HackerNewsClient) GetNews() ([]News, error) {
	resp, err := h.HTTPClient.Get(fmt.Sprintf("%s/%s", h.HTTPHost, h.URI))
	if err != nil {
		return nil, fmt.Errorf("cannot get %q: %+v", h.HTTPHost, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status %q", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ParseRSSResponse(data)
}

package morningpost

import (
	"fmt"
	"io"
	"net/http"
)

type HackerNewsClient struct {
	HTTPClient *http.Client
	HTTPHost   string
}

func NewHackerNewsClient() *HackerNewsClient {
	return &HackerNewsClient{
		HTTPClient: &http.Client{},
		HTTPHost:   "https://news.ycombinator.com",
	}
}

func (hn HackerNewsClient) GetNews() ([]News, error) {
	resp, err := hn.HTTPClient.Get(fmt.Sprintf("%s/rss", hn.HTTPHost))
	if err != nil {
		return nil, fmt.Errorf("cannot get %q: %+v", hn.HTTPHost, err)
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

package morningpost

import (
	"fmt"
	"io"
	"net/http"
)

type HNClient struct {
	HTTPClient *http.Client
	HTTPHost   string
}

func (hn HNClient) GetNews() ([]News, error) {
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

func NewHNClient() *HNClient {
	return &HNClient{
		HTTPClient: &http.Client{},
		HTTPHost:   "https://news.ycombinator.com",
	}
}

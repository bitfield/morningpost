package morningpost

import (
	"fmt"
	"io"
	"net/http"
)

type CNNClient struct {
	HTTPClient *http.Client
	HTTPHost   string
	URI        string
}

func NewCNNClient() *CNNClient {
	return &CNNClient{
		HTTPClient: &http.Client{},
		HTTPHost:   "http://rss.cnn.com",
		URI:        "rss/cnn_topstories.rss",
	}
}

func (c CNNClient) GetNews() ([]News, error) {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/%s", c.HTTPHost, c.URI))
	if err != nil {
		return nil, err
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

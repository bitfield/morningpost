package sources

import (
	"fmt"
	"io"
	"net/http"

	"github.com/thiagonache/morningpost"
)

type BITClient struct {
	HTTPClient *http.Client
	HTTPHost   string
}

func (b BITClient) GetNews() ([]morningpost.News, error) {
	resp, err := b.HTTPClient.Get(fmt.Sprintf("%s/golang?format=rss", b.HTTPHost))
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
	return morningpost.ParseRSSResponse(data)
}

func NewBITClient() *BITClient {
	return &BITClient{
		HTTPClient: &http.Client{},
		HTTPHost:   "https://bitfieldconsulting.com",
	}
}

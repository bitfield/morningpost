package morningpost

import (
	"fmt"
	"io"
	"net/http"
)

type TechCrunchClient struct {
	HTTPClient *http.Client
	HTTPHost   string
	URI        string
}

func NewTechCrunchClient() *TechCrunchClient {
	return &TechCrunchClient{
		HTTPClient: &http.Client{},
		HTTPHost:   "https://techcrunch.com",
		URI:        "feed/",
	}
}

func (t TechCrunchClient) GetNews() ([]News, error) {
	resp, err := t.HTTPClient.Get(fmt.Sprintf("%s/%s", t.HTTPHost, t.URI))
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

package morningpost

import (
	"fmt"
	"io"
	"net/http"
)

type TechCrunchClient struct {
	HTTPClient *http.Client
	HTTPHost   string
}

func NewTechCrunchClient() *TechCrunchClient {
	return &TechCrunchClient{
		HTTPClient: &http.Client{},
		HTTPHost:   "https://techcrunch.com",
	}
}

func (tc TechCrunchClient) GetNews() ([]News, error) {
	resp, err := tc.HTTPClient.Get(fmt.Sprintf("%s/feed/", tc.HTTPHost))
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

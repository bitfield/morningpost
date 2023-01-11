package morningpost

import (
	"encoding/xml"
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
	return ParseTCResponse(data)
}

func ParseTCResponse(input []byte) ([]News, error) {
	type rss struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Items []struct {
				Title string `xml:"title"`
				Link  string `xml:"link"`
			} `xml:"item"`
		} `xml:"channel"`
	}
	r := rss{}
	err := xml.Unmarshal(input, &r)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal data %q: %w", input, err)
	}
	news := make([]News, len(r.Channel.Items))
	for i, item := range r.Channel.Items {
		news[i] = News{
			Title: item.Title,
			URL:   item.Link,
		}
	}
	return news, err
}

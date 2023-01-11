package morningpost

import (
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
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

func ParseHNResponse(r io.Reader) ([]News, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("error reading content: %+v", err)
	}
	news := []News{}
	doc.Find(".titleline").Each(func(i int, s *goquery.Selection) {
		n := News{Title: s.Text()}
		href, ok := s.Find("a").Attr("href")
		if ok {
			n.URL = href
		}
		news = append(news, n)
	})
	return news, nil
}

package morningpost

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"sync"

	"github.com/pkg/browser"
	"golang.org/x/sync/errgroup"
)

const (
	defaultShowMaxNews  = 20
	defaultTemplatePath = "output.html.tpl"
	serverListenAddr    = ":33000"
)

type Source interface {
	GetNews() ([]News, error)
}

type News struct {
	Title string
	URL   string
}

type MorningPost struct {
	PageNews           []News
	ShowMaxNews        int
	Sources            []Source
	template           *template.Template
	TemplatePath       string
	loadDefaultSources bool

	mu   sync.Mutex
	News []News
}

func (m *MorningPost) GetNews() error {
	g := new(errgroup.Group)
	for _, source := range m.Sources {
		src := source
		g.Go(func() error {
			news, err := src.GetNews()
			if err != nil {
				return err
			}
			m.mu.Lock()
			m.News = append(m.News, news...)
			m.mu.Unlock()
			return nil
		})
	}
	return g.Wait()
}

func (m *MorningPost) RandomNews() error {
	totalNews := len(m.News)
	if m.ShowMaxNews > totalNews {
		return fmt.Errorf("cannot random values: ShowMaxNews (%d) is bigger than News (%d)", m.ShowMaxNews, totalNews)
	}
	randomIndexes := rand.Perm(totalNews)[:m.ShowMaxNews]
	for _, idx := range randomIndexes {
		m.PageNews = append(m.PageNews, m.News[idx])
	}
	return nil
}

func (m *MorningPost) WritePageNewsTo(w io.Writer) error {
	return m.template.Execute(w, m.PageNews)
}

func New(opts ...Option) (*MorningPost, error) {
	m := &MorningPost{
		loadDefaultSources: true,
		ShowMaxNews:        defaultShowMaxNews,
		TemplatePath:       defaultTemplatePath,
	}
	for _, o := range opts {
		o(m)
	}
	if m.loadDefaultSources {
		m.Sources = append(m.Sources, []Source{NewHNClient(), NewTGClient(), NewTechCrunchClient(), NewBITClient()}...)
	}
	tpl, err := template.New(m.TemplatePath).ParseFiles(m.TemplatePath)
	if err != nil {
		return nil, err
	}
	m.template = tpl
	return m, nil
}

func GetRSSFeed(url string) ([]News, error) {
	resp, err := http.Get(url)
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

func ParseRSSResponse(input []byte) ([]News, error) {
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

func Main() int {
	m, err := New()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	err = m.GetNews()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	err = m.RandomNews()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	buf := &bytes.Buffer{}
	err = m.WritePageNewsTo(buf)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	err = browser.OpenReader(buf)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

func Server() int {
	m, err := New()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	err = m.GetNews()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		err := m.RandomNews()
		if err != nil {
			fmt.Fprintln(w, err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		err = m.WritePageNewsTo(w)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
		}
	}
	fmt.Printf("Starting server on %s\n", serverListenAddr)
	if err := http.ListenAndServe(serverListenAddr, http.HandlerFunc(handler)); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

type Option func(*MorningPost)

func WithDefaultSources(load bool) Option {
	return func(m *MorningPost) {
		m.loadDefaultSources = load
	}
}

package morningpost

import (
	"bytes"
	"embed"
	"encoding/xml"
	"flag"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/sync/errgroup"
)

//go:embed output.html.tpl
var content embed.FS

const (
	defaultOutput      = "browser"
	defaultShowMaxNews = 20
	serverListenAddr   = ":33000"
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
	showMaxNews        int
	Sources            []Source
	template           *template.Template
	loadDefaultSources bool
	Output             string

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
	if m.ShowMaxNews() > totalNews {
		return fmt.Errorf("cannot random values: ShowMaxNews (%d) is bigger than News (%d)", m.ShowMaxNews(), totalNews)
	}
	randomIndexes := rand.Perm(totalNews)[:m.ShowMaxNews()]
	for i, idx := range randomIndexes {
		m.PageNews[i] = m.News[idx]
	}
	return nil
}

func (m *MorningPost) WritePageNewsTo(w io.Writer) error {
	return m.template.Execute(w, m.PageNews)
}

func (m *MorningPost) ShowMaxNews() int {
	return m.showMaxNews
}

func (m *MorningPost) String() string {
	output := []string{}
	for _, news := range m.PageNews {
		output = append(output, fmt.Sprintf("%s [%s]", news.Title, news.URL))
	}
	return strings.Join(output, "\n")
}

func New(opts ...Option) (*MorningPost, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	m := &MorningPost{
		loadDefaultSources: true,
		showMaxNews:        defaultShowMaxNews,
		Output:             defaultOutput,
	}
	for _, o := range opts {
		o(m)
	}
	m.PageNews = make([]News, m.ShowMaxNews())
	if m.loadDefaultSources {
		m.Sources = append(m.Sources, []Source{NewHackerNewsClient(), NewTechCrunchClient()}...)
	}
	tpl, err := template.ParseFS(content, "output.html.tpl")
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

func FromArgs(args []string) Option {
	return func(m *MorningPost) error {
		fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		output := fs.String("o", "browser", "where to send news [browser|terminal]")
		err := fs.Parse(args)
		if err != nil {
			return err
		}
		m.Output = *output
		return nil
	}
}

func Main() int {
	m, err := New(FromArgs(os.Args[1:]))
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
	if m.Output == "terminal" {
		fmt.Println(m)
		return 0
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

type Option func(*MorningPost) error

func WithDefaultSources(load bool) Option {
	return func(m *MorningPost) error {
		m.loadDefaultSources = load
		return nil
	}
}

func WithShowMaxNews(total int) Option {
	return func(m *MorningPost) error {
		m.showMaxNews = total
		return nil
	}
}

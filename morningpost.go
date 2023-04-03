package morningpost

import (
	"bytes"
	"context"
	"embed"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"golang.org/x/sync/errgroup"
)

//go:embed templates/*.gohtml
var templates embed.FS

const (
	DefaultHTTPTimeout = 30 * time.Second
	DefaultListenPort  = 33000
	DefaultShowMaxNews = 50
	FeedTypeRSS        = "RSS"
	FeedTypeAtom       = "Atom"
	FeedTypeRDF        = "RDF"
)

type News struct {
	Feed  string
	Title string
	URL   string
}

func NewNews(feed, title, URL string) (News, error) {
	if title == "" || URL == "" {
		return News{}, errors.New("empty title or url")
	}
	_, err := url.Parse(URL)
	if err != nil {
		return News{}, err
	}
	return News{
		Feed:  feed,
		Title: title,
		URL:   URL,
	}, nil
}

type MorningPost struct {
	Client      *http.Client
	ListenPort  int
	PageNews    []News
	ShowMaxNews int
	Stderr      io.Writer
	Stdout      io.Writer
	Store       *FileStore

	mu   *sync.Mutex
	News []News
}

func (m *MorningPost) GetNews() error {
	m.EmptyNews()
	g := new(errgroup.Group)
	for _, feed := range m.Store.GetAll() {
		feed := feed
		g.Go(func() error {
			news, err := feed.GetNews()
			if err != nil {
				return fmt.Errorf("%q: %w", feed.Endpoint, err)
			}
			m.AddNews(news)
			return nil
		})

	}
	return g.Wait()
}

func (m *MorningPost) RandomNews() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.News) < m.ShowMaxNews {
		return
	}
	m.PageNews = make([]News, 0, m.ShowMaxNews)
	randomIndexes := rand.Perm(len(m.News))[:m.ShowMaxNews]
	for _, idx := range randomIndexes {
		m.PageNews = append(m.PageNews, m.News[idx])
	}
}

func (m *MorningPost) AddNews(news []News) {
	m.mu.Lock()
	m.News = append(m.News, news...)
	m.mu.Unlock()
}

func (m *MorningPost) EmptyNews() {
	m.mu.Lock()
	m.News = []News{}
	m.mu.Unlock()
}

func (m *MorningPost) ReadURLFromForm(r *http.Request) (string, error) {
	err := r.ParseForm()
	if err != nil {
		return "", err
	}
	url := r.Form.Get("url")
	url = strings.TrimSpace(url)
	if url == "" {
		return "", errors.New("bad Request: please, inform the URL")
	}
	return url, nil
}

func (m *MorningPost) ReadFeedIDFromURI(uri string) string {
	urlParts := strings.Split(uri, "/")
	return urlParts[len(urlParts)-1]
}

func (m *MorningPost) FindFeeds(URL string) ([]Feed, error) {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", "MorningPost/0.1")
	req.Header.Set("accept", "*/*")
	resp, err := m.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status code %q", resp.Status)
	}
	contentType := parseContentType(resp.Header)
	switch contentType {
	case "application/rss+xml", "application/atom+xml", "text/xml", "application/xml":
		type feedType struct {
			XMLName xml.Name
		}
		feedTypeData := feedType{}
		decoder := xml.NewDecoder(resp.Body)
		decoder.CharsetReader = charset.NewReaderLabel
		err := decoder.Decode(&feedTypeData)
		if err != nil {
			return nil, fmt.Errorf("cannot decode data: %w", err)
		}
		switch strings.ToUpper(feedTypeData.XMLName.Local) {
		case "RSS":
			return []Feed{{
				Endpoint: URL,
				Type:     FeedTypeRSS,
			}}, nil
		case "RDF":
			return []Feed{{
				Endpoint: URL,
				Type:     FeedTypeRDF,
			}}, nil
		case "FEED":
			return []Feed{{
				Endpoint: URL,
				Type:     FeedTypeAtom,
			}}, nil
		default:
			return nil, errors.New("unable to detect XMLName in body")
		}
	case "text/html":
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		feeds, err := ParseLinkTags(body, URL)
		if err != nil {
			return nil, err
		}
		return feeds, nil
	default:
		return nil, fmt.Errorf("unexpected content type: %q", contentType)
	}
}

func (m *MorningPost) HandleFeeds(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(m.Stdout, r.Method, r.URL)
	switch r.Method {
	case http.MethodHead:
		w.WriteHeader(http.StatusOK)
	case http.MethodPost:
		URL, err := m.ReadURLFromForm(r)
		if err != nil {
			fmt.Fprintln(m.Stderr, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		feeds, err := m.FindFeeds(URL)
		if err != nil {
			fmt.Fprintln(m.Stderr, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, feed := range feeds {
			m.Store.Add(feed)
		}
		err = RenderHTMLTemplate(w, "templates/feeds.gohtml", m.Store.GetAll())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodGet:
		err := RenderHTMLTemplate(w, "templates/feeds.gohtml", m.Store.GetAll())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		id := m.ReadFeedIDFromURI(r.URL.Path)
		ui64, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m.Store.Delete(ui64)
		err = RenderHTMLTemplate(w, "templates/feeds.gohtml", m.Store.GetAll())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		fmt.Fprintln(m.Stderr, "Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (m *MorningPost) HandleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(m.Stdout, r.Method, r.URL)
	if r.RequestURI != "/" {
		fmt.Fprintf(m.Stderr, "%s not found\n", r.RequestURI)
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		err := m.GetNews()
		if err != nil {
			fmt.Fprintln(m.Stderr, err.Error())
		}
		m.RandomNews()
		err = RenderHTMLTemplate(w, "templates/home.gohtml", m.PageNews)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		fmt.Fprintln(m.Stderr, "Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func New(store *FileStore, opts ...Option) (*MorningPost, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	m := &MorningPost{
		Client: &http.Client{
			Timeout: DefaultHTTPTimeout,
		},
		ListenPort:  DefaultListenPort,
		mu:          &sync.Mutex{},
		ShowMaxNews: DefaultShowMaxNews,
		Stderr:      os.Stderr,
		Stdout:      os.Stdout,
		Store:       store,
	}
	for _, o := range opts {
		err := o(m)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func getenv(key string) string {
	v, _ := syscall.Getenv(key)
	return v
}

func userStateDir() string {
	switch runtime.GOOS {
	case "windows":
		dir := getenv("AppData")
		if dir == "" {
			return "./"
		}
		return dir
	case "darwin", "ios":
		dir := getenv("HOME")
		if dir == "" {
			return "./"
		}
		dir += "/Library/Application Support"
		return dir
	default: // Unix
		dir := getenv("XDG_STATE_HOME")
		if dir == "" {
			return "/var/lib"
		}
		return dir
	}
}

func Main() int {
	fileStore, err := NewFileStore()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	m, err := New(fileStore,
		FromArgs(os.Args[1:]),
	)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	http.HandleFunc("/", m.HandleHome)
	http.HandleFunc("/feeds/", m.HandleFeeds)
	listenAddr := fmt.Sprintf("0.0.0.0:%d", m.ListenPort)
	httpServer := http.Server{
		Addr: listenAddr,
	}
	idleConnections := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		log.Print("Please WAIT! Do not send the same signal AGAIN!")
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}
		close(idleConnections)
	}()
	log.Printf("Listening at http://%s", listenAddr)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe Error: %v", err)
		return 1
	}
	<-idleConnections
	err = m.Store.Save()
	if err != nil {
		log.Println(err)
		return 1
	}
	log.Printf("Done. Thank you! <3")
	return 0
}

type Feed struct {
	Endpoint string
	ID       uint64
	Type     string
}

func (f Feed) GetNews() ([]News, error) {
	req, err := http.NewRequest(http.MethodGet, f.Endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", "MorningPost/0.1")
	req.Header.Set("accept", "*/*")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	switch f.Type {
	case FeedTypeRSS:
		return ParseRSSResponse(body)
	case FeedTypeRDF:
		return ParseRDFResponse(bytes.NewReader(body))
	case FeedTypeAtom:
		return ParseAtomResponse(body)
	default:
		return nil, fmt.Errorf("unkown feed type %q", f.Type)
	}
}

func parseContentType(headers http.Header) string {
	return strings.Split(headers.Get("content-type"), ";")[0]
}

func ParseLinkTags(data []byte, baseURL string) ([]Feed, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	feeds := []Feed{}
	doc.Find("link[type='application/rss+xml']").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			u, err := url.Parse(href)
			if err != nil {
				return
			}
			feeds = append(feeds, Feed{
				Endpoint: base.ResolveReference(u).String(),
				Type:     FeedTypeRSS,
			})
		}
	})
	doc.Find("link[type='application/atom+xml']").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			u, err := url.Parse(href)
			if err != nil {
				return
			}
			feeds = append(feeds, Feed{
				Endpoint: base.ResolveReference(u).String(),
				Type:     FeedTypeAtom,
			})
		}
	})
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		title, _ := s.Attr("title")
		if strings.Contains(strings.ToLower(title), "rss") {
			href, exists := s.Attr("href")
			if exists {
				u, err := url.Parse(href)
				if err != nil {
					return
				}
				feeds = append(feeds, Feed{
					Endpoint: base.ResolveReference(u).String(),
					Type:     FeedTypeRSS,
				})
			}
		}
	})
	return feeds, nil
}

func RenderHTMLTemplate(w io.Writer, templatePath string, data any) error {
	tpl := template.Must(template.New("main").ParseFS(templates, "templates/base.gohtml", templatePath))
	err := tpl.Execute(w, data)
	if err != nil {
		return err
	}
	return nil
}

func ParseRSSResponse(input []byte) ([]News, error) {
	type rss struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Title string `xml:"title"`
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
	allNews := make([]News, 0, len(r.Channel.Items))
	for _, item := range r.Channel.Items {
		news, err := NewNews(r.Channel.Title, item.Title, item.Link)
		if err != nil {
			continue
		}
		allNews = append(allNews, news)
	}
	return allNews, nil
}

func ParseRDFResponse(r io.Reader) ([]News, error) {
	type rdf struct {
		XMLName xml.Name `xml:"RDF"`
		Channel struct {
			Title string `xml:"title"`
		} `xml:"channel"`
		Items []struct {
			Title string `xml:"title"`
			Link  string `xml:"link"`
		} `xml:"item"`
	}
	rdfData := rdf{}
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.Decode(&rdfData)
	if err != nil {
		return nil, fmt.Errorf("cannot decode data: %w", err)
	}
	allNews := make([]News, 0, len(rdfData.Items))
	for _, item := range rdfData.Items {
		news, err := NewNews(rdfData.Channel.Title, item.Title, item.Link)
		if err != nil {
			continue
		}
		allNews = append(allNews, news)
	}
	return allNews, nil
}

func ParseAtomResponse(input []byte) ([]News, error) {
	type atom struct {
		XMLName xml.Name `xml:"feed"`
		Title   string   `xml:"title"`
		Entries []struct {
			Link struct {
				Href string `xml:"href,attr"`
			} `xml:"link"`
			Title struct {
				Text string `xml:",chardata"`
			} `xml:"title"`
		} `xml:"entry"`
	}
	a := atom{}
	err := xml.Unmarshal(input, &a)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal data %q: %w", input, err)
	}
	allNews := make([]News, 0, len(a.Entries))
	for _, item := range a.Entries {
		news, err := NewNews(a.Title, item.Title.Text, item.Link.Href)
		if err != nil {
			continue
		}
		allNews = append(allNews, news)
	}
	return allNews, nil
}

type Option func(*MorningPost) error

func FromArgs(args []string) Option {
	return func(m *MorningPost) error {
		fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		fs.SetOutput(m.Stderr)
		listenPort := fs.Int("p", DefaultListenPort, "Listening port")
		err := fs.Parse(args)
		if err != nil {
			return err
		}
		m.ListenPort = *listenPort
		return nil
	}
}

func WithStderr(w io.Writer) Option {
	return func(m *MorningPost) error {
		if w == nil {
			return errors.New("standard error cannot be nil")
		}
		m.Stderr = w
		return nil
	}
}

func WithStdout(w io.Writer) Option {
	return func(m *MorningPost) error {
		if w == nil {
			return errors.New("standard stdout cannot be nil")
		}
		m.Stdout = w
		return nil
	}
}

func WithClient(c *http.Client) Option {
	return func(m *MorningPost) error {
		if c == nil {
			return errors.New("HTTP client cannot be nil")
		}
		m.Client = c
		return nil
	}
}

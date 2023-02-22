package morningpost_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/thiagonache/morningpost"
)

var emptyRSSData = []byte("<rss></rss>")

func TestParseRSS_ReturnsExpectedNewsGivenRSSWithTwoNews(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{
		{Title: "RSS Solutions for Restaurants", URL: "http://www.feedforall.com/restaurant.htm"},
		{Title: "RSS Solutions for Schools and Colleges", URL: "http://www.feedforall.com/schools.htm"},
	}
	inputData, err := os.ReadFile("testdata/rss.xml")
	if err != nil {
		t.Fatalf("Cannot read file content: %+v", err)
	}
	got, err := morningpost.ParseRSSResponse(inputData)
	if err != nil {
		t.Fatalf("Cannot parse content %q: %+v", inputData, err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGetRSSFeed_ReturnsExpectedNewsGivenRSSWithTwoNews(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{
		{Title: "RSS Solutions for Restaurants", URL: "http://www.feedforall.com/restaurant.htm"},
		{Title: "RSS Solutions for Schools and Colleges", URL: "http://www.feedforall.com/schools.htm"},
	}
	respData, err := os.ReadFile("testdata/rss.xml")
	if err != nil {
		t.Fatal(err)
	}
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.RequestURI != "/rss" {
			t.Fatal("want URI to be /rss")
		}
		w.Write(respData)
	}))
	defer ts.Close()
	got, err := morningpost.GetRSSFeed(fmt.Sprintf("%s/rss", ts.URL))
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("Request not performed")
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGetRSSFeed_ErrorsIfResponseCodeIsNotHTTPStatusOK(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer ts.Close()
	_, err := morningpost.GetRSSFeed(ts.URL)
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestNewMorningPost_SetProperTemplatePathByDefault(t *testing.T) {
	t.Parallel()
	want := "output.html.tpl"
	m, err := morningpost.New()
	if err != nil {
		t.Fatal(err)
	}
	got := m.TemplatePath
	if want != got {
		t.Fatalf("want template path %q but got %q", want, got)
	}
}

func TestNewMorningPost_SetProperShowMaxNewsByDefault(t *testing.T) {
	t.Parallel()
	want := 20
	m, err := morningpost.New()
	if err != nil {
		t.Fatal(err)
	}
	got := m.ShowMaxNews
	if want != got {
		t.Fatalf("want ShowMaxNews %d but got %d", want, got)
	}
}

type fakeSource struct {
	HTTPHost string
}

func (f fakeSource) GetNews() ([]morningpost.News, error) {
	return morningpost.GetRSSFeed(f.HTTPHost)
}

func TestGetNews_ReturnsExpectedNewsGivenLocalSource(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{
		{Title: "RSS Solutions for Restaurants", URL: "http://www.feedforall.com/restaurant.htm"},
		{Title: "RSS Solutions for Schools and Colleges", URL: "http://www.feedforall.com/schools.htm"},
	}
	respData, err := os.ReadFile("testdata/rss.xml")
	if err != nil {
		t.Fatal(err)
	}
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Write(respData)
	}))
	defer ts.Close()
	fake := fakeSource{
		HTTPHost: ts.URL,
	}
	m, err := morningpost.New(
		morningpost.WithDefaultSources(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	m.Sources = append(m.Sources, fake)
	err = m.GetNews()
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("Request not performed")
	}
	got := m.News
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGetNews_ErrorsIfSourceErrors(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer ts.Close()
	fake := fakeSource{
		HTTPHost: ts.URL,
	}
	m, err := morningpost.New(
		morningpost.WithDefaultSources(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	m.Sources = append(m.Sources, fake)
	err = m.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestRandomNews_ErrorsIfLessNewsThanShowMaxNews(t *testing.T) {
	t.Parallel()
	m, err := morningpost.New()
	if err != nil {
		t.Fatal(err)
	}
	err = m.RandomNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestRandomNews_SetProperNumberPageNewsGivenRSSWithTwoNews(t *testing.T) {
	t.Parallel()
	want := 1
	m, err := morningpost.New(
		morningpost.WithDefaultSources(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	m.News = []morningpost.News{
		{Title: "RSS Solutions for Restaurants", URL: "http://www.feedforall.com/restaurant.htm"},
		{Title: "RSS Solutions for Schools and Colleges", URL: "http://www.feedforall.com/schools.htm"},
	}
	m.ShowMaxNews = 1
	err = m.RandomNews()
	if err != nil {
		t.Fatal(err)
	}
	got := len(m.PageNews)
	if want != got {
		t.Fatalf("want total news %d but got %d", want, got)
	}
}

func TestWritePageNewsTo_RenderProperTemplateContentGivenTwoNews(t *testing.T) {
	t.Parallel()
	want := `<html>
  <title>ThiagosNews</title>
  <body>
    
    <a href="http://www.feedforall.com/restaurant.htm">RSS Solutions for Restaurants</a><br />
    
    <a href="http://www.feedforall.com/schools.htm">RSS Solutions for Schools and Colleges</a><br />
    
  </body>
</html>
`
	m, err := morningpost.New()
	if err != nil {
		t.Fatal(err)
	}
	m.PageNews = []morningpost.News{
		{Title: "RSS Solutions for Restaurants", URL: "http://www.feedforall.com/restaurant.htm"},
		{Title: "RSS Solutions for Schools and Colleges", URL: "http://www.feedforall.com/schools.htm"},
	}
	buf := &bytes.Buffer{}
	err = m.WritePageNewsTo(buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestNewMorningPost_LoadsDefaultSourcesByDefault(t *testing.T) {
	t.Parallel()
	want := []morningpost.Source{
		morningpost.NewHNClient(),
		morningpost.NewTGClient(),
		morningpost.NewTechCrunchClient(),
		morningpost.NewBITClient(),
	}
	m, err := morningpost.New()
	if err != nil {
		t.Fatal(err)
	}
	got := m.Sources
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestWithDefaultSources_SetToFalseLoadsNoSources(t *testing.T) {
	t.Parallel()
	m, err := morningpost.New(
		morningpost.WithDefaultSources(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	if m.Sources != nil {
		t.Fatalf("want sources to be nil but found %#v", m.Sources)
	}
}

package morningpost_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/thiagonache/morningpost"
)

var (
	emptyJSONData = []byte("{}")
	emptyRSSData  = []byte("<rss></rss>")
)

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

func TestParseRSS_ErrorsIfDataIsNotXML(t *testing.T) {
	t.Parallel()
	_, err := morningpost.ParseRSSResponse(emptyJSONData)
	if err == nil {
		t.Fatalf("want error but not found")
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

func TestGetRSSFeed_ErrorsIfHTTPRequestFails(t *testing.T) {
	t.Parallel()
	_, err := morningpost.GetRSSFeed("bogus")
	if err == nil {
		t.Fatal("want error but not found")
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

func TestNew_SetProperShowMaxNewsByDefault(t *testing.T) {
	t.Parallel()
	want := 20
	m, err := morningpost.New()
	if err != nil {
		t.Fatal(err)
	}
	got := m.ShowMaxNews()
	if want != got {
		t.Fatalf("want ShowMaxNews %d but got %d", want, got)
	}
}

func TestNew_SetProperOutputByDefault(t *testing.T) {
	t.Parallel()
	want := "browser"
	m, err := morningpost.New()
	if err != nil {
		t.Fatal(err)
	}
	got := m.Output
	if want != got {
		t.Fatalf("want output %q but got %q", want, got)
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
		morningpost.WithShowMaxNews(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	m.News = []morningpost.News{
		{Title: "RSS Solutions for Restaurants", URL: "http://www.feedforall.com/restaurant.htm"},
		{Title: "RSS Solutions for Schools and Colleges", URL: "http://www.feedforall.com/schools.htm"},
	}
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
	want := `<html lang="en">
  <head>
    <title>ThiagosNews</title>
    <link href="https://maxcdn.bootstrapcdn.com/bootstrap/latest/css/bootstrap.min.css" rel="stylesheet">
  </head>
  <body>
    <div class="list-group">
      <li class="list-group-item"><a href="http://www.feedforall.com/restaurant.htm">RSS Solutions for Restaurants</a></li>
      <li class="list-group-item"><a href="http://www.feedforall.com/schools.htm">RSS Solutions for Schools and Colleges</a></li>
    </div>
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

func TestNew_LoadsDefaultSourcesByDefault(t *testing.T) {
	t.Parallel()
	want := []morningpost.Source{
		morningpost.NewHackerNewsClient(),
		morningpost.NewTechCrunchClient(),
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

func TestNew_SetProperPageNewsSizeByDefault(t *testing.T) {
	t.Parallel()
	want := 20
	m, err := morningpost.New()
	if err != nil {
		t.Fatal(err)
	}
	got := len(m.PageNews)
	if want != got {
		t.Fatalf("want PageNews size of %d but got %d", want, got)
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

func TestWithShowMaxNews_SetPageNewsProperPageNewsSize(t *testing.T) {
	t.Parallel()
	want := 10
	m, err := morningpost.New(
		morningpost.WithShowMaxNews(10),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := len(m.PageNews)
	if want != got {
		t.Fatalf("want PageNews size of %d but got %d", want, got)
	}
}

func TestString_PrintsExpectedMessageGivenTwoNews(t *testing.T) {
	t.Parallel()
	want := `RSS Solutions for Restaurants [http://www.feedforall.com/restaurant.htm]
RSS Solutions for Schools and Colleges [http://www.feedforall.com/schools.htm]`
	m, err := morningpost.New()
	if err != nil {
		t.Fatal(err)
	}
	m.PageNews = []morningpost.News{
		{Title: "RSS Solutions for Restaurants", URL: "http://www.feedforall.com/restaurant.htm"},
		{Title: "RSS Solutions for Schools and Colleges", URL: "http://www.feedforall.com/schools.htm"},
	}
	got := m.String()
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFromArgs_OFlagSetOutput(t *testing.T) {
	t.Parallel()
	want := "terminal"
	args := []string{"-o", "terminal"}
	m, err := morningpost.New(
		morningpost.FromArgs(args),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := m.Output
	if want != got {
		t.Errorf("wants -o flag to set output to %q, got %q", want, got)
	}
}

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"morningpost": morningpost.Main,
	}))
}

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}

package morningpost_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/thiagonache/morningpost"
)

func newServerWithContentTypeResponse(t *testing.T, contentType string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", contentType)
	}))
	t.Cleanup(func() { ts.Close() })
	return ts
}

func newServerWithContentTypeAndBodyResponse(t *testing.T, contentType string, filePath string) *httptest.Server {
	t.Helper()
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", contentType)
		w.Write(data)
	}))
	t.Cleanup(func() { ts.Close() })
	return ts
}

func newMorningPostWithBogusFileStoreAndNoOutput(t *testing.T) *morningpost.MorningPost {
	t.Helper()
	m, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.WithStdout(io.Discard),
		morningpost.WithStderr(io.Discard),
	)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func generateOneThousandsNews(t *testing.T) []morningpost.News {
	t.Helper()
	allNews := []morningpost.News{}
	for x := 0; x < 1000; x++ {
		title := fmt.Sprintf("News #%d", x)
		URL := fmt.Sprintf("http://fake.url/news-%d", x)
		news, err := morningpost.NewNews("Feed Unit test", title, URL)
		if err != nil {
			t.Fatal(err)
		}
		allNews = append(allNews, news)
	}
	return allNews
}

func TestNew_SetsDefaultShowMaxNewsByDefault(t *testing.T) {
	t.Parallel()
	want := morningpost.DefaultShowMaxNews
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	got := m.ShowMaxNews
	if want != got {
		t.Fatalf("want ShowMaxNews %d but got %d", want, got)
	}
}

func TestNew_SetsDefaultListenPortByDefault(t *testing.T) {
	t.Parallel()
	want := morningpost.DefaultListenPort
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	got := m.ListenPort
	if want != got {
		t.Fatalf("want DefaultListenPort %d, got %d", want, got)
	}
}

func TestNew_SetsStderrByDefault(t *testing.T) {
	t.Parallel()
	want := os.Stderr
	m, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.WithStdout(io.Discard),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := m.Stderr
	if want != got {
		t.Fatalf("want stderr %+v, got %+v", want, got)
	}
}

func TestNew_SetsStdoutByDefault(t *testing.T) {
	t.Parallel()
	want := os.Stdout
	m, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.WithStderr(io.Discard),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := m.Stdout
	if want != got {
		t.Fatalf("want stdout %+v, got %+v", want, got)
	}
}

func TestNew_SetsHTTPTimeoutByDefault(t *testing.T) {
	t.Parallel()
	want := morningpost.DefaultHTTPTimeout
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	got := m.Client.Timeout
	if want != got {
		t.Fatalf("want timeout %v, got %v", want, got)
	}
}

func TestFromArgs_ErrorsIfUnkownFlag(t *testing.T) {
	t.Parallel()
	args := []string{"-asdfaewrawers", "8080"}
	_, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.WithStderr(io.Discard),
		morningpost.FromArgs(args),
	)
	if err == nil {
		t.Fatal("want error but got nil")
	}
}

func TestFromArgs_pFlagSetListenPort(t *testing.T) {
	t.Parallel()
	want := 8080
	args := []string{"-p", "8080"}
	m, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.FromArgs(args),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := m.ListenPort
	if want != got {
		t.Errorf("wants -p flag to set listen port to %d, got %d", want, got)
	}
}

func TestFromArgs_SetsListenPortByDefaultToDefaultListenPort(t *testing.T) {
	t.Parallel()
	want := morningpost.DefaultListenPort
	m, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.FromArgs([]string{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := m.ListenPort
	if want != got {
		t.Errorf("wants -p flag to set listen port to %d, got %d", want, got)
	}
}

func TestWithStderr_ErrorsIfInputIsNil(t *testing.T) {
	t.Parallel()
	_, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.WithStderr(nil),
	)
	if err == nil {
		t.Fatal("want error but got nil")
	}
}

func TestWithStderr_SetsStderrGivenWriter(t *testing.T) {
	t.Parallel()
	want := "this is a string\n"
	buf := &bytes.Buffer{}
	m, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.WithStderr(buf),
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(m.Stderr, "this is a string")
	got := buf.String()
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestWithStdout_ErrorsIfInputIsNil(t *testing.T) {
	t.Parallel()
	_, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.WithStdout(nil),
	)
	if err == nil {
		t.Fatal("want error but got nil")
	}
}

func TestWithStdout_SetsStdoutGivenWriter(t *testing.T) {
	t.Parallel()
	want := "this is a string\n"
	buf := &bytes.Buffer{}
	m, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.WithStdout(buf),
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(m.Stdout, "this is a string")
	got := buf.String()
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestWithClient_ErrorsIfInputIsNil(t *testing.T) {
	t.Parallel()
	_, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.WithClient(nil),
	)
	if err == nil {
		t.Fatal("want error but got nil")
	}
}

func TestWithClient_SetsDisableKeepAlivesGivenHTTPClientWithKeepAlivesDisabled(t *testing.T) {
	t.Parallel()
	want := true
	m, err := morningpost.New(
		newFileStoreWithBogusPath(t),
		morningpost.WithClient(&http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := m.Client.Transport.(*http.Transport).Clone().DisableKeepAlives
	if want != got {
		t.Fatalf("want DisableKeepAlives %t, got %t", want, got)
	}
}

func TestParseRSSResponse_ReturnsNewsGivenRSSFeedData(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{
		{
			Feed:  "FeedForAll Sample Feed",
			Title: "RSS Solutions for Restaurants",
			URL:   "http://www.feedforall.com/restaurant.htm",
		},
		{
			Feed:  "FeedForAll Sample Feed",
			Title: "RSS Solutions for Schools and Colleges",
			URL:   "http://www.feedforall.com/schools.htm",
		},
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

func TestParseRSSResponse_SkipItemGivenNewsWithNoURL(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{}
	inputData, err := os.ReadFile("testdata/no-url-item-rss.xml")
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

func TestParseRSSResponse_SkipItemGivenInvalidURL(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{}
	inputData, err := os.ReadFile("testdata/invalid-url-item-rss.xml")
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

func TestParseRSSResponse_SkipItemGivenNoTitle(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{}
	inputData, err := os.ReadFile("testdata/no-title-item-rss.xml")
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

func TestParseRSSResponse_ErrorsIfDataIsNotXML(t *testing.T) {
	t.Parallel()
	_, err := morningpost.ParseRSSResponse([]byte("{}"))
	if err == nil {
		t.Fatalf("want error but not found")
	}
}

func TestParseRDFResponse_ReturnsNewsGivenRDFData(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{
		{
			Feed:  "Slashdot",
			Title: "Ask Slashdot:  What Was Your Longest-Lived PC?",
			URL:   "https://ask.slashdot.org/story/23/04/02/0058226/ask-slashdot-what-was-your-longest-lived-pc?utm_source=rss1.0mainlinkanon&utm_medium=feed",
		},
		{
			Feed:  "Slashdot",
			Title: "San Francisco Faces 'Doom Loop' from Office Workers Staying Home, Gutting Tax Base",
			URL:   "https://it.slashdot.org/story/23/04/01/2059224/san-francisco-faces-doom-loop-from-office-workers-staying-home-gutting-tax-base?utm_source=rss1.0mainlinkanon&utm_medium=feed",
		},
	}
	rdf, err := os.Open("testdata/rdf.xml")
	if err != nil {
		t.Fatalf("Cannot read file content: %+v", err)
	}
	got, err := morningpost.ParseRDFResponse(rdf)
	if err != nil {
		t.Fatalf("Cannot parse rdf content: %+v", err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestParseRDFResponse_ErrorsIfDataIsNotXML(t *testing.T) {
	t.Parallel()
	_, err := morningpost.ParseRSSResponse([]byte("{}"))
	if err == nil {
		t.Fatalf("want error but not found")
	}
}

func TestHandleFeeds_AnswersMethodNotAllowedGivenRequestWithBogusMethod(t *testing.T) {
	t.Parallel()
	want := http.StatusMethodNotAllowed
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	ts := httptest.NewServer(http.HandlerFunc(m.HandleFeeds))
	defer ts.Close()
	req, err := http.NewRequest("bogus", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	got := resp.StatusCode
	if want != got {
		t.Fatalf("want status code %d, got %d", want, got)
	}
}

func TestHandleFeeds_AnswersStatusOKGivenRequestWithMethodHead(t *testing.T) {
	t.Parallel()
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	ts := httptest.NewServer(http.HandlerFunc(m.HandleFeeds))
	defer ts.Close()
	resp, err := http.Head(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code %d", resp.StatusCode)
	}
}

func TestHandleFeeds_ReturnsExpectedStatusCodeGivenRequestWithMethodPostAndBody(t *testing.T) {
	t.Parallel()
	want := http.StatusOK
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	tsHandler := httptest.NewServer(http.HandlerFunc(m.HandleFeeds))
	defer tsHandler.Close()
	ts := newServerWithContentTypeAndBodyResponse(t, "application/rss+xml", "testdata/rss.xml")
	reqBody := url.Values{
		"url": {ts.URL},
	}
	resp, err := http.PostForm(tsHandler.URL, reqBody)
	if err != nil {
		t.Fatal(err)
	}
	got := resp.StatusCode
	if want != got {
		t.Fatalf("want status code %d, got %d", want, got)
	}
}

func TestHandleFeeds_AnswersBadRequestGivenRequestWithMethodPostAndNoBody(t *testing.T) {
	t.Parallel()
	want := http.StatusBadRequest
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	ts := httptest.NewServer(http.HandlerFunc(m.HandleFeeds))
	defer ts.Close()
	resp, err := http.PostForm(ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := resp.StatusCode
	if want != got {
		t.Fatalf("want status code %d, got %d", want, got)
	}
}

func TestHandleFeeds_AnswersBadRequestGivenRequestWithMethodPostAndBlankSpacesInBodyURL(t *testing.T) {
	t.Parallel()
	want := http.StatusBadRequest
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	ts := httptest.NewServer(http.HandlerFunc(m.HandleFeeds))
	defer ts.Close()
	reqBody := url.Values{
		"url": {" "},
	}
	resp, err := http.PostForm(ts.URL, reqBody)
	if err != nil {
		t.Fatal(err)
	}
	got := resp.StatusCode
	if want != got {
		t.Fatalf("want status code %d, got %d", want, got)
	}
}

func TestHandleNews_AnswersMethodNotAllowedGivenRequestWithBogusMethod(t *testing.T) {
	t.Parallel()
	want := http.StatusMethodNotAllowed
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	ts := httptest.NewServer(http.HandlerFunc(m.HandleHome))
	defer ts.Close()
	req, err := http.NewRequest("bogus", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	got := resp.StatusCode
	if want != got {
		t.Fatalf("want status code %d, got %d", want, got)
	}
}

func TestFeedGetNews_ReturnsNewsFromFeed(t *testing.T) {
	testCases := []struct {
		contentType string
		desc        string
	}{
		{
			contentType: "application/rss+xml",
			desc:        "Given Response Content-Type application/rss+xml",
		},
		{
			contentType: "application/xml",
			desc:        "Given Response Content-Type application/xml",
		},
		{
			contentType: "text/xml; charset=UTF-8",
			desc:        "Given Response Content-Type text/xml",
		},
	}
	want := []morningpost.News{
		{
			Feed:  "FeedForAll Sample Feed",
			Title: "RSS Solutions for Restaurants",
			URL:   "http://www.feedforall.com/restaurant.htm",
		},
		{
			Feed:  "FeedForAll Sample Feed",
			Title: "RSS Solutions for Schools and Colleges",
			URL:   "http://www.feedforall.com/schools.htm",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ts := newServerWithContentTypeAndBodyResponse(t, tC.contentType, "testdata/rss.xml")
			feed := morningpost.Feed{
				Endpoint: ts.URL,
				Type:     morningpost.FeedTypeRSS,
			}
			got, err := feed.GetNews()
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(want, got) {
				t.Fatal(cmp.Diff(want, got))
			}
		})
	}
}

func TestFeedGetNews_ReturnsNewsFromFeedGivenResponseContentTypeApplicationAtomXML(t *testing.T) {
	want := []morningpost.News{
		{
			Feed:  "Chris's Wiki :: blog",
			Title: "ZFS on Linux and when you get stale NFSv3 mounts",
			URL:   "https://utcc.utoronto.ca/~cks/space/blog/linux/ZFSAndNFSMountInvalidation",
		},
		{
			Feed:  "Chris's Wiki :: blog",
			Title: "Debconf's questions, or really whiptail, doesn't always work in xterms",
			URL:   "https://utcc.utoronto.ca/~cks/space/blog/linux/DebconfWhiptailVsXterm",
		},
	}
	ts := newServerWithContentTypeAndBodyResponse(t, "application/atom+xml", "testdata/atom.xml")
	feed := morningpost.Feed{
		Endpoint: ts.URL,
		Type:     morningpost.FeedTypeAtom,
	}
	got, err := feed.GetNews()
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFeedGetNews_SetsHeadersOnHTTPRequest(t *testing.T) {
	t.Parallel()
	wantHeaders := map[string]string{
		"user-agent": "MorningPost/0.1",
		"accept":     "*/*",
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for header, want := range wantHeaders {
			got := r.Header.Get(header)
			if want != got {
				t.Errorf("want value %q, got %q for header %q", want, got, header)
			}
		}
		w.Header().Set("content-type", "application/rss+xml")
		w.Write([]byte(`<rss></rss>`))
	}))
	defer ts.Close()
	feed := morningpost.Feed{
		Endpoint: ts.URL,
		Type:     morningpost.FeedTypeRSS,
	}
	_, err := feed.GetNews()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFeedGetNews_ErrorsGivenInvalidEndpoint(t *testing.T) {
	t.Parallel()
	feed := morningpost.Feed{
		Endpoint: "bogus://",
	}
	_, err := feed.GetNews()
	if err == nil {
		t.Fatal("want error but got nil")
	}
}

func TestFeedGetNews_ErrorsIfResponseContentTypeIsUnexpected(t *testing.T) {
	t.Parallel()
	ts := newServerWithContentTypeResponse(t, "bogus")
	feed := morningpost.Feed{
		Endpoint: ts.URL,
	}
	_, err := feed.GetNews()
	if err == nil {
		t.Fatal("want error but got nil")
	}
}

func TestGetNews_ReturnsNewsFromAllFeedsGivenPopulatedStore(t *testing.T) {
	t.Parallel()
	want := 2
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	ts1 := newServerWithContentTypeAndBodyResponse(t, "application/rss+xml", "testdata/rss-with-one-news-1.xml")
	feed1 := morningpost.Feed{
		Endpoint: ts1.URL,
		Type:     morningpost.FeedTypeRSS,
	}
	m.Store.Data[0] = feed1
	ts2 := newServerWithContentTypeAndBodyResponse(t, "application/rss+xml", "testdata/rss-with-one-news-2.xml")
	feed2 := morningpost.Feed{
		Endpoint: ts2.URL,
		Type:     morningpost.FeedTypeRSS,
	}
	m.Store.Data[1] = feed2
	err := m.GetNews()
	if err != nil {
		t.Fatal(err)
	}
	got := len(m.News)
	if want != got {
		t.Fatalf("want %d news, got %d", want, got)
	}
}

func TestGetNews_CleansUpNewsOnEachExecution(t *testing.T) {
	t.Parallel()
	want := 0
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	m.News = []morningpost.News{{
		Title: "I hope it disappear",
		URL:   "http://fake.url/fake-news",
	}}
	err := m.GetNews()
	if err != nil {
		t.Fatal(err)
	}
	got := len(m.News)
	if want != got {
		t.Fatalf("want number of news %d, got %d", want, got)
	}
}

func TestGetNews_ErrorsIfFeedGetNewsErrors(t *testing.T) {
	t.Parallel()
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	feed := morningpost.Feed{
		Endpoint: "bogus://",
	}
	m.Store.Data[0] = feed
	err := m.GetNews()
	if err == nil {
		t.Fatal("want error but got nil")
	}
}

func TestRandomNews_RandomizePageNewsToSameNumberAsShowMaxNews(t *testing.T) {
	t.Parallel()
	want := 1
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	m.ShowMaxNews = 1
	m.News = []morningpost.News{
		{Title: "RSS Solutions for Restaurants", URL: "http://www.feedforall.com/restaurant.htm"},
		{Title: "RSS Solutions for Schools and Colleges", URL: "http://www.feedforall.com/schools.htm"},
	}
	m.RandomNews()
	got := len(m.PageNews)
	if want != got {
		t.Fatalf("want PageNews %d, got %d", want, got)
	}
}

func TestParseAtomResponse_ReturnsNewsGivenAtomWithTwoNews(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{
		{
			Feed:  "Chris's Wiki :: blog",
			Title: "ZFS on Linux and when you get stale NFSv3 mounts",
			URL:   "https://utcc.utoronto.ca/~cks/space/blog/linux/ZFSAndNFSMountInvalidation",
		},
		{
			Feed:  "Chris's Wiki :: blog",
			Title: "Debconf's questions, or really whiptail, doesn't always work in xterms",
			URL:   "https://utcc.utoronto.ca/~cks/space/blog/linux/DebconfWhiptailVsXterm",
		},
	}
	inputData, err := os.ReadFile("testdata/atom.xml")
	if err != nil {
		t.Fatalf("Cannot read file content: %+v", err)
	}
	got, err := morningpost.ParseAtomResponse(inputData)
	if err != nil {
		t.Fatalf("Cannot parse content %q: %+v", inputData, err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestParseAtomResponse_SkipItemGivenNewsWithNoURL(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{}
	inputData, err := os.ReadFile("testdata/no-url-item-atom.xml")
	if err != nil {
		t.Fatalf("Cannot read file content: %+v", err)
	}
	got, err := morningpost.ParseAtomResponse(inputData)
	if err != nil {
		t.Fatalf("Cannot parse content %q: %+v", inputData, err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestParseAtomResponse_SkipItemGivenInvalidURL(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{}
	inputData, err := os.ReadFile("testdata/invalid-url-item-atom.xml")
	if err != nil {
		t.Fatalf("Cannot read file content: %+v", err)
	}
	got, err := morningpost.ParseAtomResponse(inputData)
	if err != nil {
		t.Fatalf("Cannot parse content %q: %+v", inputData, err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestParseAtomResponse_SkipItemGivenNoTitle(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{}
	inputData, err := os.ReadFile("testdata/no-title-item-atom.xml")
	if err != nil {
		t.Fatalf("Cannot read file content: %+v", err)
	}
	got, err := morningpost.ParseAtomResponse(inputData)
	if err != nil {
		t.Fatalf("Cannot parse content %q: %+v", inputData, err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestParseAtomResponse_ErrorsIfDataIsNotXML(t *testing.T) {
	t.Parallel()
	_, err := morningpost.ParseAtomResponse([]byte("{}"))
	if err == nil {
		t.Fatalf("want error but not found")
	}
}

func TestHandleHome_AnswersNotFoundForUnkownRoute(t *testing.T) {
	t.Parallel()
	want := http.StatusNotFound
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	ts := httptest.NewServer(http.HandlerFunc(m.HandleHome))
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/bogus")
	if err != nil {
		t.Fatal(err)
	}
	got := resp.StatusCode
	if want != got {
		t.Fatalf("want response status code %d, got %d", want, got)
	}
}

func TestAddNews_AddsCorrectNewsGivenConcurrentAccess(t *testing.T) {
	t.Parallel()
	want := 1000
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	allNews := generateOneThousandsNews(t)
	var wg sync.WaitGroup
	for x := 0; x < 1000; x++ {
		wg.Add(1)
		go func(x int) {
			m.AddNews([]morningpost.News{allNews[x]})
			wg.Done()
		}(x)
	}
	wg.Wait()
	got := len(m.News)
	if want != got {
		t.Fatalf("want %d news, got %d", want, got)
	}
}

func TestParseLinkTags_ReturnsFeedEndpointGivenHTMLPageWithRSSFeedInBodyElement(t *testing.T) {
	t.Parallel()
	want := []morningpost.Feed{{
		Endpoint: "https://bitfieldconsulting.com/golang?format=rss",
		Type:     morningpost.FeedTypeRSS,
	}}
	got, err := morningpost.ParseLinkTags([]byte(`<a href="https://bitfieldconsulting.com/golang?format=rss" title="Go RSS" class="social-rss">Go RSS</a>`), "")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestParseLinkTags_ReturnsFeedEndpointGivenHTMLPageWithRSSFeedInLinkElement(t *testing.T) {
	t.Parallel()
	want := []morningpost.Feed{{
		Endpoint: "http://fake.url/rss",
		Type:     morningpost.FeedTypeRSS,
	}}
	got, err := morningpost.ParseLinkTags([]byte(`<link type="application/rss+xml" title="Unit Test" href="http://fake.url/rss" />`), "")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestParseLinkTags_ReturnsFeedEndpointGivenHTMLPageWithAtomFeedInLinkElement(t *testing.T) {
	t.Parallel()
	want := []morningpost.Feed{{
		Endpoint: "http://fake.url/feed/",
		Type:     morningpost.FeedTypeAtom,
	}}
	got, err := morningpost.ParseLinkTags([]byte(`<link type="application/atom+xml" title="Unit Test" href="http://fake.url/feed/" />`), "")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFindFeeds_ErrorsIfStatusCodeIsNotStatusOK(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer ts.Close()
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	_, err := m.FindFeeds(ts.URL)
	if err == nil {
		t.Fatal("want error but got nil")
	}
}

func TestFindFeeds_ErrorsGivenURLWithSchemeBogus(t *testing.T) {
	t.Parallel()
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	_, err := m.FindFeeds("bogus://")
	if err == nil {
		t.Fatal("want error but got nil")
	}
}

func TestFindFeeds_ErrorsGivenUnexpectedContentType(t *testing.T) {
	t.Parallel()
	ts := newServerWithContentTypeResponse(t, "bogus")
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	_, err := m.FindFeeds(ts.URL)
	if err == nil {
		t.Fatal("want error but got nil")
	}
}

func TestFindFeeds_ReturnsExpectedFeedsGivenApplicationRSSXMLContentType(t *testing.T) {
	t.Parallel()
	ts := newServerWithContentTypeAndBodyResponse(t, "application/rss+xml", "testdata/rss.xml")
	want := []morningpost.Feed{{
		Endpoint: ts.URL,
		Type:     morningpost.FeedTypeRSS,
	}}
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	got, err := m.FindFeeds(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFindFeeds_ReturnsExpectedFeedsGivenApplicationXMLContentTypeAndRSSData(t *testing.T) {
	t.Parallel()
	ts := newServerWithContentTypeAndBodyResponse(t, "application/xml", "testdata/rss.xml")
	want := []morningpost.Feed{{
		Endpoint: ts.URL,
		Type:     morningpost.FeedTypeRSS,
	}}
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	got, err := m.FindFeeds(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFindFeeds_ReturnsExpectedFeedsGivenTextXMLContentTypeAndRSSData(t *testing.T) {
	t.Parallel()
	ts := newServerWithContentTypeAndBodyResponse(t, "text/xml", "testdata/rss.xml")
	want := []morningpost.Feed{{
		Endpoint: ts.URL,
		Type:     morningpost.FeedTypeRSS,
	}}
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	got, err := m.FindFeeds(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFindFeeds_ReturnsExpectedFeedsGivenTextXMLContentTypeAndRDFData(t *testing.T) {
	t.Parallel()
	ts := newServerWithContentTypeAndBodyResponse(t, "text/xml", "testdata/rdf.xml")
	want := []morningpost.Feed{{
		Endpoint: ts.URL,
		Type:     morningpost.FeedTypeRDF,
	}}
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	got, err := m.FindFeeds(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFindFeeds_ReturnsExpectedFeedsGivenAtomApplicationContentType(t *testing.T) {
	t.Parallel()
	ts := newServerWithContentTypeAndBodyResponse(t, "application/atom+xml", "testdata/atom.xml")
	want := []morningpost.Feed{{
		Endpoint: ts.URL,
		Type:     "Atom",
	}}
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	got, err := m.FindFeeds(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFindFeeds_ReturnsExpectedFeedsGivenHTMLPageWithFeedsInFullLinkFormat(t *testing.T) {
	t.Parallel()
	want := []morningpost.Feed{
		{
			Endpoint: "http://example.com/rss",
			Type:     morningpost.FeedTypeRSS,
		},
		{
			Endpoint: "http://example.com/atom",
			Type:     morningpost.FeedTypeAtom,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/":
			w.Header().Set("content-type", "text/html")
			w.Write([]byte(`<link type="application/rss+xml" title="RSS Unit Test" href="http://example.com/rss" />
	                        <link type="application/atom+xml" title="Atom Unit Test" href="http://example.com/atom" />`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	got, err := m.FindFeeds(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFindFeeds_ReturnsExpectedFeedsGivenHTMLPageWithFeedInRelativeLinkFormat(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/":
			w.Header().Set("content-type", "text/html")
			w.Write([]byte(`<link type="application/rss+xml" title="RSS Unit Test" href="rss" />`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	want := []morningpost.Feed{
		{
			Endpoint: ts.URL + "/rss",
			Type:     morningpost.FeedTypeRSS,
		},
	}
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	got, err := m.FindFeeds(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestFeedFinds_SetsHeadersOnHTTPRequest(t *testing.T) {
	t.Parallel()
	wantHeaders := map[string]string{
		"user-agent": "MorningPost/0.1",
		"accept":     "*/*",
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for header, want := range wantHeaders {
			got := r.Header.Get(header)
			if want != got {
				t.Errorf("want value %q, got %q for header %q", want, got, header)
			}
		}
		w.Header().Set("content-type", "application/rss+xml")
		w.Write([]byte(`<rss></rss>`))
	}))
	defer ts.Close()
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	_, err := m.FindFeeds(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
}

package morningpost_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/thiagonache/morningpost"
)

func newServerWithContentTypeAndBodyResponse(t *testing.T, contentType string, filePath string) *httptest.Server {
	t.Helper()
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", contentType)
		w.Write(data)
	}))
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
	ts := newServerWithContentTypeAndBodyResponse(t, "application/rss+xml", "testdata/rss-empty-tag.xml")
	defer ts.Close()
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

func TestHandleFeeds_ReturnsExpectedStatusCodeGivenDeleteRequest(t *testing.T) {
	t.Parallel()
	want := http.StatusOK
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	tsHandler := httptest.NewServer(http.HandlerFunc(m.HandleFeeds))
	defer tsHandler.Close()
	ts := newServerWithContentTypeAndBodyResponse(t, "application/rss+xml", "testdata/rss-empty-tag.xml")
	defer ts.Close()
	req, err := http.NewRequest(http.MethodDelete, tsHandler.URL, nil)
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

func TestHandleFeeds_RendersProperHTMLPageGivenEmptyStore(t *testing.T) {
	t.Parallel()
	want, err := os.ReadFile("testdata/golden/feeds.html")
	if err != nil {
		t.Fatal(err)
	}
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	ts := httptest.NewServer(http.HandlerFunc(m.HandleFeeds))
	defer ts.Close()
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code %d", resp.StatusCode)
	}
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestHandleHome_RendersProperHTMLPageGivenEmptyStore(t *testing.T) {
	t.Parallel()
	want, err := os.ReadFile("testdata/golden/home.html")
	if err != nil {
		t.Fatal(err)
	}
	m := newMorningPostWithBogusFileStoreAndNoOutput(t)
	ts := httptest.NewServer(http.HandlerFunc(m.HandleHome))
	defer ts.Close()
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code %d", resp.StatusCode)
	}
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
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
		{Title: "RSS Solutions for Restaurants", URL: "http://www.feedforall.com/restaurant.htm"},
		{Title: "RSS Solutions for Schools and Colleges", URL: "http://www.feedforall.com/schools.htm"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ts := newServerWithContentTypeAndBodyResponse(t, tC.contentType, "testdata/rss.xml")
			defer ts.Close()
			feed := morningpost.Feed{
				Endpoint: ts.URL,
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
			Title: "ZFS on Linux and when you get stale NFSv3 mounts",
			URL:   "https://utcc.utoronto.ca/~cks/space/blog/linux/ZFSAndNFSMountInvalidation",
		},
		{
			Title: "Debconf's questions, or really whiptail, doesn't always work in xterms",
			URL:   "https://utcc.utoronto.ca/~cks/space/blog/linux/DebconfWhiptailVsXterm",
		},
	}
	ts := newServerWithContentTypeAndBodyResponse(t, "application/atom+xml", "testdata/atom.xml")
	defer ts.Close()
	feed := morningpost.Feed{
		Endpoint: ts.URL,
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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "bogus")
	}))
	defer ts.Close()
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
	defer ts1.Close()
	feed1 := morningpost.Feed{
		Endpoint: ts1.URL,
	}
	m.Store.Data[feed1.Endpoint] = feed1
	ts2 := newServerWithContentTypeAndBodyResponse(t, "application/rss+xml", "testdata/rss-with-one-news-2.xml")
	defer ts2.Close()
	feed2 := morningpost.Feed{
		Endpoint: ts2.URL,
	}
	m.Store.Data[feed2.Endpoint] = feed2
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
	m.Store.Data[feed.Endpoint] = feed
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

func TestParseAtomResponse_ReturnsExpectedNewsGivenRSSWithTwoNews(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{
		{
			Title: "ZFS on Linux and when you get stale NFSv3 mounts",
			URL:   "https://utcc.utoronto.ca/~cks/space/blog/linux/ZFSAndNFSMountInvalidation",
		},
		{
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

func TestNewFeed_ReturnsExpectedFeedGivenFullURL(t *testing.T) {
	t.Parallel()
	want := morningpost.Feed{
		Endpoint: "http://fake.url",
	}
	got, err := morningpost.NewFeed("http://fake.url")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestNewFeed_AddsHTTPSSchemeGivenURLWithoutScheme(t *testing.T) {
	t.Parallel()
	want := "https"
	feed, err := morningpost.NewFeed("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(feed.Endpoint)
	if err != nil {
		t.Fatal(err)
	}
	got := u.Scheme
	if want != got {
		t.Fatalf("want scheme %q, got %q", want, got)
	}
}

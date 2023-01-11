package morningpost_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/thiagonache/morningpost"
)

func TestNewsString_FillsTitleWithSpacesUpToTitleMaxSizeGivenNewsWithSmallTitle(t *testing.T) {
	t.Parallel()
	want := "ChatGPT is going to take my job, what should I do?                               https://iamnonsense.com/chatgpt-going-to-take-jobs"
	news := morningpost.News{
		Title: "ChatGPT is going to take my job, what should I do?",
		URL:   "https://iamnonsense.com/chatgpt-going-to-take-jobs",
	}
	got := news.String()
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestNewsString_TrucantesTitleToTitleMaxSizeGivenNewsWithBigTitle(t *testing.T) {
	t.Parallel()
	want := "ChatGPT is going to take my job, what should I do? ChatGPT is going to take my j https://iamnonsense.com/chatgpt-going-to-take-jobs"
	news := morningpost.News{
		Title: "ChatGPT is going to take my job, what should I do? ChatGPT is going to take my job, what should I do?",
		URL:   "https://iamnonsense.com/chatgpt-going-to-take-jobs",
	}
	got := news.String()
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestParseRSS_ReturnsExpectedNewsGivenRSSWithTwoNews(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{
		{Title: "RSS Solutions for Restaurants", URL: "http://www.feedforall.com/restaurant.htm"},
		{Title: "RSS Solutions for Schools and Colleges", URL: "http://www.feedforall.com/schools.htm"},
	}
	f, err := os.Open("testdata/rss.xml")
	if err != nil {
		t.Fatalf("Cannot open file testdata/rss.xml: %+v", err)
	}
	defer f.Close()
	input, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("Cannot read file content: %+v", err)
	}
	got, err := morningpost.ParseRSSResponse(input)
	if err != nil {
		t.Fatalf("Cannot parse content %q: %+v", input, err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGetRSSFeedReturnsExpectedNewsGivenRSSWithTwoNews(t *testing.T) {
	t.Parallel()
	want := []morningpost.News{
		{Title: "RSS Solutions for Restaurants", URL: "http://www.feedforall.com/restaurant.htm"},
		{Title: "RSS Solutions for Schools and Colleges", URL: "http://www.feedforall.com/schools.htm"},
	}
	f, err := os.Open("testdata/rss.xml")
	if err != nil {
		t.Fatalf("Cannot open file testdata/rss.xml: %+v", err)
	}
	defer f.Close()
	input, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("Cannot read file content: %+v", err)
	}
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.RequestURI != "/rss" {
			t.Fatal("want URI to be /rss")
		}
		fmt.Fprintln(w, string(input))
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

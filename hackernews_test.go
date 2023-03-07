package morningpost_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thiagonache/morningpost"
)

func TestNewHackerNewsClient_SetCorrectHTTPHostByDefault(t *testing.T) {
	t.Parallel()
	want := "https://news.ycombinator.com"
	client := morningpost.NewHackerNewsClient()
	got := client.HTTPHost
	if want != got {
		t.Fatalf("\n(want) %q\n(got)  %q", want, got)
	}
}

func TestNewHackerNewsClient_SetCorrectURIByDefault(t *testing.T) {
	t.Parallel()
	want := "rss"
	client := morningpost.NewHackerNewsClient()
	got := client.URI
	if want != got {
		t.Fatalf("Wrong URI\n(want) %q\n(got)  %q", want, got)
	}
}

func TestNewHackerNewsClient_SetCorrectHTTPTimeoutByDefault(t *testing.T) {
	t.Parallel()
	want := 5 * time.Second
	client := morningpost.NewHackerNewsClient()
	got := client.HTTPClient.Timeout
	if want != got {
		t.Fatalf("Wrong timeout\n(want) %q\n(got)  %q", want, got)
	}
}

func TestHackerNewsGetNews_RequestsCorrectURIByDefault(t *testing.T) {
	t.Parallel()
	want := "/rss"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.RequestURI
		if want != got {
			t.Fatalf("Unexpected URI\n(want) %q\n(got)  %q", want, got)
		}
		w.Write(emptyRSSData)
	}))
	defer ts.Close()
	client := morningpost.NewHackerNewsClient()
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	_, err := client.GetNews()
	if err != nil {
		t.Fatal(err)
	}
}

func TestHackerNewsGetNews_ErrorsIfResponseCodeIsNotHTTPStatusOK(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write(emptyRSSData)
	}))
	defer ts.Close()
	client := morningpost.NewHackerNewsClient()
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	_, err := client.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestHackerNewsGetNews_ErrorsIfHTTPRequestErrors(t *testing.T) {
	t.Parallel()
	client := morningpost.NewHackerNewsClient()
	client.HTTPHost = "bogus"
	_, err := client.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}
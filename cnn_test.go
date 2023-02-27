package morningpost_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thiagonache/morningpost"
)

func TestNewCNNClient_SetCorrectHTTPHostByDefault(t *testing.T) {
	t.Parallel()
	want := "http://rss.cnn.com"
	client := morningpost.NewCNNClient()
	got := client.HTTPHost
	if want != got {
		t.Fatalf("Wrong HTTP host\n(want) %q\n(got)  %q", want, got)
	}
}

func TestNewCNNClient_SetCorrectURIByDefault(t *testing.T) {
	t.Parallel()
	want := "rss/cnn_topstories.rss"
	client := morningpost.NewCNNClient()
	got := client.URI
	if want != got {
		t.Fatalf("Wrong URI\n(want) %q\n(got)  %q", want, got)
	}
}

func TestCNNGetNews_RequestsCorrectURIByDefault(t *testing.T) {
	t.Parallel()
	want := "/rss/cnn_topstories.rss"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.RequestURI
		if want != got {
			t.Fatalf("Unexpected URI\n(want) %q\n(got)  %q", want, got)
		}
		w.Write(emptyRSSData)
	}))
	defer ts.Close()
	client := morningpost.NewCNNClient()
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	_, err := client.GetNews()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCNNGetNews_ErrorsIfResponseCodeIsNotHTTPStatusOK(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write(emptyRSSData)
	}))
	defer ts.Close()
	client := morningpost.NewCNNClient()
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	_, err := client.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestCNNGetNews_ErrorsIfHTTPRequestErrors(t *testing.T) {
	t.Parallel()
	client := morningpost.NewCNNClient()
	client.HTTPHost = "bogus"
	_, err := client.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

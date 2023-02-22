package morningpost_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thiagonache/morningpost"
)

func TestNewBITClient_SetCorrectHTTPHostByDefault(t *testing.T) {
	t.Parallel()
	want := "https://bitfieldconsulting.com"
	client := morningpost.NewBITClient()
	got := client.HTTPHost
	if want != got {
		t.Fatalf("\n(want) %q\n(got)  %q", want, got)
	}
}

func TestBITGetNews_RequestsCorrectURIByDefault(t *testing.T) {
	t.Parallel()
	want := "/golang?format=rss"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.RequestURI
		if want != got {
			t.Fatalf("Unexpected URI\n(want) %q\n(got)  %q", want, got)
		}
		w.Write(emptyRSSData)
	}))
	defer ts.Close()
	client := morningpost.NewBITClient()
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	_, err := client.GetNews()
	if err != nil {
		t.Fatal(err)
	}
}

func TestBITGetNews_ErrorsIfResponseCodeIsNotHTTPStatusOK(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write(emptyRSSData)
	}))
	defer ts.Close()
	client := morningpost.NewBITClient()
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	_, err := client.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

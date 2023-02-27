package morningpost_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thiagonache/morningpost"
)

func TestNewTechCrunchClient_SetCorrectHTTPHostByDefault(t *testing.T) {
	t.Parallel()
	want := "https://techcrunch.com"
	client := morningpost.NewTechCrunchClient()
	got := client.HTTPHost
	if want != got {
		t.Fatalf("\n(want) %q\n(got)  %q", want, got)
	}
}

func TestNewTechCrunchClient_SetCorrectURIByDefault(t *testing.T) {
	t.Parallel()
	want := "feed/"
	client := morningpost.NewTechCrunchClient()
	got := client.URI
	if want != got {
		t.Fatalf("Wrong URI\n(want) %q\n(got)  %q", want, got)
	}
}

func TestTechCrunchGetNews_RequestsCorrectURIByDefault(t *testing.T) {
	t.Parallel()
	want := "/feed/"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.RequestURI
		if want != got {
			t.Fatalf("Unexpected URI\n(want) %q\n(got)  %q", want, got)
		}
		w.Write(emptyRSSData)
	}))
	defer ts.Close()
	client := morningpost.NewTechCrunchClient()
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	_, err := client.GetNews()
	if err != nil {
		t.Fatal(err)
	}
}

func TestTechCrunchGetNews_ErrorsIfResponseCodeIsNotHTTPStatusOK(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write(emptyRSSData)
	}))
	defer ts.Close()
	client := morningpost.NewTechCrunchClient()
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	_, err := client.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestTechCrunchGetNews_ErrorsIfHTTPRequestErrors(t *testing.T) {
	t.Parallel()
	client := morningpost.NewTechCrunchClient()
	client.HTTPHost = "bogus"
	_, err := client.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

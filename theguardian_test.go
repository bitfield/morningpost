package morningpost_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/thiagonache/morningpost"
)

func TestNewTheGuardianClient_SetCorrectHTTPHostByDefault(t *testing.T) {
	t.Parallel()
	want := "https://content.guardianapis.com"
	client := morningpost.NewTheGuardianClient()
	got := client.HTTPHost
	if want != got {
		t.Fatalf("\n(want) %q\n(got)  %q", want, got)
	}
}

func TestTheGuardianGetNews_RequestsCorrectURIByDefault(t *testing.T) {
	t.Parallel()
	respContent, err := os.ReadFile("testdata/theguardian.json")
	if err != nil {
		t.Fatal(err)
	}
	want := "/search?api-key=fake"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.RequestURI
		if want != got {
			t.Fatalf("Unexpected URI\n(want) %q\n(got)  %q", want, got)
		}
		w.Write(respContent)
	}))
	defer ts.Close()
	client := morningpost.NewTheGuardianClient()
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	os.Setenv("TheGuardianAPIKey", "fake")
	_, err = client.GetNews()
	if err != nil {
		t.Fatal(err)
	}
}

func TestTheGuardianGetNews_ErrorsIfResponseCodeIsNotHTTPStatusOK(t *testing.T) {
	t.Parallel()
	respContent, err := os.ReadFile("testdata/theguardian.json")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write(respContent)
	}))
	defer ts.Close()
	client := morningpost.NewTheGuardianClient()
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	os.Setenv("TheGuardianAPIKey", "fake")
	_, err = client.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestTheGuardianGetNews_ErrorsIfEnvVarIsNotSet(t *testing.T) {
	//t.Parallel()
	// this test cannot run in paralell since the API key is set in other tests
	respContent, err := os.ReadFile("testdata/theguardian.json")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(respContent)
	}))
	defer ts.Close()
	client := morningpost.NewTheGuardianClient()
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	_, err = client.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestParseGuardianResponse_ReturnsExpectedNewsGivenJSONWithOneNews(t *testing.T) {
	t.Parallel()
	input, err := os.ReadFile("testdata/theguardian.json")
	if err != nil {
		t.Fatal(err)
	}
	got, err := morningpost.ParseTheGuardianResponse(input)
	if err != nil {
		t.Fatal(err)
	}
	want := []morningpost.News{
		{
			Title: "Australian Open 2023 day one: Norrie and Tsitsipas through, Swiatek in action – live",
			URL:   "https://www.theguardian.com/sport/live/2023/jan/16/australian-open-2023-day-one-swiatek-tsitsipas-and-medvedev-in-action-live",
		},
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

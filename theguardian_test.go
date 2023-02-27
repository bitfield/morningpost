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
	os.Setenv("TheGuardianAPIKey", "fake")
	client, err := morningpost.NewTheGuardianClient()
	if err != nil {
		t.Fatal(err)
	}
	got := client.HTTPHost
	if want != got {
		t.Fatalf("\n(want) %q\n(got)  %q", want, got)
	}
}

func TestNewTheGuardianClient_SetCorrectURIByDefault(t *testing.T) {
	t.Parallel()
	want := "search?api-key=fake"
	os.Setenv("TheGuardianAPIKey", "fake")
	client, err := morningpost.NewTheGuardianClient()
	if err != nil {
		t.Fatal(err)
	}
	got := client.URI
	if want != got {
		t.Fatalf("Wrong URI\n(want) %q\n(got)  %q", want, got)
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
	os.Setenv("TheGuardianAPIKey", "fake")
	client, err := morningpost.NewTheGuardianClient()
	if err != nil {
		t.Fatal(err)
	}
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
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
	os.Setenv("TheGuardianAPIKey", "fake")
	client, err := morningpost.NewTheGuardianClient()
	if err != nil {
		t.Fatal(err)
	}
	client.HTTPHost = ts.URL
	client.HTTPClient = ts.Client()
	_, err = client.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestNewTheGuardianClient_ErrorsIfEnvVarIsNotSet(t *testing.T) {
	//t.Parallel()
	// this test cannot run in paralell since the API key is set in other tests
	_, err := morningpost.NewTheGuardianClient()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestParseTheGuardianResponse_ReturnsExpectedNewsGivenJSONWithOneNews(t *testing.T) {
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

func TestTheGuardianGetNews_ErrorsIfHTTPRequestErrors(t *testing.T) {
	t.Parallel()
	os.Setenv("TheGuardianAPIKey", "fake")
	client, err := morningpost.NewTheGuardianClient()
	if err != nil {
		t.Fatal(err)
	}
	client.HTTPHost = "bogus"
	_, err = client.GetNews()
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestParseTheGuardianResponse_ErrorsIfDataIsNotJSON(t *testing.T) {
	t.Parallel()
	_, err := morningpost.ParseTheGuardianResponse(emptyRSSData)
	if err == nil {
		t.Fatal("want error but not found")
	}
}

func TestParseTheGuardianResponse_ErrorsIfResponseStatusIsNotGuardianStatusOK(t *testing.T) {
	t.Parallel()
	_, err := morningpost.ParseTheGuardianResponse(emptyJSONData)
	if err == nil {
		t.Fatal("want error but not found")
	}
}

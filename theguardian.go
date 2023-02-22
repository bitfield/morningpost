package morningpost

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const GuardianStatusOK = "ok"

type TheGuardianClient struct {
	HTTPClient *http.Client
	HTTPHost   string
}

func (tg TheGuardianClient) GetNews() ([]News, error) {
	apiKey := os.Getenv("TheGuardianAPIKey")
	if apiKey == "" {
		return nil, fmt.Errorf("OS environment variable TheGuardianAPIKey not found")
	}
	resp, err := tg.HTTPClient.Get(fmt.Sprintf("%s/search?api-key=%s", tg.HTTPHost, apiKey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status %q", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	news, err := ParseTGResponse(data)
	if err != nil {
		return nil, err
	}
	return news, nil
}

func NewTGClient() *TheGuardianClient {
	return &TheGuardianClient{
		HTTPClient: http.DefaultClient,
		HTTPHost:   "https://content.guardianapis.com",
	}
}

func ParseTGResponse(input []byte) ([]News, error) {
	type guardianResponse struct {
		Response struct {
			Status  string
			Results []struct {
				ID       string
				WebTitle string
				WebURL   string
			}
		}
	}
	resp := guardianResponse{}
	err := json.Unmarshal(input, &resp)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal data %q: %w", input, err)
	}
	if resp.Response.Status != GuardianStatusOK {
		return nil, fmt.Errorf("unexpected response status %q: %+v", resp.Response.Status, resp.Response)
	}
	news := make([]News, len(resp.Response.Results))
	for i, r := range resp.Response.Results {
		news[i] = News{
			Title: r.WebTitle,
			URL:   r.WebURL,
		}
	}
	return news, err
}

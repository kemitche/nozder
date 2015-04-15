package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// TwitchConfig contains configuration for an HTTP client connecting to twitch.tv
type TwitchConfig struct {
	version    int
	clientID   string
	httpClient http.Client
}

func (t *TwitchConfig) headers() *map[string]string {
	var hdrs map[string]string

	hdrs["Accept"] = fmt.Sprintf("application/vnd.twitchtv.v%d+json", t.version)
	hdrs["Client-ID"] = t.clientID

	return &hdrs
}

func (t *TwitchConfig) searchForGameStreams(gameQuery string, isLive bool) ([]string, *error) {
	var queryParams url.Values
	queryParams.Add("query", gameQuery)
	reqURL := "https://twitch.tv/kraken/search/streams?" + queryParams.Encode()
	req, err := t.constructRequest("GET", reqURL)
	if err != nil {
		return nil, err
	}
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return []string{body}, nil
}

func (t *TwitchConfig) constructRequest(method string, reqURL string) (*http.Response, *error) {
	req, err := http.NewRequest(method)
	if err != nil {
		return nil, err
	}
	for key, val := range t.headers() {
		req.Header.Add(key, val)
	}
	return req, nil
}

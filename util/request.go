package util

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var UserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:88.0) Gecko/20100101 Firefox/88.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:87.0) Gecko/20100101 Firefox/87.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:85.0) Gecko/20100101 Firefox/85.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0",
}

func init() {
	rand.Seed(time.Now().Unix())
}

func RandUA() string {
	return UserAgents[rand.Int()%len(UserAgents)]
}

func GetResponse(URL string, referer *string) (*http.Response, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	if referer != nil {
		req.Header.Add("referer", *referer)
	}
	req.Header.Add("user-agent", RandUA())
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("get %s: %s", URL, resp.Status)
	}
	return resp, nil
}

func GetText(URL string) (string, error) {
	resp, err := GetResponse(URL, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ctt, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(ctt), nil
}

func GetHtml(URL string) (*goquery.Document, error) {
	resp, err := GetResponse(URL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return goquery.NewDocumentFromReader(resp.Body)
}

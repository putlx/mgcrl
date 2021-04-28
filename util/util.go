package util

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
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
		return nil, errors.New("response status " + resp.Status)
	}
	return resp, nil
}

func GetText(URL string) (string, error) {
	resp, err := GetResponse(URL, nil)
	if err != nil {
		return "", err
	}
	ctt, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return "", err
	}
	err = resp.Body.Close()
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
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func EvalJavaScript(script string) string {
	val, err := otto.New().Run(script)
	if err != nil {
		panic(err)
	}
	s, err := val.ToString()
	if err != nil {
		panic(err)
	}
	return s
}

func GbkToUtf8(s string) string {
	reader := transform.NewReader(bytes.NewReader([]byte(s)), simplifiedchinese.GBK.NewDecoder())
	data, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func Reverse(a interface{}) {
	n := reflect.ValueOf(a).Len()
	swap := reflect.Swapper(a)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func AsFileBasename(name string) string {
	return strings.Map(func(c rune) rune {
		if strings.ContainsRune(`"\/:*?<>|`, c) {
			return ' '
		}
		return c
	}, name)
}

package ext

import (
	"encoding/base64"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/putlx/mgcrl/util"
)

func Fzdm(URL, _ string) (m Manga, err error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return
	}
	m.Title = html.Find("#content h2").Text()
	if i := strings.Index(m.Title, "漫畫"); i != -1 {
		m.Title = m.Title[:i]
	} else if i = strings.Index(m.Title, "漫画"); i != -1 {
		m.Title = m.Title[:i]
	}
	if URL[len(URL)-1] != '/' {
		URL += "/"
	}
	html.Find("#content li a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			panic(`no attribute "href"`)
		}
		m.Chapters = append(m.Chapters, Chapter{URL + href, s.Text()})
	})
	util.Reverse(m.Chapters)
	return
}

func FzdmImages(URL string) ([]string, error) {
	text, err := util.GetText(URL)
	if err != nil {
		return nil, err
	}
	script := regexp.MustCompile(`var temps = "(.+?)"`).FindStringSubmatch(text)[1]
	data, err := base64.StdEncoding.DecodeString(script)
	if err != nil {
		panic(err)
	}
	imgs := strings.Split(string(data), "\r\n")
	for i := range imgs {
		if regexp.MustCompile(`^20(16|17|18|19|20|21)`).MatchString(imgs[i]) {
			imgs[i] = "http://p1.manhuapan.com/" + imgs[i]
		} else {
			imgs[i] = "http://p5.manhuapan.com/" + imgs[i]
		}
	}
	return imgs, nil
}

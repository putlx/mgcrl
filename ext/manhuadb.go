package ext

import (
	"encoding/base64"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/putlx/mgcrl/util"
)

func Manhuadb(URL, _ string) (m Manga, err error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return
	}
	m.Title = html.Find("h1.comic-title").Text()
	html.Find("div#comic-book-list div ol li a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			panic(`no attribute "href"`)
		}
		m.Chapters = append(m.Chapters, Chapter{"https://www.manhuadb.com" + href, s.Text()})
	})
	return
}

func ManhuadbImages(URL string) ([]string, error) {
	text, err := util.GetText(URL)
	if err != nil {
		return nil, err
	}

	script := regexp.MustCompile(`var img_data = '(.+?)'`).FindStringSubmatch(text)[1]
	data, err := base64.StdEncoding.DecodeString(script)
	if err != nil {
		panic(err)
	}

	pre := regexp.MustCompile(`data-img_pre="(.+?)"`).FindStringSubmatch(text)[1]
	mat := regexp.MustCompile(`"img":"(.+?)"`).FindAllSubmatch(data, -1)
	imgs := make([]string, len(mat))
	for i := range mat {
		imgs[i] = "https://i1.manhuadb.com" + pre + string(mat[i][1])
	}
	return imgs, nil
}

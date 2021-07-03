package ext

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/putlx/mgcrl/util"
)

func Guoguomh(URL, _ string) (m Manga, err error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return
	}
	m.Title = html.Find("h1").Text()
	html.Find("#chapter-list-1 li a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			panic(`no attribute "href"`)
		}
		m.Chapters = append(m.Chapters, Chapter{
			"https://mh.guoguomh.com" + href,
			strings.TrimSpace(s.Find("span.list_con_zj").First().Text()),
		})
	})
	return
}

func GuoguomhImages(URL string) ([]string, error) {
	text, err := util.GetText(URL)
	if err != nil {
		return nil, err
	}
	text = regexp.MustCompile(`var chapterImages = \[(.*?)\]`).FindStringSubmatch(text)[1]
	mat := regexp.MustCompile(`"(.+?)"`).FindAllStringSubmatch(text, -1)
	imgs := make([]string, len(mat))
	for i := range imgs {
		imgs[i] = strings.Replace(mat[i][1], "\\", "", -1)
	}
	return imgs, nil
}

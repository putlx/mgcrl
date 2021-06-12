package ext

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/putlx/mgcrl/util"
)

func Mm1316(URL, _ string) (m Manga, err error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return
	}
	m.Title = html.Find("#intro_l div.title h1").Text()
	html.Find("#chapter-list-1 li a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			panic(`no attribute "href"`)
		}
		m.Chapters = append(m.Chapters, Chapter{"https://www.mm1316.com/" + href, s.Text()})
	})
	return
}

func Mm1316Images(URL string) ([]string, error) {
	text, err := util.GetText(URL)
	if err != nil {
		return nil, err
	}
	script := regexp.MustCompile(`chapterImages = \[(.+?)\]`).FindStringSubmatch(text)[1]
	mat := regexp.MustCompile(`"(.+?)"`).FindAllStringSubmatch(script, -1)
	imgs := make([]string, len(mat))
	for i := range mat {
		imgs[i] = "https://res.mm1316.com" + strings.ReplaceAll(mat[i][1], "\\", "")
	}
	return imgs, nil
}

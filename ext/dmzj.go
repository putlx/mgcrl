package ext

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/putlx/mgcrl/util"
)

func Dmzj(URL, _ string) (m Manga, err error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return
	}
	m.Title = html.Find("span.anim_title_text a h1").Text()
	host := regexp.MustCompile(`dmzj1?\.com`).FindString(URL)
	html.Find("div.cartoon_online_border ul li a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			panic(`no attribute "href"`)
		}
		m.Chapters = append(m.Chapters, Chapter{"https://manhua." + host + href, s.Text()})
	})
	return
}

func DmzjImages(URL string) ([]string, error) {
	text, err := util.GetText(URL)
	if err != nil {
		return nil, err
	}
	script := regexp.MustCompile(`eval(.+)`).FindStringSubmatch(text)[1]
	script = util.EvalJavaScript(script)
	script = regexp.MustCompile(`'(.+)'`).FindStringSubmatch(script)[1]
	mat := regexp.MustCompile(`"(.+?)"`).FindAllStringSubmatch(script, -1)
	imgs := make([]string, len(mat))
	host := regexp.MustCompile(`dmzj1?\.com`).FindString(URL)
	for i := range mat {
		imgs[i] = "https://images." + host + "/" + strings.ReplaceAll(mat[i][1], "\\", "")
	}
	return imgs, nil
}

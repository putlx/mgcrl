package ext

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/putlx/mgcrl/util"
)

func Dmzj2(URL, _ string) (m Manga, err error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return
	}
	m.Title = html.Find("div.comic_deCon h1 a").Text()
	html.Find("div.zj_list.autoHeight div:nth-child(4) ul li a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			panic(`no attribute "href"`)
		}
		m.Chapters = append(m.Chapters, Chapter{href, s.Find("span.list_con_zj").Text()})
	})
	return
}

func Dmzj2Images(URL string) ([]string, error) {
	text, err := util.GetText(URL)
	if err != nil {
		return nil, err
	}
	script := regexp.MustCompile(`eval(.+)`).FindStringSubmatch(text)[1]
	script = util.EvalJavaScript(script)
	script = regexp.MustCompile(`"page_url":"(.+?)"`).FindStringSubmatch(script)[1]
	imgs := strings.Split(script, `\r\n`)
	host := regexp.MustCompile(`dmzj1?\.com`).FindString(URL)
	for i := range imgs {
		imgs[i] = "https://images." + host + "/" + strings.ReplaceAll(imgs[i], "\\", "")
	}
	return imgs, nil
}

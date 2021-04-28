package ext

import (
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/putlx/mgcrl/util"
)

func Manhua123(URL, _ string) (m Manga, err error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return
	}
	m.Title = html.Find("body div.dbox div.data h4").Text()
	html.Find("body div.tbox.tabs div.tabs_block ul li a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			panic(`no attribute "href"`)
		}
		m.Chapters = append(m.Chapters, Chapter{"https://m.manhua123.net" + href, s.Text()})
	})
	util.Reverse(m.Chapters)
	return
}

func Manhua123Images(URL string) ([]string, error) {
	text, err := util.GetText(URL)
	if err != nil {
		return nil, err
	}
	script := regexp.MustCompile(`z_img='\[(.+?)\]'`).FindStringSubmatch(text)[1]
	mat := regexp.MustCompile(`"(.+?)"`).FindAllStringSubmatch(script, -1)
	imgs := make([]string, len(mat))
	for i := range mat {
		imgs[i] = "https://img.qinyiku.com/" + mat[i][1]
	}
	return imgs, nil
}

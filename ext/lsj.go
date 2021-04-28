package ext

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/putlx/mgcrl/util"
)

func Lsj(URL, _ string) (m Manga, err error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return
	}
	m.Title = html.Find("body main div.row.detailsTitle div div span.detailsName").Text()
	html.Find("body main div.row.detailsList div div a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			panic(`no attribute "href"`)
		}
		m.Chapters = append(m.Chapters, Chapter{"https://lsj.ac" + href, s.Text()})
	})
	return
}

func LsjImages(URL string) ([]string, error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return nil, err
	}
	var imgs []string
	html.Find("body div.contentImageList ul li img").Each(func(_ int, s *goquery.Selection) {
		src, exists := s.Attr("data-original")
		if exists {
			imgs = append(imgs, src)
		}
	})
	return imgs, nil
}

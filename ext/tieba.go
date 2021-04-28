package ext

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/putlx/mgcrl/util"
)

func Tieba(URL, _ string) (m Manga, err error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return
	}
	m.Title = strings.TrimSpace(html.Find("#j_core_title_wrap h3").Text())
	m.Chapters = []Chapter{{URL, m.Title}}
	return
}

func TiebaImages(URL string) ([]string, error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return nil, err
	}
	var imgs []string
	html.Find("#j_p_postlist div img.BDE_Image").Each(func(_ int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists && !strings.HasPrefix(src, "https://feed-image") {
			path := strings.Split(src, "/")
			imgs = append(imgs, "https://tiebapic.baidu.com/forum/pic/item/"+path[len(path)-1])
		}
	})
	return imgs, nil
}

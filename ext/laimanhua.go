package ext

import (
	"encoding/base64"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/putlx/mgcrl/util"
)

func Laimanhua(URL, _ string) (m Manga, err error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return
	}
	m.Title = util.GbkToUtf8(html.Find("#intro_l div.title h1").Text())
	html.Find("#play_0 ul li a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			panic(`no attribute "href"`)
		}
		m.Chapters = append(m.Chapters, Chapter{"https://www.laimanhua.com" + href, util.GbkToUtf8(s.Text())})
	})
	util.Reverse(m.Chapters)
	return
}

func LaimanhuaImages(URL string) ([]string, error) {
	text, err := util.GetText(URL)
	if err != nil {
		return nil, err
	}
	script := regexp.MustCompile(`picTree\s*=\s*'(.+)'`).FindStringSubmatch(text)[1]
	data, err := base64.StdEncoding.DecodeString(script)
	if err != nil {
		panic(err)
	}
	imgs := strings.Split(string(data), "$qingtiandy$")
	for i := range imgs {
		imgs[i] = "https://mhpic5.gezhengzhongyi.cn:8443" + imgs[i]
	}
	return imgs, nil
}

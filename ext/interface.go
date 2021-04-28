package ext

import (
	"regexp"
	"time"
)

type Chapter struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type Manga struct {
	Title    string        `json:"title"`
	Delay    time.Duration `json:"delay"`
	Chapters []Chapter     `json:"chapters"`
}

type Extractor struct {
	URLRegexp *regexp.Regexp
	GetManga  func(string, string) (Manga, error)
	GetImages func(string) ([]string, error)
}

var Extractors = []Extractor{
	{regexp.MustCompile(`^https?://manhua\.dmzj1?\.com/\w+/?$`), Dmzj, DmzjImages},
	{regexp.MustCompile(`^https?://www\.dmzj1?\.com/info/\w+\.html`), Dmzj2, Dmzj2Images},
	{regexp.MustCompile(`^https?://www\.700mh\.com/manhua/\d+/?$`), Katui, KatuiImages},
	{regexp.MustCompile(`^https?://lsj\.ac/comic/\w+$`), Lsj, LsjImages},
	{regexp.MustCompile(`^https?://www\.laimanhua\.com/kanmanhua/\d+/?$`), Laimanhua, LaimanhuaImages},
	{regexp.MustCompile(`^https?://mangadex\.org/title/\d+`), Mangadex, MangadexImages},
	{regexp.MustCompile(`^https?://m\.manhua123\.net/comic/\d+\.html$`), Manhua123, Manhua123Images},
	{regexp.MustCompile(`^https?://www\.manhuadb\.com/manhua/\d+/?$`), Manhuadb, ManhuadbImages},
	{regexp.MustCompile(`^https?://(www\.manhuagui|tw\.manhuagui|www\.mhgui)\.com/comic/\d+/?$`), Manhuagui, ManhuaguiImages},
	{regexp.MustCompile(`^https?://tieba\.baidu\.com/p/\d+`), Tieba, TiebaImages},
}

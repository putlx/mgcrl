package ext

import (
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/putlx/mgcrl/util"
)

func Mangadex(URL, group string) (m Manga, err error) {
	URL = "https://mangadex.org/api/manga/" + regexp.MustCompile(`title/(\d+)`).FindStringSubmatch(URL)[1]
	text, err := util.GetText(URL)
	if err != nil {
		return
	}

	var manga struct {
		Manga struct {
			Title string
		}
		Chapter map[string]struct {
			Chapter   string
			Title     string
			Volume    string
			GroupName string `json:"group_name"`
		}
	}
	err = json.Unmarshal([]byte(text), &manga)
	if err != nil {
		panic(err)
	}

	m.Title = manga.Manga.Title
	m.Delay = 700 * time.Millisecond

	ids := make([]int, 0, len(manga.Chapter))
	for i := range manga.Chapter {
		id, err := strconv.Atoi(i)
		if err != nil {
			panic(err)
		}
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for i := range ids {
		id := strconv.Itoa(ids[i])
		ch := manga.Chapter[id]
		if len(group) == 0 || ch.GroupName == group {
			title := "Ch. " + ch.Chapter
			if len(ch.Volume) != 0 {
				title = "Vol. " + ch.Volume + " " + title
			}
			if len(ch.Title) != 0 {
				title += " - " + ch.Title
			}
			m.Chapters = append(m.Chapters, Chapter{"https://mangadex.org/api/chapter/" + id, title})
		}
	}

	return
}

func MangadexImages(URL string) ([]string, error) {
	text, err := util.GetText(URL)
	if err != nil {
		return nil, err
	}

	var chapter struct {
		Hash      string
		Server    string
		PageArray []string `json:"page_array"`
	}
	err = json.Unmarshal([]byte(text), &chapter)
	if err != nil {
		panic(err)
	}

	imgs := make([]string, len(chapter.PageArray))
	for i := range chapter.PageArray {
		imgs[i] = chapter.Server + chapter.Hash + "/" + chapter.PageArray[i]
	}
	return imgs, nil
}

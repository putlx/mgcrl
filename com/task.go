package com

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/putlx/mgcrl/ext"
	"github.com/putlx/mgcrl/util"
)

type Task struct {
	ext.Extractor
	ext.Manga
	URL      string
	Output   string
	MaxRetry int
}

type Progress struct {
	Size      int64 `json:"size"`
	TotalSize int64 `json:"total_size"`
}

type Error struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
}

func NewTask(URL, ver, output string, maxRetry int) (*Task, error) {
	for _, e := range ext.Extractors {
		if e.URLRegexp.MatchString(URL) {
			m, err := e.GetManga(URL, ver)
			if err != nil {
				return nil, err
			}
			m.Title = util.AsFileBasename(m.Title)
			for i := range m.Chapters {
				m.Chapters[i].Title = util.AsFileBasename(m.Chapters[i].Title)
			}
			return &Task{e, m, URL, output, maxRetry}, nil
		}
	}
	return nil, errors.New("unsupported URL")
}

var extFinder = regexp.MustCompile(`(\.\w+)[^\.]*$`)

func (t *Task) GetChapter(i int) (<-chan Progress, <-chan *Error, error) {
	c := t.Chapters[i]
	output := path.Join(t.Output, t.Title, c.Title)
	if err := os.MkdirAll(output, os.ModeDir); err != nil {
		return nil, nil, err
	}
	imgs, err := t.GetImages(c.URL)
	if err != nil {
		return nil, nil, err
	}

	s := make(chan Progress)
	e := make(chan *Error, len(imgs))
	feedSize := make(chan int64)
	feedTotalSize := make(chan int64)
	finish := make(chan struct{})

	for i := range imgs {
		go func(i int) {
			URL, err := url.Parse(imgs[i])
			if err != nil {
				panic(err)
			}
			ext := extFinder.FindStringSubmatch(URL.Path)[1]
			filename := fmt.Sprintf("%03d", i+1) + ext
			fullname := path.Join(output, filename)

			if t.Delay > 0 {
				time.Sleep(time.Duration(i) * t.Delay)
			}
			for s := -1; t.MaxRetry < 0 || s < t.MaxRetry; s++ {
				if t.Delay > 0 {
					time.Sleep(t.Delay + time.Duration(rand.Int()%100)*time.Millisecond)
				}
				err = func() error {
					resp, err := util.GetResponse(imgs[i], &c.URL)
					if err != nil {
						return err
					}
					if resp.ContentLength > 0 {
						feedTotalSize <- resp.ContentLength
					}

					file, err := os.Create(fullname)
					if err != nil {
						return err
					}
					for written := int64(0); err == nil; feedSize <- written {
						written, err = io.CopyN(file, resp.Body, 65536)
					}

					if err != io.EOF {
						resp.Body.Close()
						file.Close()
						return err
					} else if err = resp.Body.Close(); err != nil {
						file.Close()
						return err
					} else if err = file.Close(); err != nil {
						return err
					}
					return nil
				}()
				if err == nil {
					break
				}
			}
			if err != nil {
				e <- &Error{filename, err.Error()}
			}
			finish <- struct{}{}
		}(i)
	}

	go func() {
		progress := Progress{}
		s <- progress
		for r := len(imgs); r > 0; {
			select {
			case <-finish:
				r--
			case n := <-feedSize:
				progress.Size += n
				s <- progress
			case n := <-feedTotalSize:
				progress.TotalSize += n
				s <- progress
			}
		}
		close(s)
		close(e)
	}()
	return s, e, nil
}

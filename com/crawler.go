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

type Crawler struct {
	ext.Extractor
	ext.Manga
	URL      string
	Output   string
	MaxRetry uint
}

type Progress struct {
	Size      int64 `json:"size"`
	TotalSize int64 `json:"total_size"`
}

type Error struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
}

func NewCrawler(URL, version, output string, maxRetry uint) (*Crawler, error) {
	for _, e := range ext.Extractors {
		if e.URLRegexp.MatchString(URL) {
			var lastErr error
			for t := uint(0); t < maxRetry+1; t++ {
				m, err := e.GetManga(URL, version)
				if err == nil {
					return &Crawler{e, m, URL, output, maxRetry}, nil
				}
				lastErr = err
			}
			return nil, lastErr
		}
	}
	return nil, errors.New("unsupported URL")
}

var extFinder = regexp.MustCompile(`(\.\w+)[^\.]*$`)

func (c *Crawler) FetchChapter(idx int) (<-chan Progress, <-chan *Error, <-chan struct{}) {
	prg := make(chan Progress)
	errs := make(chan *Error)
	done := make(chan struct{})

	output := path.Join(
		c.Output,
		util.AsFileBasename(c.Title),
		util.AsFileBasename(c.Chapters[idx].Title),
	)
	err := os.MkdirAll(output, os.ModeDir)
	if err != nil {
		go func() {
			errs <- &Error{"", err.Error()}
			done <- struct{}{}
		}()
		return prg, errs, done
	}

	imgs, err := c.GetImages(c.Chapters[idx].URL)
	if err != nil {
		go func() {
			errs <- &Error{"", err.Error()}
			done <- struct{}{}
		}()
		return prg, errs, done
	}

	size := make(chan int64)
	totalSize := make(chan int64)
	fin := make(chan struct{})

	for i := range imgs {
		go func(i int) {
			URL, err := url.Parse(imgs[i])
			if err != nil {
				panic(err)
			}
			name := fmt.Sprintf("%03d", i+1) + extFinder.FindStringSubmatch(URL.Path)[1]

			for t := uint(0); t < c.MaxRetry+1; t++ {
				if c.Delay > 0 {
					if t == 0 {
						time.Sleep(time.Duration(i) * c.Delay)
					} else {
						time.Sleep(c.Delay + time.Duration(rand.Int()%100)*time.Millisecond)
					}
				}
				if err = func() error {
					resp, err := util.GetResponse(imgs[i], &c.Chapters[idx].URL)
					if err != nil {
						return err
					}
					if resp.ContentLength > 0 {
						totalSize <- resp.ContentLength
					}

					file, err := os.Create(path.Join(output, name))
					if err != nil {
						return err
					}
					for written := int64(0); err == nil; size <- written {
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
				}(); err == nil {
					break
				}
			}
			if err != nil {
				errs <- &Error{name, err.Error()}
			}
			fin <- struct{}{}
		}(i)
	}

	go func() {
		s := Progress{}
		for r := len(imgs); r > 0; {
			select {
			case <-fin:
				r--
			case n := <-size:
				s.Size += n
				prg <- s
			case n := <-totalSize:
				s.TotalSize += n
				prg <- s
			}
		}
		done <- struct{}{}
	}()
	return prg, errs, done
}

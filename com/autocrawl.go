package com

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gen2brain/beeep"
)

type Config struct {
	Assets []struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		Version     string `json:"version"`
		LastChapter string `json:"last_chapter"`
	} `json:"assets"`
	Frequency int    `json:"frequency_in_hour"`
	Output    string `json:"output"`
}

func (c *Config) Read(filename string) error {
	if data, err := os.ReadFile(filename); err != nil {
		return err
	} else if err = json.Unmarshal(data, c); err != nil {
		return err
	} else if len(c.Assets) == 0 {
		return errors.New("no asset")
	}
	return nil
}

func (c *Config) WriteTo(filename string) error {
	if data, err := json.MarshalIndent(c, "", "    "); err != nil {
		return err
	} else if err = os.WriteFile(filename, data, 0666); err != nil {
		return err
	}
	return nil
}

func AutoCrawl(filename string) {
	const maxTry = 4
	const duration = time.Duration(10) * time.Minute

	config := &Config{}
	if err := config.Read(filename); err != nil {
		beeep.Notify("mgcrl - Error", err.Error(), "")
		log.Fatalln(err)
	}
	if len(config.Output) == 0 {
		config.Output = "."
		log.Println("set output to .")
	}
	if config.Frequency == 0 {
		config.Frequency = 12
		log.Println("set frequency to 12 hours")
	}

	for {
		for i := range config.Assets {
			a := &config.Assets[i]
			var c *Crawler
			var err error
			for t := 0; ; t++ {
				if t != 0 {
					time.Sleep(duration)
					log.Printf("retry getting %s\n", a.URL)
				}
				c, err = NewCrawler(a.URL, a.Version, config.Output, maxTry)
				if err == nil {
					break
				}
				log.Println(err)
			}

			if c == nil {
				continue
			} else if len(a.LastChapter) == 0 {
				if len(c.Chapters) > 0 {
					a.LastChapter = c.Chapters[len(c.Chapters)-1].Title
					if err := config.WriteTo(filename); err != nil {
						log.Println(err)
						beeep.Notify("mgcrl - Error", err.Error(), "")
					}
				}
				continue
			}

			last := len(c.Chapters) - 1
			for i := range c.Chapters {
				if a.LastChapter == c.Chapters[i].Title {
					last = i
					break
				}
			}
			for idx := last + 1; idx < len(c.Chapters); idx++ {
				for t := 0; t < maxTry; t++ {
					if t != 0 {
						time.Sleep(duration)
						log.Printf("retry crawling 「%s / %s」\n", c.Title, c.Chapters[idx].Title)
					}
					prg, errs, done := c.FetchChapter(idx)
					var err *Error
				loop:
					for {
						select {
						case <-prg:
						case err = <-errs:
							log.Println(err)
						case <-done:
							break loop
						}
					}
					if err == nil {
						title := fmt.Sprintf("「%s / %s」", c.Title, c.Chapters[idx].Title)
						log.Println(title + " is downloaded")
						beeep.Notify("mgcrl - Notification", title+"下载完毕。", "")
						a.LastChapter = c.Chapters[idx].Title
						if err := config.WriteTo(filename); err != nil {
							log.Println(err)
							beeep.Notify("mgcrl - Error", err.Error(), "")
						}
						break
					}
				}
			}
		}
		time.Sleep(time.Duration(config.Frequency) * time.Hour)
	}
}

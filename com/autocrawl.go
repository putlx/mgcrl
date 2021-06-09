package com

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gen2brain/beeep"
)

type Asset struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Version     string `json:"version"`
	LastChapter string `json:"last_chapter"`
}

type Config struct {
	Assets    []Asset `json:"assets"`
	Frequency int     `json:"frequency_in_hour"`
	Output    string  `json:"output"`
}

func NewConfig(filename string) (*Config, error) {
	var c Config
	if data, err := os.ReadFile(filename); err != nil {
		return nil, err
	} else if err = json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if len(c.Output) == 0 {
		c.Output = "."
	}
	if c.Frequency <= 0 {
		c.Frequency = 6
	}
	return &c, nil
}

func (c *Config) WriteTo(filename string) error {
	if data, err := json.MarshalIndent(c, "", "    "); err != nil {
		return err
	} else if err = os.WriteFile(filename, data, 0666); err != nil {
		return err
	}
	return nil
}

func AutoCrawl(configFile string, log *log.Logger) {
	const maxTry = 4
	const duration = time.Duration(10) * time.Minute

	for {
		config, err := NewConfig(configFile)
		if err != nil {
			beeep.Notify("错误", err.Error(), "")
			log.Fatalln(err)
		}

		for i := range config.Assets {
			a := &config.Assets[i]
			var c *Crawler
			var err error
			for t := 0; ; t++ {
				if t != 0 {
					time.Sleep(duration)
					log.Println("retry getting " + a.URL)
				}
				c, err = NewCrawler(a.URL, a.Version, config.Output, maxTry)
				if err == nil {
					break
				}
				log.Println(err)
			}

			if c == nil {
				log.Println("fail to get " + a.URL)
				continue
			} else if len(a.LastChapter) == 0 {
				if len(c.Chapters) > 0 {
					a.LastChapter = c.Chapters[len(c.Chapters)-1].Title
					if err := config.WriteTo(configFile); err != nil {
						log.Println(err)
						beeep.Notify("错误", err.Error(), "")
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
						beeep.Notify("下载完成", title+"下载完毕。", "")
						a.LastChapter = c.Chapters[idx].Title
						if err := config.WriteTo(configFile); err != nil {
							log.Println(err)
							beeep.Notify("错误", err.Error(), "")
						}
						break
					}
				}
			}
		}
		time.Sleep(time.Duration(config.Frequency) * time.Hour)
	}
}

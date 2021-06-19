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
	Name          string `json:"name"`
	URL           string `json:"url"`
	Version       string `json:"version"`
	LatestChapter string `json:"latest_chapter"`
}

type Config struct {
	Assets    []Asset `json:"assets"`
	Frequency int     `json:"frequency_in_hour"`
	Output    string  `json:"output"`
	filename  string
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
	c.filename = filename
	return &c, nil
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.filename, data, 0666)
}

func AutoCrawl(configFile string, log *log.Logger) {
	const maxRetry = 3
	const duration = time.Duration(10) * time.Minute

	for {
		config, err := NewConfig(configFile)
		if err != nil {
			beeep.Notify("错误", err.Error(), "")
			log.Fatalln(err)
		}

		for i := range config.Assets {
			a := &config.Assets[i]
			c, err := NewCrawler(a.URL, a.Version, config.Output, maxRetry)
			if err != nil {
				log.Println(err)
				continue
			} else if len(a.LatestChapter) == 0 {
				if len(c.Chapters) > 0 {
					a.LatestChapter = c.Chapters[len(c.Chapters)-1].Title
					if err := config.Save(); err != nil {
						log.Println(err)
						beeep.Notify("错误", err.Error(), "")
					}
				}
				continue
			}

			latest := len(c.Chapters) - 1
			for i := range c.Chapters {
				if a.LatestChapter == c.Chapters[i].Title {
					latest = i
					break
				}
			}
			for idx := latest + 1; idx < len(c.Chapters); idx++ {
				title := fmt.Sprintf("「%s / %s」", c.Title, c.Chapters[idx].Title)
				for t := -1; t < maxRetry; t++ {
					if t >= 0 {
						time.Sleep(duration)
						log.Println("retry crawling " + title)
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
						log.Println(title + " is downloaded")
						beeep.Notify("下载完成", title+"下载完毕。", "")
						a.LatestChapter = c.Chapters[idx].Title
						if err := config.Save(); err != nil {
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

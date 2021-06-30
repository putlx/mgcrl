package com

import (
	"fmt"
	"log"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/putlx/mgcrl/util"
)

func AutoCrawl(csv, output string, frequency, maxRetry uint, log *log.Logger) {
	const duration = time.Duration(10) * time.Minute

	for {
		records, err := util.ReadCSV(csv)
		if err != nil {
			beeep.Notify("错误", err.Error(), "")
			log.Fatalln(err)
		}

		log.Println("checking for updates")
		for _, record := range records {
			if len(record) != 4 {
				const msg = "each record should have 4 fields"
				log.Println(msg)
				beeep.Notify("错误", msg, "")
				continue
			}
			c, err := NewCrawler(record[1], record[2], output, maxRetry)
			if err != nil {
				log.Println(err)
				continue
			} else if len(record[3]) == 0 {
				if len(c.Chapters) > 0 {
					record[3] = c.Chapters[len(c.Chapters)-1].Title
					if err := util.WriteCSV(csv, records); err != nil {
						log.Println(err)
						beeep.Notify("错误", err.Error(), "")
					}
				}
				continue
			}

			latest := len(c.Chapters) - 1
			for i := range c.Chapters {
				if record[3] == c.Chapters[i].Title {
					latest = i
					break
				}
			}
			for idx := latest + 1; idx < len(c.Chapters); idx++ {
				title := fmt.Sprintf("「%s / %s」", c.Title, c.Chapters[idx].Title)
				for t := uint(0); t < maxRetry+1; t++ {
					if t > 0 {
						time.Sleep(duration)
					}
					prg, errs, done := c.FetchChapter(idx)
					var err *Error
				progress:
					for {
						select {
						case <-prg:
						case err = <-errs:
							log.Println(err)
						case <-done:
							break progress
						}
					}
					if err == nil {
						log.Println(title, "is downloaded")
						beeep.Notify("下载完成", title+"下载完毕。", "")
						record[3] = c.Chapters[idx].Title
						if err := util.WriteCSV(csv, records); err != nil {
							log.Println(err)
							beeep.Notify("错误", err.Error(), "")
						}
						break
					} else if t == maxRetry {
						log.Println("fail to get", title)
					}
				}
			}
		}
		log.Println("all updates are complete")
		time.Sleep(time.Duration(frequency) * time.Hour)
	}
}

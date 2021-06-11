package com

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gen2brain/beeep"
)

func AutoCrawl(db *sql.DB) error {
	log := func(m string) {
		db.Exec("INSERT INTO ac_log VALUES (?, ?)", time.Now().Unix(), m)
	}

	getConfigItem := func(attr string, val interface{}) error {
		rows, err := db.Query("SELECT val FROM config WHERE attr = ?", attr)
		if err != nil {
			return err
		}
		defer rows.Close()
		rows.Next()
		return rows.Scan(val)
	}

	for {
		var maxRetry, duration int
		var output string
		if err := getConfigItem("max_retry", &maxRetry); err != nil {
			return err
		} else if err = getConfigItem("duration", &duration); err != nil {
			return err
		} else if err = getConfigItem("output", &output); err != nil {
			return err
		}

		rows, err := db.Query("SELECT * FROM assets")
		if err != nil {
			return err
		}
		for rows.Next() {
			var name, url, version, lastChapter string
			if err = rows.Scan(&name, &url, &version, &lastChapter); err != nil {
				return err
			}
			c, err := NewCrawler(url, version, output, maxRetry)
			if err != nil {
				log(err.Error())
				continue
			} else if len(lastChapter) == 0 {
				if len(c.Chapters) > 0 {
					lastChapter = c.Chapters[len(c.Chapters)-1].Title
					_, err = db.Exec("UPDATE assets SET last_chapter = ? WHERE name = ?", lastChapter, name)
					if err != nil {
						log(err.Error())
					}
				}
				continue
			}

			last := len(c.Chapters) - 1
			for i := range c.Chapters {
				if lastChapter == c.Chapters[i].Title {
					last = i
					break
				}
			}
			for idx := last + 1; idx < len(c.Chapters); idx++ {
				title := fmt.Sprintf("「%s / %s」", c.Title, c.Chapters[idx].Title)
				for t := -1; t < maxRetry; t++ {
					if t >= 0 {
						time.Sleep(time.Duration(duration) * time.Second)
						log("retry crawling " + title)
					}
					prg, errs, done := c.FetchChapter(idx)
					var err *Error
				loop:
					for {
						select {
						case <-prg:
						case err = <-errs:
							log("get " + err.Filename + ": " + err.Error)
						case <-done:
							break loop
						}
					}
					if err == nil {
						log(title + " is downloaded")
						beeep.Notify("下载完成", title+"下载完毕。", "")
						lastChapter = c.Chapters[idx].Title
						_, err := db.Exec("UPDATE assets SET last_chapter = ? WHERE name = ?", lastChapter, name)
						if err != nil {
							log(err.Error())
						}
						break
					}
				}
			}
		}
		if err = rows.Close(); err != nil {
			log(err.Error())
		}

		var freq int
		if err = getConfigItem("freq_in_hour", &freq); err != nil {
			return err
		}
		time.Sleep(time.Duration(freq) * time.Hour)
	}
}

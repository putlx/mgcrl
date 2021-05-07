package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/putlx/mgcrl/com"
	"github.com/putlx/mgcrl/ext"
)

const (
	RESET  = "\033[0m"
	RED    = "\033[31m"
	GREEN  = "\033[32m"
	YELLOW = "\033[33m"
	BLUE   = "\033[34m"
	PURPLE = "\033[35m"
)

func main() {
	enabled := true
	defer colorable.EnableColorsStdout(&enabled)()

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: mgcrl [options] URL")
		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		flag.PrintDefaults()
	}
	var ver, selector, output string
	var maxRetry int
	flag.StringVar(&ver, "v", "", "manga version")
	flag.StringVar(&selector, "c", "1:-1", "volumes or chapters")
	flag.StringVar(&output, "o", ".", "output directory")
	flag.IntVar(&maxRetry, "m", 3, "max retry time")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		return
	}

	c, err := com.NewCrawler(flag.Args()[0], ver, output, maxRetry)
	if err != nil {
		fmt.Println(RED + err.Error() + RESET)
		return
	}

	if len(c.Chapters) == 0 {
		fmt.Println(YELLOW + "no volume/chapter found" + RESET)
		return
	} else if cs, ok := filter(selector, c.Chapters); !ok {
		fmt.Println(RED + "bad range" + RESET)
		return
	} else if len(cs) == 0 {
		fmt.Println(YELLOW + "no volume/chapter selected" + RESET)
		return
	} else {
		c.Chapters = cs
	}

	if len(c.Chapters) > 1 {
		fmt.Printf("(%d items) %s%s%s\n", len(c.Chapters), BLUE, c.Title, RESET)
	} else {
		fmt.Printf("(%d item) %s%s%s\n", len(c.Chapters), BLUE, c.Title, RESET)
	}
	for i := range c.Chapters {
		prg, errs, done := c.FetchChapter(i)
		var errors []*com.Error
	loop:
		for hasPrg := false; ; {
			select {
			case s := <-prg:
				fmt.Printf("\r[%s%03d%s / %s%03d%s] %s%s%s [%s%.1fMB%s / %s%.1fMB%s]",
					GREEN, i+1, RESET, GREEN, len(c.Chapters), RESET,
					BLUE, c.Chapters[i].Title, RESET,
					PURPLE, float64(s.Size)/1048576, RESET,
					PURPLE, float64(s.TotalSize)/1048576, RESET)
				hasPrg = true
			case err := <-errs:
				errors = append(errors, err)
			case <-done:
				if hasPrg {
					fmt.Println()
				}
				break loop
			}
		}
		for _, err := range errors {
			fmt.Printf("[%s%03d%s / %s%03d%s] %s%s%s / %s%s%s : %s%s%s\n",
				RED, i+1, RESET, RED, len(c.Chapters), RESET,
				BLUE, c.Chapters[i].Title, RESET,
				BLUE, err.Filename, RESET,
				RED, err.Error, RESET)
		}
	}
}

func filter(selector string, chapters []ext.Chapter) ([]ext.Chapter, bool) {
	var cs []ext.Chapter
	for _, c := range strings.Split(selector, ",") {
		s := strings.Split(c, ":")
		if len(s) == 0 || len(s) > 2 {
			return nil, false
		}
		var begin, end int
		var err error
		if begin, err = strconv.Atoi(s[0]); err != nil {
			begin = -1
		} else if begin <= 0 {
			begin += len(chapters)
		} else {
			begin--
		}
		if len(s) == 2 {
			if end, err = strconv.Atoi(s[1]); err != nil {
				end = -1
			} else if end <= 0 {
				end += len(chapters) + 1
			}
		}
		for i := range chapters {
			if begin != -1 && (len(s) == 1 || end != -1) {
				break
			}
			if begin == -1 && s[0] == chapters[i].Title {
				begin = i
			}
			if len(s) == 2 && end == -1 && s[1] == chapters[i].Title {
				end = i + 1
			}
		}
		if begin == -1 {
			begin = len(chapters)
		}
		if len(s) == 2 {
			if end == -1 {
				end = len(chapters)
			}
			cs = append(cs, chapters[begin:end]...)
		} else {
			cs = append(cs, chapters[begin])
		}
	}
	return cs, true
}

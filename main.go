package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/putlx/mgcrl/com"
	"github.com/putlx/mgcrl/ext"
)

func main() {
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
	flag.IntVar(&maxRetry, "m", 5, "max retry time")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		return
	}

	enabled := true
	defer colorable.EnableColorsStdout(&enabled)()

	t, err := com.NewTask(flag.Args()[0], ver, output, maxRetry)
	if err != nil {
		fmt.Println("\033[31m" + err.Error() + "\033[0m")
		return
	}

	if len(t.Chapters) == 0 {
		fmt.Println("\033[33mno volume/chapter found\033[0m")
		return
	} else if cs, ok := filter(selector, t.Chapters); !ok {
		fmt.Println("\033[31mbad range\033[0m")
		return
	} else if len(cs) == 0 {
		fmt.Println("\033[33mno volume/chapter selected\033[0m")
		return
	} else {
		t.Chapters = cs
	}

	err = os.MkdirAll(path.Join(output, t.Title), os.ModeDir)
	if err != nil {
		fmt.Println("\033[31m" + err.Error() + "\033[0m")
		return
	}

	fmt.Printf("(\033[34m%03d\033[0m) \033[34m%s\033[0m\n", len(t.Chapters), t.Title)
	for i := range t.Chapters {
		progress, errs, err := t.GetChapter(i)
		if err != nil {
			fmt.Printf("[\033[31m%03d\033[0m / \033[31m%03d\033[0m]", i+1, len(t.Chapters))
			fmt.Printf(" \033[34m%s\033[0m", t.Chapters[i].Title)
			fmt.Printf(" \033[31m%s\033[0m\n", err)
			continue
		}
		for s := range progress {
			fmt.Print("\r")
			fmt.Printf("[\033[32m%03d\033[0m / \033[32m%03d\033[0m]", i+1, len(t.Chapters))
			fmt.Printf(" \033[34m%s\033[0m", t.Chapters[i].Title)
			fmt.Printf(" [\033[35m%.1fMB\033[0m / \033[35m%.1fMB\033[0m]", float64(s.Size)/1048576, float64(s.TotalSize)/1048576)
		}
		fmt.Println()
		for err := range errs {
			fmt.Printf("[\033[31m%03d\033[0m / \033[31m%03d\033[0m]", i+1, len(t.Chapters))
			fmt.Printf(" \033[34m%s\033[0m / \033[34m%s\033[0m", t.Chapters[i].Title, err.Filename)
			fmt.Printf(" \033[31m%s\033[0m\n", err.Error)
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

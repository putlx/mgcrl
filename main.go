package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/putlx/mgcrl/com"
	"github.com/putlx/mgcrl/ext"
	"github.com/putlx/mgcrl/webui"
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
	dlFlags := flag.NewFlagSet("get", flag.ExitOnError)
	svFlags := flag.NewFlagSet("serve", flag.ExitOnError)
	dlFlags.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: mgcrl get <URL> [options]")
		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		dlFlags.PrintDefaults()
	}
	svFlags.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: mgcrl serve <PORT> <DATABASE_NAME>")
	}
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: mgcrl get <URL> [options]")
		fmt.Fprintln(flag.CommandLine.Output(), "       mgcrl serve <PORT> <DATABASE_NAME>")
		fmt.Fprintln(flag.CommandLine.Output(), "\nOptions for get:")
		dlFlags.PrintDefaults()
	}
	var version, selector, output string
	var maxRetry int
	dlFlags.StringVar(&version, "v", "", "manga version")
	dlFlags.StringVar(&selector, "c", "1:-1", "volumes or chapters")
	dlFlags.StringVar(&output, "o", ".", "output directory")
	dlFlags.IntVar(&maxRetry, "m", 3, "max retry time")

	enabled := true
	defer colorable.EnableColorsStdout(&enabled)()

	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "-help" {
		flag.Usage()
		return
	} else if os.Args[1] == "serve" {
		if len(os.Args) < 4 {
			svFlags.Usage()
		} else if port, err := strconv.Atoi(os.Args[2]); err != nil {
			fmt.Println(RED + "invalid port" + RESET)
		} else if db, err := com.OpenDB(os.Args[3]); err != nil {
			fmt.Println(RED + err.Error() + RESET)
		} else {
			go func() {
				if err := com.AutoCrawl(db); err != nil {
					fmt.Println(RED + err.Error() + RESET)
					os.Exit(1)
				}
			}()
			webui.Serve(port, db)
		}
		return
	} else if os.Args[1] != "get" {
		fmt.Println(RED + os.Args[1] + ": unknown command" + RESET)
		return
	} else if len(os.Args) < 3 {
		dlFlags.Usage()
		return
	} else if err := dlFlags.Parse(os.Args[3:]); err != nil {
		fmt.Println(RED + err.Error() + RESET)
		return
	}

	c, err := com.NewCrawler(os.Args[2], version, output, maxRetry)
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
		fmt.Printf("%s%s%s (%d items)\n", BLUE, c.Title, RESET, len(c.Chapters))
	} else {
		fmt.Printf("%s%s%s (%d item)\n", BLUE, c.Title, RESET, len(c.Chapters))
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
		if begin != -1 {
			if len(s) == 2 {
				if end == -1 {
					end = len(chapters)
				}
				cs = append(cs, chapters[begin:end]...)
			} else {
				cs = append(cs, chapters[begin])
			}
		}
	}
	return cs, true
}

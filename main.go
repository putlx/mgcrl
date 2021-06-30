package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/putlx/mgcrl/com"
	"github.com/putlx/mgcrl/util"
	"github.com/putlx/mgcrl/webui"
)

func main() {
	var version, selector, output, csvFile, logFile string
	var maxRetry, frequency uint

	dlFlags := flag.NewFlagSet("get", flag.ExitOnError)
	dlFlags.StringVar(&version, "v", "", "manga version")
	dlFlags.StringVar(&selector, "c", "1:-1", "volumes or chapters")
	dlFlags.StringVar(&output, "o", ".", "output directory")
	dlFlags.UintVar(&maxRetry, "m", 3, "max retry time")
	dlFlags.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: mgcrl get <URL> [options]")
		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		dlFlags.PrintDefaults()
	}

	svFlags := flag.NewFlagSet("serve", flag.ExitOnError)
	svFlags.StringVar(&csvFile, "c", "", "csv file contains manga records used for auto crawling")
	svFlags.StringVar(&output, "o", ".", "output directory")
	svFlags.UintVar(&maxRetry, "m", 3, "max retry time")
	svFlags.UintVar(&frequency, "f", 6, "update frequency in hour")
	svFlags.StringVar(&logFile, "l", "", "redirect log to file")
	svFlags.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: mgcrl serve <PORT> [options]")
		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		svFlags.PrintDefaults()
	}

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: mgcrl get <URL> [options]")
		fmt.Fprintln(flag.CommandLine.Output(), "       mgcrl serve <PORT> [options]")
		fmt.Fprintln(flag.CommandLine.Output(), "       mgcrl help")
		fmt.Fprintln(flag.CommandLine.Output(), "\nOptions for get:")
		dlFlags.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output(), "\nOptions for serve:")
		svFlags.PrintDefaults()
	}

	if len(os.Args) < 2 {
		flag.Usage()
	} else if os.Args[1] == "help" {
		flag.Usage()
	} else if os.Args[1] == "get" {
		if len(os.Args) < 3 {
			dlFlags.Usage()
		} else if err := dlFlags.Parse(os.Args[3:]); err != nil {
			fmt.Println(err)
		} else {
			get(os.Args[2], version, selector, output, maxRetry)
		}
	} else if os.Args[1] == "serve" {
		if len(os.Args) < 3 {
			svFlags.Usage()
		} else if port, err := strconv.Atoi(os.Args[2]); err != nil || port < 0 {
			fmt.Println("invalid port")
		} else if err := svFlags.Parse(os.Args[3:]); err != nil {
			fmt.Println(err)
		} else {
			var w io.Writer = os.Stderr
			if len(logFile) != 0 {
				w = util.NewWriter(logFile)
			}
			if len(csvFile) != 0 {
				logger := log.New(w, "[autocrawl] ", log.LstdFlags|log.Lmsgprefix)
				go com.AutoCrawl(csvFile, output, frequency, maxRetry, logger)
			}
			logger := log.New(w, "[webui] ", log.LstdFlags|log.Lmsgprefix)
			webui.Serve(port, output, csvFile, logFile, maxRetry, frequency, logger)
		}
	} else {
		fmt.Println(os.Args[1] + ": unknown command")
	}
}

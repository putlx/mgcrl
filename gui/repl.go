package main

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/putlx/mgcrl/com"
)

type Status struct {
	com.Progress
	Index  int          `json:"index"`
	Errors []*com.Error `json:"errors"`
	Done   bool         `json:"done"`
}

type Request struct {
	URL     string `json:"url"`
	Output  string `json:"output"`
	Indexes []int  `json:"indexes"`
}

func writeLine(v interface{}, f *bufio.Writer) {
	if b, err := json.Marshal(v); err != nil {
		panic(err)
	} else if _, err = f.Write(b); err != nil {
		panic(err)
	} else if err = f.WriteByte(byte('\n')); err != nil {
		panic(err)
	} else if err = f.Flush(); err != nil {
		panic(err)
	}
}

func main() {
	var (
		stdin  = bufio.NewScanner(os.Stdin)
		stdout = bufio.NewWriter(os.Stdout)
		stderr = bufio.NewWriter(os.Stderr)
	)

	var i = 0
	var c *com.Crawler
	var sc = make(chan *Status)
	go func() {
		for s := range sc {
			writeLine(s, stderr)
		}
	}()

	for stdin.Scan() {
		req := Request{}
		err := json.Unmarshal([]byte(stdin.Text()), &req)
		if err != nil {
			panic(err)
		}

		if c == nil || req.URL != c.URL {
			cc, err := com.NewCrawler(req.URL, "", req.Output, 3)
			if err != nil {
				writeLine(err.Error(), stdout)
				continue
			}
			c = cc
		} else {
			c.Output = req.Output
		}

		if len(req.Indexes) == 0 {
			writeLine(c.Manga, stdout)
		} else {
			for _, idx := range req.Indexes {
				go func(id, idx int, c *com.Crawler) {
					prg, errs, done := c.FetchChapter(idx)
					s := &Status{Index: id}
				loop:
					for {
						select {
						case s.Progress = <-prg:
							sc <- s
						case err := <-errs:
							s.Errors = append(s.Errors, err)
						case <-done:
							s.Done = true
							sc <- s
							break loop
						}
					}
				}(i, idx, c)
				i++
			}
			writeLine(nil, stdout)
		}
	}
	panic(stdin.Err())
}

package main

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	"github.com/putlx/mgcrl/com"
)

type Status struct {
	com.Progress
	Index  int          `json:"index"`
	Errors []*com.Error `json:"errors"`
}

var (
	scanner = bufio.NewScanner(os.Stdin)
	stdout  = bufio.NewWriter(os.Stdout)
	stderr  = bufio.NewWriter(os.Stderr)
)

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

var mutex = &sync.Mutex{}

func update(s *Status) {
	mutex.Lock()
	writeLine(s, stderr)
	mutex.Unlock()
}

func main() {
	var t *com.Task
	var i = 0

	for scanner.Scan() {
		req := struct {
			URL     string `json:"url"`
			Output  string `json:"output"`
			Indexes []int  `json:"indexes"`
		}{}
		err := json.Unmarshal([]byte(scanner.Text()), &req)
		if err != nil {
			panic(err)
		}

		if t == nil || req.URL != t.URL {
			tt, err := com.NewTask(req.URL, "", req.Output, 3)
			if err != nil {
				writeLine(err.Error(), stdout)
				continue
			}
			t = tt
		} else {
			t.Output = req.Output
		}

		if len(req.Indexes) == 0 {
			writeLine(t.Manga, stdout)
		} else {
			for _, idx := range req.Indexes {
				progress, errs, err := t.GetChapter(idx)
				if err != nil {
					s := &Status{Index: i, Errors: []*com.Error{{Filename: "", Error: err.Error()}}}
					update(s)
				} else {
					go func(i int, progress <-chan com.Progress, errs <-chan *com.Error) {
						s := &Status{Index: i}
						for m := range progress {
							s.Progress = m
							update(s)
						}
						for err := range errs {
							s.Errors = append(s.Errors, err)
						}
						if s.Errors == nil {
							s.Errors = make([]*com.Error, 0)
						}
						update(s)
					}(i, progress, errs)
				}
				i++
			}
			writeLine(nil, stdout)
		}
	}
	panic(scanner.Err())
}

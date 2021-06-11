package webui

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/putlx/mgcrl/com"
)

//go:embed index.html
var html []byte

//go:embed index.js
var js []byte

//go:embed style.css
var css []byte

//go:embed reader.html
var reader []byte

//go:embed log.html
var logHtml []byte

//go:embed toast.js
var toast []byte

//go:embed autocrawl.html
var autocrawl []byte

//go:embed favicon.ico
var favicon []byte

type Task struct {
	ID      int          `json:"id"`
	Manga   string       `json:"manga"`
	Chapter string       `json:"chapter"`
	Errors  []*com.Error `json:"errors"`
	Done    bool         `json:"done"`
	mutex   *sync.Mutex
	com.Progress
}

func Serve(port int, configFile, logFile string, log *log.Logger) {
	var id = make(chan int)
	var tc = make(chan Task)
	var crawler atomic.Value
	var tasks = &sync.Map{}
	var upgrader = websocket.Upgrader{}
	go func() {
		for i := 0; ; i++ {
			id <- i
		}
	}()

	writeJSON := func(w http.ResponseWriter, v interface{}) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if data, err := json.Marshal(v); err != nil {
			log.Fatalln(err)
		} else if _, err = w.Write(data); err != nil {
			log.Println(err)
		}
	}

	readJSON := func(req *http.Request, v interface{}) error {
		if data, err := io.ReadAll(req.Body); err != nil {
			return err
		} else if err = json.Unmarshal(data, v); err != nil {
			return err
		}
		return nil
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(html)
	})

	http.HandleFunc("/index.js", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Write(js)
	})

	http.HandleFunc("/toast.js", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Write(toast)
	})

	http.HandleFunc("/style.css", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write(css)
	})

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(favicon)
	})

	http.HandleFunc("/reader", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(reader)
	})

	http.HandleFunc("/autocrawl", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(autocrawl)
	})

	http.HandleFunc("/log", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" && len(logFile) != 0 {
			if _, err := os.Stat(logFile); err == nil {
				if err = os.Remove(logFile); err != nil {
					log.Println(err)
				}
			}
			return
		}

		idx := bytes.Index(logHtml, []byte("{{}}"))
		if idx == -1 {
			panic(errors.New("unable to locate the table body in log.html"))
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if f, err := os.Open(logFile); err != nil {
			if os.IsNotExist(err) {
				w.Write(logHtml[:idx])
				w.Write(logHtml[idx+4:])
			} else {
				w.Write([]byte(err.Error()))
			}
		} else {
			defer f.Close()
			if data, err := io.ReadAll(f); err != nil {
				w.Write([]byte(err.Error()))
			} else {
				w.Write(logHtml[:idx])
				rows := bytes.Split(data, []byte("\n"))
				for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
					rows[i], rows[j] = rows[j], rows[i]
				}
				for _, row := range rows {
					if len(row) > 0 {
						w.Write([]byte(`<tr>`))
						for _, m := range regexp.MustCompile(`(\d{4}/\d{2}/\d{2}) (\d{2}:\d{2}:\d{2}) \[(\w+)\] (.+)`).FindSubmatch(row)[1:] {
							w.Write([]byte(`<td class="pe-4">`))
							w.Write(m)
							w.Write([]byte(`</td>`))
						}
						w.Write([]byte(`</tr>`))
					}
				}
				w.Write(logHtml[idx+4:])
			}
		}
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, req *http.Request) {
		var URL string
		if err := readJSON(req, &URL); err != nil {
			writeJSON(w, err.Error())
		} else if c, err := com.NewCrawler(URL, "", ".", 3); err != nil {
			writeJSON(w, err.Error())
		} else {
			writeJSON(w, c.Chapters)
			crawler.Store(c)
		}
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, req *http.Request) {
		var ID int
		if err := readJSON(req, &ID); err != nil {
			writeJSON(w, err.Error())
		} else {
			tasks.Delete(ID)
			writeJSON(w, nil)
		}
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, req *http.Request) {
		var r struct {
			Indexes []int
			Output  string
		}
		if err := readJSON(req, &r); err != nil {
			writeJSON(w, err.Error())
			return
		}

		c := *crawler.Load().(*com.Crawler)
		if len(r.Output) != 0 {
			c.Output = r.Output
		}
		for _, idx := range r.Indexes {
			go func(idx int, c *com.Crawler) {
				t := Task{
					ID:      <-id,
					Manga:   c.Title,
					Chapter: c.Chapters[idx].Title,
					mutex:   &sync.Mutex{},
				}
				tc <- t
				tasks.Store(t.ID, &t)
				prg, errs, done := c.FetchChapter(idx)
				for {
					select {
					case s := <-prg:
						t.mutex.Lock()
						t.Progress = s
						tc <- t
						t.mutex.Unlock()
					case err := <-errs:
						t.mutex.Lock()
						t.Errors = append(t.Errors, err)
						tc <- t
						t.mutex.Unlock()
					case <-done:
						t.mutex.Lock()
						t.Done = true
						tc <- t
						t.mutex.Unlock()
						return
					}
				}
			}(idx, &c)
		}
		writeJSON(w, nil)
	})

	http.HandleFunc("/downloading", func(w http.ResponseWriter, req *http.Request) {
		c, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer c.Close()
		tasks.Range(func(key, value interface{}) bool {
			v := value.(*Task)
			v.mutex.Lock()
			t := *v
			v.mutex.Unlock()
			if err := c.WriteJSON(t); err != nil {
				log.Println(err)
			}
			return true
		})
		for {
			if err := c.WriteJSON(<-tc); err != nil {
				log.Println(err)
			}
		}
	})

	http.HandleFunc("/config", func(w http.ResponseWriter, req *http.Request) {
		if len(configFile) == 0 {
			writeJSON(w, nil)
		} else if req.Method == "GET" {
			if c, err := com.NewConfig(configFile); err != nil {
				writeJSON(w, err.Error())
			} else {
				writeJSON(w, c)
			}
		} else {
			var u struct {
				Remove    int    `json:"remove"`
				Frequency int    `json:"frequency_in_hour"`
				Output    string `json:"output"`
				com.Asset
			}
			if err := readJSON(req, &u); err != nil {
				writeJSON(w, err.Error())
			} else if c, err := com.NewConfig(configFile); err != nil {
				writeJSON(w, err.Error())
			} else {
				if u.Remove > 0 {
					c.Assets = append(c.Assets[:u.Remove-1], c.Assets[u.Remove:]...)
				} else if len(u.URL) > 0 {
					c.Assets = append([]com.Asset{u.Asset}, c.Assets...)
				} else {
					c.Output = u.Output
					c.Frequency = u.Frequency
				}
				if err = c.Save(); err != nil {
					writeJSON(w, err.Error())
				} else {
					writeJSON(w, nil)
				}
			}
		}
	})

	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

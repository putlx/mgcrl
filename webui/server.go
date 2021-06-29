package webui

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/putlx/mgcrl/com"
	"github.com/putlx/mgcrl/util"
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
	ID      uint32       `json:"id"`
	Manga   string       `json:"manga"`
	Chapter string       `json:"chapter"`
	Errors  []*com.Error `json:"errors"`
	Done    bool         `json:"done"`
	mutex   *sync.Mutex
	com.Progress
}

func Serve(port int, output, csv, logFile string, maxRetry, frequency uint, log *log.Logger) {
	output, err := filepath.Abs(output)
	if err != nil {
		log.Fatalln(err)
	}
	var update = make(chan Task)
	var id uint32
	var crawler atomic.Value
	var tasks = &sync.Map{}
	var upgrader = websocket.Upgrader{}

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
		w.Write(bytes.Replace(html, []byte("{{output}}"), []byte(output), 1))
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
		if req.Method == "POST" {
			if len(logFile) != 0 {
				if _, err := os.Stat(logFile); err == nil {
					if err = os.Remove(logFile); err != nil {
						log.Println(err)
					}
				}
			}
			return
		}

		idx := bytes.Index(logHtml, []byte("{{}}"))
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
				for i := len(rows) - 1; i >= 0; i-- {
					if len(rows[i]) > 0 {
						w.Write([]byte(`<tr>`))
						for _, m := range regexp.MustCompile(`(\d{4}/\d{2}/\d{2}) (\d{2}:\d{2}:\d{2}) \[(\w+)\] (.+)`).FindSubmatch(rows[i])[1:] {
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
		} else if c, err := com.NewCrawler(URL, "", output, maxRetry); err != nil {
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
					ID:      atomic.AddUint32(&id, 1),
					Manga:   c.Title,
					Chapter: c.Chapters[idx].Title,
					mutex:   &sync.Mutex{},
				}
				update <- t
				tasks.Store(t.ID, &t)
				prg, errs, done := c.FetchChapter(idx)
				for {
					select {
					case s := <-prg:
						t.mutex.Lock()
						t.Progress = s
						update <- t
						t.mutex.Unlock()
					case err := <-errs:
						t.mutex.Lock()
						t.Errors = append(t.Errors, err)
						update <- t
						t.mutex.Unlock()
					case <-done:
						t.mutex.Lock()
						t.Done = true
						update <- t
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
			if err := c.WriteJSON(<-update); err != nil {
				log.Println(err)
			}
		}
	})

	http.HandleFunc("/records", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			if len(csv) == 0 {
				writeJSON(w, nil)
			} else {
				if records, err := util.ReadCSV(csv); err != nil {
					writeJSON(w, err.Error())
				} else {
					writeJSON(w, records)
				}
			}
		} else {
			var r struct {
				Remove int
				Record []string
			}
			if err := readJSON(req, &r); err != nil {
				writeJSON(w, err.Error())
			} else if records, err := util.ReadCSV(csv); err != nil {
				writeJSON(w, err.Error())
			} else {
				if len(r.Record) > 0 {
					records = append([][]string{r.Record}, records...)
				} else {
					records = append(records[:r.Remove], records[r.Remove+1:]...)
				}
				if err = util.WriteCSV(csv, records); err != nil {
					writeJSON(w, err.Error())
				} else {
					writeJSON(w, nil)
				}
			}
		}
	})

	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

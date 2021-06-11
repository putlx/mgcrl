package webui

import (
	"bytes"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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

func Serve(port int, db *sql.DB) {
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

	log := func(m string) {
		db.Exec("INSERT INTO sv_log VALUES (?, ?)", time.Now().Unix(), m)
	}

	writeJSON := func(w http.ResponseWriter, v interface{}) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if data, err := json.Marshal(v); err != nil {
			panic(err)
		} else if _, err = w.Write(data); err != nil {
			log(err.Error())
		}
	}

	readJSON := func(req *http.Request, v interface{}) error {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, v)
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
		if req.Method == "POST" {
			if _, err := db.Exec("DELETE FROM sv_log"); err != nil {
				log(err.Error())
			}
			return
		}

		idx := bytes.Index(logHtml, []byte("{{}}"))
		rows, err := db.Query("SELECT time, 'server', message FROM sv_log" +
			" UNION SELECT time, 'autocrawl', message FROM ac_log ORDER BY time DESC")
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(logHtml[:idx])
		for rows.Next() {
			var timestamp int64
			var from, message string
			if err = rows.Scan(&timestamp, &from, &message); err != nil {
				panic(err)
			}
			dt := strings.Split(strings.Split(time.Unix(timestamp, 0).String(), ".")[0], " ")
			w.Write([]byte(`<tr><td class="pe-4">`))
			w.Write([]byte(dt[0]))
			w.Write([]byte(`</td><td class="pe-4">`))
			w.Write([]byte(dt[1]))
			w.Write([]byte(`</td><td class="pe-4">`))
			w.Write([]byte(from))
			w.Write([]byte(`</td><td class="pe-4">`))
			w.Write([]byte(message))
			w.Write([]byte(`</td></tr>`))
		}
		w.Write(logHtml[idx+4:])
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, req *http.Request) {
		var URL string
		if err := readJSON(req, &URL); err != nil {
			writeJSON(w, err.Error())
		} else if c, err := com.NewCrawler(URL, "", ".", 3); err != nil {
			writeJSON(w, err.Error())
		} else {
			crawler.Store(c)
			writeJSON(w, c.Chapters)
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
			log(err.Error())
			return
		}
		defer c.Close()
		tasks.Range(func(key, value interface{}) bool {
			v := value.(*Task)
			v.mutex.Lock()
			err := c.WriteJSON(v)
			v.mutex.Unlock()
			if err != nil {
				log(err.Error())
			}
			return true
		})
		for {
			if err := c.WriteJSON(<-tc); err != nil {
				log(err.Error())
			}
		}
	})

	http.HandleFunc("/config", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			type Asset struct {
				Name        string `json:"name"`
				URL         string `json:"url"`
				Version     string `json:"version"`
				LastChapter string `json:"last_chapter"`
			}
			var config struct {
				Output string  `json:"output"`
				Freq   int     `json:"freq"`
				Assets []Asset `json:"assets"`
			}
			var attr string
			rows, err := db.Query("SELECT * FROM config WHERE attr = 'output'")
			if err != nil {
				panic(err)
			}
			rows.Next()
			if err = rows.Scan(&attr, &config.Output); err != nil {
				panic(err)
			}
			rows.Close()
			rows, err = db.Query("SELECT * FROM config WHERE attr = 'freq_in_hour'")
			if err != nil {
				panic(err)
			}
			rows.Next()
			if err = rows.Scan(&attr, &config.Freq); err != nil {
				panic(err)
			}
			rows.Close()
			rows, err = db.Query("SELECT * FROM assets")
			if err != nil {
				panic(err)
			}
			for rows.Next() {
				var a Asset
				if err = rows.Scan(&a.Name, &a.URL, &a.Version, &a.LastChapter); err != nil {
					panic(err)
				}
				config.Assets = append(config.Assets, a)
			}
			rows.Close()
			writeJSON(w, config)
		} else {
			var r struct {
				Remove      string `json:"remove"`
				Freq        int    `json:"freq"`
				Output      string `json:"output"`
				Name        string `json:"name"`
				URL         string `json:"url"`
				Version     string `json:"version"`
				LastChapter string `json:"last_chapter"`
			}
			if err := readJSON(req, &r); err != nil {
				writeJSON(w, err.Error())
			} else {
				var err error
				if len(r.Remove) > 0 {
					_, err = db.Exec("DELETE FROM assets WHERE name = ?", r.Remove)
				} else if len(r.Name) > 0 {
					_, err = db.Exec("INSERT INTO assets VALUES (?, ?, ?, ?)", r.Name, r.URL, r.Version, r.LastChapter)
				} else if len(r.Output) > 0 {
					_, err = db.Exec("UPDATE config SET val = ? WHERE attr = 'output'", r.Output)
				} else if r.Freq > 0 {
					_, err = db.Exec("UPDATE config SET val = ? WHERE attr = 'freq_in_hour'", r.Freq)
				}
				if err == nil {
					writeJSON(w, nil)
				} else {
					writeJSON(w, err.Error())
				}
			}
		}
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log(err.Error())
	}
}

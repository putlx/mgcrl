package webui

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/putlx/mgcrl/com"
)

//go:embed index.html
var html string

//go:embed index.js
var js string

//go:embed style.css
var css string

type Task struct {
	ID      int          `json:"id"`
	Manga   string       `json:"manga"`
	Chapter string       `json:"chapter"`
	Errors  []*com.Error `json:"errors"`
	Done    bool         `json:"done"`
	mutex   *sync.Mutex
	com.Progress
}

var tasks = &sync.Map{}
var upgrader = websocket.Upgrader{}

func WriteJSON(w http.ResponseWriter, v interface{}, lg *log.Logger) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if data, err := json.Marshal(v); err != nil {
		lg.Fatalln(err)
	} else if _, err = io.WriteString(w, string(data)); err != nil {
		lg.Println(err)
	}
}

func ReadJSON(req *http.Request, v interface{}) error {
	if data, err := io.ReadAll(req.Body); err != nil {
		return err
	} else if err = json.Unmarshal(data, v); err != nil {
		return err
	}
	return nil
}

func Serve(port int, w io.Writer) {
	var id = make(chan int)
	var tc = make(chan Task)
	var crawler atomic.Value
	go func() {
		for i := 0; ; i++ {
			id <- i
		}
	}()

	lg := log.New(w, "[webui] ", log.LstdFlags|log.Lmsgprefix)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, html)
	})

	http.HandleFunc("/index.js", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		io.WriteString(w, fmt.Sprintf("const port = %d;\n\n%s", port, js))
	})

	http.HandleFunc("/style.css", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		io.WriteString(w, css)
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, req *http.Request) {
		var URL string
		if err := ReadJSON(req, &URL); err != nil {
			WriteJSON(w, err.Error(), lg)
		} else if c, err := com.NewCrawler(URL, "", ".", 3); err != nil {
			WriteJSON(w, err.Error(), lg)
		} else {
			WriteJSON(w, c.Chapters, lg)
			crawler.Store(c)
		}
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, req *http.Request) {
		var ID int
		if err := ReadJSON(req, &ID); err != nil {
			WriteJSON(w, err.Error(), lg)
		} else {
			tasks.Delete(ID)
			WriteJSON(w, nil, lg)
		}
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, req *http.Request) {
		var r struct {
			Indexes []int
			Output  string
		}
		if err := ReadJSON(req, &r); err != nil {
			WriteJSON(w, err.Error(), lg)
		} else {
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
			WriteJSON(w, nil, lg)
		}
	})

	http.HandleFunc("/downloading", func(w http.ResponseWriter, req *http.Request) {
		c, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			lg.Fatalln(err)
		}
		defer c.Close()
		tasks.Range(func(key, value interface{}) bool {
			v := value.(*Task)
			v.mutex.Lock()
			t := *v
			v.mutex.Unlock()
			if err := c.WriteJSON(t); err != nil {
				lg.Println(err)
			}
			return true
		})
		for {
			if err := c.WriteJSON(<-tc); err != nil {
				lg.Println(err)
			}
		}
	})

	lg.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

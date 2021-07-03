package webui

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

//go:embed toast.js
var toast []byte

//go:embed favicon.ico
var favicon []byte

//go:embed reader.html
var reader []byte

//go:embed autocrawl.html
var autocrawl []byte

//go:embed log.html
var logHTML []byte

var logParser = regexp.MustCompile(`(\d{4}/\d{2}/\d{2}) (\d{2}:\d{2}:\d{2}) \[(\w+)\] (.+)`)

type Task struct {
	ID      uint32       `json:"id"`
	Manga   string       `json:"manga"`
	Chapter string       `json:"chapter"`
	Errors  []*com.Error `json:"errors"`
	Done    bool         `json:"done"`
	mutex   *sync.Mutex
	com.Progress
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	w.Write(data)
}

func readJSON(req *http.Request, v interface{}) error {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
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

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(bytes.Replace(html, []byte("{{output}}"), []byte(output), 1))
		case "PUT":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			var URL string
			if err := readJSON(req, &URL); err != nil {
				writeJSON(w, err.Error())
			} else if c, err := com.NewCrawler(URL, "", output, maxRetry); err != nil {
				writeJSON(w, err.Error())
			} else {
				writeJSON(w, c.Chapters)
				crawler.Store(c)
			}
		case "POST":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			var r struct {
				Indexes []int
				Output  string
			}
			if err := readJSON(req, &r); err != nil {
				w.Write([]byte(err.Error()))
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
		case "DELETE":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			var ID uint32
			if err := readJSON(req, &ID); err != nil {
				w.Write([]byte(err.Error()))
			} else {
				tasks.Delete(ID)
			}
		}
	})

	http.HandleFunc("/downloading", func(w http.ResponseWriter, req *http.Request) {
		c, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer c.Close()
		c.SetCloseHandler(func(_ int, _ string) error {
			crawler = atomic.Value{}
			return nil
		})
		go func() {
			for {
				if _, _, err := c.NextReader(); err != nil {
					c.Close()
					break
				}
			}
		}()
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
		switch req.Method {
		case "GET":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(reader)
		case "PUT":
			type File struct {
				Name  string `json:"name"`
				IsDir bool   `json:"is_dir"`
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			var path string
			if err := readJSON(req, &path); err != nil {
				writeJSON(w, err.Error())
				return
			} else if len(path) == 0 {
				path = output
			}
			f, err := os.Stat(path)
			if err != nil {
				writeJSON(w, err.Error())
				return
			}
			var resp struct {
				Path   string   `json:"path"`
				Files  []File   `json:"files"`
				Images []string `json:"images"`
			}
			if f.IsDir() {
				resp.Path, err = filepath.Abs(path)
			} else {
				resp.Path, err = filepath.Abs(path + "/..")
			}
			if err != nil {
				panic(err)
			}
			files, err := os.ReadDir(resp.Path)
			if err != nil {
				writeJSON(w, err.Error())
				return
			}
			resp.Files = append(resp.Files, File{"..", true})
			for _, file := range files {
				if file.IsDir() || file.Type().IsRegular() {
					resp.Files = append(resp.Files, File{file.Name(), file.IsDir()})
				}
				if f.IsDir() && file.Type().IsRegular() && util.IsImageFile(file.Name()) {
					if f, err := os.Open(resp.Path + "/" + file.Name()); err != nil {
						log.Println(err)
					} else if data, err := io.ReadAll(f); err != nil {
						log.Println(err)
						f.Close()
					} else {
						resp.Images = append(resp.Images, base64.StdEncoding.EncodeToString(data))
						f.Close()
					}
				}
			}
			if !f.IsDir() {
				r, err := zip.OpenReader(path)
				if err != nil {
					writeJSON(w, err.Error())
					return
				}
				defer r.Close()
				for _, f := range r.File {
					if util.IsImageFile(f.Name) {
						if rc, err := f.Open(); err != nil {
							log.Println(err)
						} else if data, err := io.ReadAll(rc); err != nil {
							log.Println(err)
							rc.Close()
						} else {
							resp.Images = append(resp.Images, base64.StdEncoding.EncodeToString(data))
							rc.Close()
						}
					}
				}
			}
			writeJSON(w, resp)
		case "DELETE":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			var path string
			var resp struct {
				Ok      bool   `json:"ok"`
				Message string `json:"message"`
			}
			if err := readJSON(req, &path); err != nil {
				resp.Message = err.Error()
			} else if err := os.RemoveAll(path); err != nil {
				resp.Message = err.Error()
			} else {
				for {
					path, err = filepath.Abs(path + "/..")
					if err != nil {
						break
					} else if err := os.Remove(path); err != nil {
						break
					}
				}
				resp.Ok = true
				resp.Message = path
			}
			writeJSON(w, resp)
		}
	})

	http.HandleFunc("/autocrawl", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(autocrawl)
		case "PUT":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			if len(csv) == 0 {
				writeJSON(w, nil)
			} else if records, err := util.ReadCSV(csv); err != nil {
				writeJSON(w, err.Error())
			} else {
				writeJSON(w, records)
			}
		case "POST":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			var record []string
			if err := readJSON(req, &record); err != nil {
				w.Write([]byte(err.Error()))
			} else if records, err := util.ReadCSV(csv); err != nil {
				w.Write([]byte(err.Error()))
			} else {
				records = append([][]string{record}, records...)
				if err := util.WriteCSV(csv, records); err != nil {
					w.Write([]byte(err.Error()))
				}
			}
		case "DELETE":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			var remove int
			if err := readJSON(req, &remove); err != nil {
				w.Write([]byte(err.Error()))
			} else if records, err := util.ReadCSV(csv); err != nil {
				w.Write([]byte(err.Error()))
			} else {
				records = append(records[:remove], records[remove+1:]...)
				if err := util.WriteCSV(csv, records); err != nil {
					w.Write([]byte(err.Error()))
				}
			}
		}
	})

	http.HandleFunc("/log", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(logHTML)
		case "DELETE":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			if _, err := os.Stat(logFile); err == nil {
				if err := os.Remove(logFile); err != nil {
					w.Write([]byte(err.Error()))
				}
			} else if !os.IsNotExist(err) {
				w.Write([]byte(err.Error()))
			}
		case "PUT":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			var logs = make([][]string, 0)
			if len(logFile) == 0 {
				writeJSON(w, nil)
			} else if f, err := os.Open(logFile); err == nil {
				defer f.Close()
				if data, err := io.ReadAll(f); err != nil {
					writeJSON(w, err.Error())
				} else {
					rows := strings.Split(string(data), "\n")
					for i := len(rows) - 1; i >= 0; i-- {
						if len(rows[i]) > 0 {
							logs = append(logs, logParser.FindStringSubmatch(rows[i])[1:])
						}
					}
					writeJSON(w, logs)
				}
			} else if os.IsNotExist(err) {
				writeJSON(w, logs)
			} else {
				writeJSON(w, err.Error())
			}
		}
	})

	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

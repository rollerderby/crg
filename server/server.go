package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/go-fsnotify/fsnotify"
	"github.com/rollerderby/crg/leagues"
	"github.com/rollerderby/crg/scoreboard"
	"github.com/rollerderby/crg/statemanager"
	"github.com/rollerderby/crg/websocket"
)

var urls []string

func printStartup(port uint16) {
	log.Print("")
	log.Printf("CRG Scoreboard and Game System Version %v", version)
	log.Print("")
	log.Print("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	log.Print("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	log.Print("Double-click/open the 'start.html' file, or")
	log.Print("Open a web browser (either Google Chrome or Mozilla Firefox recommended) to:")
	log.Printf("http://localhost:%d/", port)
	log.Print("or try one of these URLs:")
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print("Cannot get interfaces:", err)
	} else {
		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err != nil {
				log.Printf("Cannot get addresses on %v: %v", i, err)
			} else {
				for _, addr := range addrs {
					var ip net.IP
					switch v := addr.(type) {
					case *net.IPNet:
						ip = v.IP
					case *net.IPAddr:
						ip = v.IP
					}

					if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
						continue
					}
					var url string
					if ip.To4() != nil {
						url = fmt.Sprintf("http://%v:%d/", ip, port)
					} else {
						url = fmt.Sprintf("http://[%v]:%d/", ip, port)
					}
					urls = append(urls, url)
					log.Print(url)
				}
			}
		}
	}
	log.Print("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
	log.Print("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
}

func setDefaultHeaders(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("%s: %s %s %s", r.Host, r.RemoteAddr, r.Method, r.URL)
		w.Header().Set("cache-control", "private, max-age=0, no-cache")
		handler.ServeHTTP(w, r)
	})
}

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(version))
}

func urlsHandler(w http.ResponseWriter, _ *http.Request) {
	for _, url := range urls {
		w.Write([]byte(url))
		w.Write([]byte("\n"))
	}
}

func addDirWatcher(path string) (*fsnotify.Watcher, error) {
	mediaType := filepath.Base(path)
	fullpath := filepath.Join(statemanager.BaseFilePath(), path)
	os.MkdirAll(fullpath, 0775)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = watcher.Add(fullpath)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	f, err := os.Open(fullpath)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	for _, name := range names {
		short := filepath.Base(name)
		full := filepath.Join(path, short)

		statemanager.StateUpdate(fmt.Sprintf("Media.Type(%v).File(%v)", mediaType, short), full)
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				short := filepath.Base(event.Name)
				full := filepath.Join(path, short)

				if event.Op&fsnotify.Create == fsnotify.Create {
					statemanager.StateUpdate(fmt.Sprintf("Media.Type(%v).File(%v)", mediaType, short), full)
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					statemanager.StateUpdate(fmt.Sprintf("Media.Type(%v).File(%v)", mediaType, short), nil)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					statemanager.StateUpdate(fmt.Sprintf("Media.Type(%v).File(%v)", mediaType, short), nil)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	return watcher, nil
}

func setSettings(k, v string) error {
	return statemanager.StateUpdate(k, v)
}

func openLog() *os.File {
	path := filepath.Join(statemanager.BaseFilePath(), "logs", fmt.Sprintf("scoreboard-%v.log", time.Now().Format(time.RFC3339)))
	os.MkdirAll(filepath.Dir(path), 0775)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		log.Printf("Error opening file %v: %v", path, err)
		return nil
	}

	log.SetOutput(io.MultiWriter(f, os.Stdout))
	return f
}

// Start initalizes all scoreboard subsystems and starts up a webserver on port
func Start(port uint16) {
	l := openLog()
	if l != nil {
		defer l.Close()
	}
	mux := http.NewServeMux()
	var savers []*statemanager.Saver

	// Initialize statemanager and load Settings.*
	statemanager.Initialize()
	statemanager.Lock()
	statemanager.RegisterPatternUpdaterString("Settings", 0, setSettings)
	statemanager.Unlock()
	savers = append(savers, statemanager.NewSaver("config/settings", "Settings", time.Duration(5)*time.Second, true, true))

	// Initialize leagues and load Leagues.*
	leagues.Initialize()
	savers = append(savers, statemanager.NewSaver("config/leagues", "Leagues", time.Duration(5)*time.Second, true, true))

	// Initialize scoreboard and load ScoreBoard.*
	statemanager.Lock()
	scoreboard.New()
	statemanager.Unlock()
	savers = append(savers, statemanager.NewSaver("config/scoreboard", "ScoreBoard", time.Duration(5)*time.Second, true, true))

	// Initialize websocket interface
	websocket.Initialize(mux)

	addDirWatcher("html/images/teamlogo")
	addDirWatcher("html/images/sponsor_banner")
	addDirWatcher("html/images/fullscreen")

	printStartup(port)
	mux.Handle("/", http.FileServer(http.Dir(filepath.Join(statemanager.BaseFilePath(), "html"))))

	c := make(chan os.Signal, 1)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", port), setDefaultHeaders(mux))
		if err != nil {
			log.Print(err)
		}
		c <- os.Kill
	}()

	mux.Handle("/version", http.HandlerFunc(versionHandler))
	mux.Handle("/urls", http.HandlerFunc(urlsHandler))

	signal.Notify(c, os.Interrupt, os.Kill)
	s := <-c
	log.Printf("Server received signal: %v.  Shutting down", s)

	for _, saver := range savers {
		saver.Close()
	}
}

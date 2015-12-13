// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

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

	"github.com/rollerderby/crg/leagues"
	"github.com/rollerderby/crg/scoreboard"
	"github.com/rollerderby/crg/state"
	"github.com/rollerderby/crg/utils"
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

func openLog() *os.File {
	path := utils.Path("logs", fmt.Sprintf("scoreboard-%v.log", time.Now().Format(time.RFC3339)))
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
	var savers []*state.Saver

	// Initialize state and load Settings.*
	state.Initialize()
	savers = append(savers, initSettings("config/settings"))

	// Initialize leagues and load Leagues.*
	leagues.Initialize()
	savers = append(savers, state.NewSaver("config/leagues", "Leagues", time.Duration(5)*time.Second, true, true))

	// Initialize scoreboard and load Scoreboard.*
	state.Lock()
	scoreboard.New()
	state.Unlock()
	savers = append(savers, state.NewSaver("config/scoreboard", "Scoreboard", time.Duration(5)*time.Second, true, true))

	// Initialize websocket interface
	websocket.Initialize(mux)

	addFileWatcher("TeamLogos", "html", "/images/teamlogo")
	addFileWatcher("Sponsors", "html", "/images/sponsor_banner")
	addFileWatcher("Image", "html", "/images/fullscreen")
	addFileWatcher("Video", "html", "/videos")
	addFileWatcher("CustomHtml", "html", "/customhtml")

	printStartup(port)
	mux.Handle("/", http.FileServer(http.Dir(utils.Path("html"))))

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

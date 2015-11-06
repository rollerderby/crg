package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rollerderby/crg/scoreboard"
	"github.com/rollerderby/crg/statemanager"
	"github.com/rollerderby/crg/websocket"
)

func printStartup(port uint16) {
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
					if ip.To4() != nil {
						log.Printf("http://%v:%d/", ip, port)
					} else {
						log.Printf("http://[%v]:%d/", ip, port)
					}
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

// Start initalizes all scoreboard subsystems and starts up a webserver on port
func Start(port uint16) {
	mux := http.NewServeMux()
	statemanager.Initialize()

	scoreboard.New()
	websocket.Initialize(mux)

	// filename, base string, interval time.Duration, version bool
	saver, savedState := statemanager.NewSaver("state.json", "ScoreBoard", time.Duration(5)*time.Second, true)
	statemanager.StateSetGroup(savedState)

	printStartup(port)
	mux.Handle("/", http.FileServer(http.Dir("html")))

	c := make(chan os.Signal, 1)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", port), setDefaultHeaders(mux))
		if err != nil {
			log.Print(err)
		}
		c <- os.Kill
	}()

	signal.Notify(c, os.Interrupt, os.Kill)
	s := <-c
	log.Printf("Server received signal: %v.  Shutting down", s)
	saver.Close()
}

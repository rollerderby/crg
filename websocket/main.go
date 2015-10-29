package websocket

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	ws "github.com/gorilla/websocket"
	"github.com/rollerderby/crg/statemanager"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type connection struct {
	sync.Mutex
	conn     *ws.Conn
	paths    []string
	ch       chan map[string]*string
	state    map[string]*string
	listener *statemanager.Listener
}

func newConnection(conn *ws.Conn) *connection {
	c := &connection{
		conn: conn,
		ch:   make(chan map[string]*string, 10),
	}
	c.listener = statemanager.NewListener(fmt.Sprintf("websocket(%v)", conn.RemoteAddr()), c.processUpdates)

	return c
}

func (c *connection) Close() {
	c.listener.Close()
	c.conn.Close()
}

func (c *connection) Run() {
	defer c.Close()

	for {
		var cmd command
		err := c.conn.ReadJSON(&cmd)
		if err != nil {
			log.Print("Cannot read command: ", err)
			return
		}

		switch cmd.Action {
		case "Register":
			c.listener.RegisterPaths(cmd.Data)
		default:
			// Try to send a command through the statemanager
			err := statemanager.Command(cmd.Action, cmd.Data)
			if err != nil {
				log.Print("Error processing command: ", err)
			}
			log.Printf("cmd: %+v  returned error: %v", cmd, err)
		}

	}
}

func (c *connection) requestUpdates(paths []string) {
	c.Lock()
	defer c.Unlock()

	c.paths = append(c.paths, paths...)
	for _, p1 := range paths {
		found := false
		for _, p2 := range c.paths {
			if p1 == p2 {
				found = true
				break
			}
		}
		if !found {
			c.paths = append(c.paths, p1)
		}
	}
}

func (c *connection) processUpdates(s map[string]*string) {
	c.Lock()
	defer c.Unlock()

	err := c.conn.WriteJSON(state{State: s})
	if err != nil {
		log.Print("Cannot send JSON to client: ", err)
		c.Close()
		return
	}
	return
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c := newConnection(conn)
	c.Run()
}

func Initialize(mux *http.ServeMux) {
	mux.HandleFunc("/ws", wsHandler)
}

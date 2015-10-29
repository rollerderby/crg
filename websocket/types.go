package websocket

type command struct {
	Action string   `json:"action"`
	Data   []string `json:"data"`
}

type state struct {
	State map[string]*string `json:"state"`
}

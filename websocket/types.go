package websocket

type command struct {
	Action    string            `json:"action"`
	Data      []string          `json:"data"`
	Field     string            `json:"field"`
	FieldData map[string]string `json:"fieldData"`
}

type state struct {
	State map[string]*string `json:"state"`
}

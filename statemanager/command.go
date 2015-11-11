package statemanager

import "errors"

// CommandFunc is a callback function when a command is triggered.
// Commands can be registered by RegisterCommand and unregistered
// by UnregisterCommand
type CommandFunc func([]string) error

// ErrCommandNotFound is returned by Command when the requested
// command is not currently registered with the system
var ErrCommandNotFound = errors.New("Command Not Found")
var ErrCommandArguments = errors.New("Incorrect Argument Count")
var commands = make(map[string]CommandFunc)

// Command requests the command registered with name be called
// and passed data as parameters.  Returns nil error on success,
// ErrCommandNotFound, or an error from the registered command
// function
func Command(name string, data []string) error {
	Lock()
	defer Unlock()

	if name == "Set" {
		if len(data) != 2 {
			return ErrCommandArguments
		}
		return StateSet(data[0], data[1])
	}

	c, ok := commands[name]
	if !ok {
		return ErrCommandNotFound
	}
	return c(data)
}

// RegisterCommand registers the CommandFunc with the command
// subsystem with name
func RegisterCommand(name string, c CommandFunc) {
	commands[name] = c
}

// UnregisterCommand removes the CommandFunc registered with
// name from the command subsystem
func UnregisterCommand(name string) {
	delete(commands, name)
}

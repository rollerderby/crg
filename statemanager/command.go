// Copyright 2015-2016 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package statemanager

import "errors"

// CommandFunc is a callback function when a command is triggered.
// Commands can be registered by RegisterCommand and unregistered
// by UnregisterCommand
type CommandFunc func([]string) error

var errCommandNotFound = errors.New("Command Not Found")
var errCommandArguments = errors.New("Incorrect Argument Count")
var commands = make(map[string]CommandFunc)

// Command requests the command registered with name be called
// and passed data as parameters.  Returns nil error on success,
// errCommandNotFound, errCommandArguments, or an error from the
// registered command function
func Command(name string, data []string) error {
	Lock()
	defer Unlock()

	if name == "Set" {
		if len(data) != 2 {
			return errCommandArguments
		}
		return StateSet(data[0], data[1])
	}

	c, ok := commands[name]
	if !ok {
		return errCommandNotFound
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

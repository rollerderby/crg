// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/rollerderby/crg/statemanager"
)

func setSettings(k, v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return statemanager.StateUpdate(k, nil)
	}
	return statemanager.StateUpdate(k, v)
}

// setup default settings
func initSettings(saveFile string) *statemanager.Saver {
	defaults := []struct{ name, value string }{
		{"BackgroundStyle", "bg_blacktowhite"},
		{"BoxStyle", "box_flat"},
		{"CurrentView", "scoreboard"},
		{"HideJamTotals", "false"},
		{"SwapTeams", "false"},
	}
	views := []string{"View", "Preview"}
	statemanager.Lock()
	for _, v := range views {
		for _, d := range defaults {
			statemanager.StateUpdate(fmt.Sprintf("Settings.%v.%v", v, d.name), d.value)
		}
	}
	statemanager.RegisterPatternUpdaterString("Settings", 0, setSettings)
	statemanager.Unlock()
	return statemanager.NewSaver(saveFile, "Settings", time.Duration(5)*time.Second, true, true)
}

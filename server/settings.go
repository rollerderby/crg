// Copyright 2015-2016 The CRG Authors (see AUTHORS file).
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
		return statemanager.StateDelete(k)
	}
	return statemanager.StateUpdateString(k, v)
}

// setup default settings
func initSettings(saveFile string) *statemanager.Saver {
	defaults := []struct{ name, value string }{
		{"BackgroundStyle", "bg_blacktowhite"},
		{"BoxStyle", "box_flat"},
		{"CurrentView", "scoreboard"},
		{"HideJamTotals", "false"},
		{"SwapTeams", "false"},
		{"Image", "/images/fullscreen/American Flag.jpg"},
		{"Video", "/videos/American Flag.webm"},
		{"CustomHtml", "/customhtml/example"},
	}
	views := []string{"View", "Preview"}
	statemanager.Lock()
	for _, v := range views {
		for _, d := range defaults {
			statemanager.StateUpdateString(fmt.Sprintf("Settings.%v.%v", v, d.name), d.value)
		}
	}
	statemanager.RegisterPatternUpdaterString("Settings", 0, setSettings)
	statemanager.Unlock()
	return statemanager.NewSaver(saveFile, "Settings", time.Duration(5)*time.Second, true, true)
}

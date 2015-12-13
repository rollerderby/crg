// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/rollerderby/crg/state"
)

func setSettings(k, v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return state.Delete(k)
	}
	state.SetStateString(k, v)
	return nil
}

// setup default settings
func initSettings(saveFile string) *state.Saver {
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
	state.Lock()
	for _, v := range views {
		for _, d := range defaults {
			state.SetStateString(fmt.Sprintf("Settings.%v.%v", v, d.name), d.value)
		}
	}
	state.RegisterPatternUpdaterString("Settings", 0, setSettings)
	state.Unlock()
	return state.NewSaver(saveFile, "Settings", time.Duration(5)*time.Second, true, true)
}

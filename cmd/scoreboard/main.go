// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package main

import (
	"flag"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/pprof"

	"github.com/kardianos/osext"
	"github.com/rollerderby/crg/server"
	"github.com/rollerderby/crg/statemanager"
)

var port int
var cpuProf bool

func init() {
	flag.IntVar(&port, "port", 8000, "Server Port")
	flag.BoolVar(&cpuProf, "cpuprof", false, "Dump CPU Profile")
}

func exists(dir bool, path ...string) bool {
	p := filepath.Join(path...)
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	if dir && fi.IsDir() {
		return true
	} else if !dir && !fi.IsDir() {
		return true
	}
	return false
}

func main() {
	path, err := osext.ExecutableFolder()
	if err == nil {
		if exists(true, path, "html") && exists(false, path, "html/index.html") {
			statemanager.SetBaseFilePath(path)
		} else if exists(true, path, "..", "html") && exists(false, path, "..", "html/index.html") {
			statemanager.SetBaseFilePath(path, "..")
		}
	}
	flag.Parse()

	if cpuProf {
		f, err := os.Create(filepath.Join(statemanager.BaseFilePath(), "profile.cpu"))
		if err != nil {
			log.Print(err)
		} else {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	}
	server.Start(uint16(port))
}

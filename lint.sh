#!/bin/bash

find . -type d -and -not -path "../../cmd/scoreboard/html/*" -exec golint {} \;
go vet `find . -name "*.go" -exec dirname {} \; | sort -u`

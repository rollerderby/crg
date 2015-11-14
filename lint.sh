#!/bin/bash

find . -type d -and -not -path "../../cmd/scoreboard/html/*" -exec golint {} \;

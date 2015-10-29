#!/bin/sh
GOOS=linux GOARCH=amd64 go build -v -o scoreboard-Linux-amd64
GOOS=linux GOARCH=x86 go build -v -o scoreboard-Linux-x86
GOOS=windows GOARCH=amd64 go build -v -o scoreboard-Windows-amd64.exe
GOOS=windows GOARCH=x86 go build -v -o scoreboard-Windows-x86.exe

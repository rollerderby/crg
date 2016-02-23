#!/bin/bash

build() {
	echo "Building for $1-$2 to $3"
	GOOS=$1 GOARCH=$2 go install -v std
	GOOS=$1 GOARCH=$2 go build -v -o crg-scoreboard_$VERSION/$3 ./cmd/scoreboard
	echo
}

ZIP=0
RELEASE=0

if [ "z$1" = "z-zip" ]; then ZIP=1; fi
if [ "z$1" = "z-release" ]; then RELEASE=1; ZIP=1; fi

VERSION=3.9.0
if [ $RELEASE -eq 0 ]; then VERSION=$VERSION-`date +%Y%m%d%H%M%S`; fi

echo Building Version $VERSION
echo

cat > server/version.go <<END
package server

const version = "$VERSION"
END

go get -u github.com/gorilla/websocket
go get -u github.com/satori/go.uuid
go get -u github.com/go-fsnotify/fsnotify
go get -u github.com/kardianos/osext
if [ $ZIP -eq 0 ]; then
	mkdir -p bin
	PLATFORM=$(uname -s)
	if [[ $PLATFORM == MINGW32_NT* ]] || [[ $PLATFORM == CYGWIN* ]]; then
		EXT=".exe"
	else
		EXT=""
	fi
	rm -f ./bin/scoreboard$EXT
	go build -v -o ./bin/scoreboard$EXT ./cmd/scoreboard
else
	rm -f scoreboard-*
	mkdir -p release
	rm -f release/scoreboard-$VERSION.zip

	mkdir -p crg-scoreboard_$VERSION
	cp -r html crg-scoreboard_$VERSION
	cp start.html AUTHORS LICENSE crg-scoreboard_$VERSION

	build "linux" "386" "scoreboard-Linux-32"
	build "linux" "amd64" "scoreboard-Linux-64"
	build "windows" "386" "scoreboard-Windows-32.exe"
	build "windows" "amd64" "scoreboard-Windows-64.exe"
	build "darwin" "386" "scoreboard-MacOS-32"
	build "darwin" "amd64" "scoreboard-MacOS-64"

	echo Zipping to release/crg-scoreboard_$VERSION.zip
	zip -qr release/crg-scoreboard_$VERSION.zip crg-scoreboard_$VERSION
	rm -rf crg-scoreboard_$VERSION
fi

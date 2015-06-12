#!/bin/bash
set -e
if [ ! -d bin ]; then
	mkdir bin
fi
export GOPATH=`pwd`
for l in \
	"github.com/craigmj/commander" \
	"github.com/craigmj/aptlastupdate" \
	; do
	if [ ! -d src/$l ]; then
		go get $l
	fi
done
go build -o bin/djaproxy src/cmd/djaproxy.go
if [ ! -e /usr/bin/djaproxy ]; then
	sudo ln -s `pwd`/bin/djaproxy /usr/bin
fi

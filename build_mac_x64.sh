#!/bin/sh
rm -fr ./mac-x64
mkdir -p ./mac-x64
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w"  -o ./mac-x64/icomplie

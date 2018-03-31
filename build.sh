#!/bin/bash

export GOPATH=$(pwd)

go build amspackager.go
go build ams.go

cp amspackager /usr/local/bin/
cp ams /usr/local/bin/


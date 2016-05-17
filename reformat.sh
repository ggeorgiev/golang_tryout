#!/bin/sh

BASEDIR=$(pwd)/$(dirname $0)
cd $BASEDIR

find src -name *.go | xargs gofmt -w -s
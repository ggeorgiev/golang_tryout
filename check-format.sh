find src -name *.go | xargs gofmt -l

if [ "`find src -name *.go | xargs gofmt -l | wc -l`" -ne "0" ]; then exit 1; fi

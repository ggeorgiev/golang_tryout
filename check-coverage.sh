if [ "`go test --cover pack | awk '{ print $5 }'`" \< "80.0%" ]; then exit 1; fi

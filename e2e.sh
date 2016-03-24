#!/bin/bash

echo "> Should get non-zero exit code if config is missing"
./owl > /dev/null 2>&1
if [ "$?" = "0" ]; then
    echo "Got $?"
    exit -1
fi

echo "> Files should be formatted"
gofiles=$(git ls-files | grep -v Godeps | grep '.go$')
[ -z "$gofiles" ] || unformatted=$(gofmt -l $gofiles)
if [ ! -z "$unformatted" ]; then
    echo >&2 "Go files must be formatted with gofmt. Please run:"
    for fn in $unformatted; do
        echo >&2 "  gofmt -s -w $PWD/$fn"
    done
    exit -1
fi

echo "All tests passed"

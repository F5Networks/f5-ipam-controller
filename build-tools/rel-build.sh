#!/bin/bash

set -ex

go mod download

CGO_ENABLED=1
GOOS=linux
GOARCH=amd64
go build -v -ldflags "-extldflags \"-static\" -X main.version=${BUILD_VERSION} -X main.buildInfo=${BUILD_INFO}" -o /bin/f5-ipam-controller $REPOPATH/cmd/f5-ipam-controller

RUN_TESTS=${RUN_TESTS:-1}
. $REPOPATH/build-tools/_build-lib.sh

if [ $RUN_TESTS -eq 1 ]; then
	  go get github.com/onsi/ginkgo/ginkgo@v1.16.2
	  go get github.com/onsi/gomega@v1.12.0
	  go get github.com/wadey/gocovmerge
    go get golang.org/x/tools/cmd/cover
    go get github.com/mattn/goveralls
	  GO111MODULE=off
    echo "Gathering unit test code coverage for 'release' build..."
    ginkgo_test_with_coverage
    # push coverage data to coveralls if F5 repo or if configured for fork.
    if [ "$COVERALLS_TOKEN" ]; then
      echo "Pushing coverage data to coveralls"
      goveralls -coverprofile=./coverage.out -service=azure
    fi
fi


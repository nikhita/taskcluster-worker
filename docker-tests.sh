#!/bin/bash

touch .bash_history
ARGS="--tty --interactive --rm --privileged -e DEBUG -v `pwd`/.bash_history:/root/.bash_history -v `pwd`:/src taskcluster/tc-worker-env";

if [[ "$@" == go\ * ]]; then
  docker run $ARGS $@;
elif [[ "$1" == -- ]]; then
  shift;
  docker run $ARGS $@;
elif [[ "$@" == bash ]]; then
  docker run $ARGS bash --login;
elif [[ "$@" == goconvey ]]; then
  docker run -p 8080:8080 $ARGS goconvey -packages 1 -launchBrowser=false --host 0.0.0.0 -port 8080;
elif [[ "$@" == "" ]]; then
  docker run $ARGS go test -race -tags 'qemu network' -p 1 -v \
  `go list ./... | grep -v ^github.com/taskcluster/taskcluster-worker/vendor/`;
else
  docker run $ARGS go test -v -race -tags 'qemu network' -p 1 $@;
fi;

if [[ "$?" != "0" ]]; then
  echo "### TEST FAILED";
  exit 1;
else
  echo "### TEST PASSED";
fi

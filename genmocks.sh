#!/bin/bash
#
# Generate mocks for go tests using mockery
#
# To install mockery:
#   go get -u github.com/vektra/mockery

set -e

CURDIR=$(readlink -f $(dirname $0))
REPO=github.com/percona/mongodb-orchestration-tools

for SUBPATH in "executor" "executor/job" "executor/metrics" "internal/api" "internal/pod" "watchdog/watcher"; do
	pushd $CURDIR/$SUBPATH
		$GOPATH/bin/mockery -all
		for MOCK in mocks/*.go; do
			sed -i -e s@"${CURDIR}/${SUBPATH}/"@"${REPO}/${SUBPATH}"@g $MOCK
		done
	popd
done

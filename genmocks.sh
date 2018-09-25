#!/bin/bash
#
# Generate mocks for go tests using mockery
#
# To install mockery:
#   go get -u github.com/vektra/mockery

set -e

CURDIR=$(readlink -f $(dirname $0))
REPO=github.com/percona/dcos-mongo-tools

for SUBPATH in "internal/api" "internal/pod"; do
	pushd $CURDIR/$SUBPATH
		mockery -all
		for mock in $CURDIR/$SUBPATH/mocks/*.go; do
			sed -i -e s@"${CURDIR}/${SUBPATH}/"@"${REPO}/${SUBPATH}"@g $mock
		done
	popd
done

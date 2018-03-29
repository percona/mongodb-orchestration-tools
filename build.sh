#!/usr/bin/env bash

set -e
set -x

if [ $# -ne 1 ] ; then
    printf "Did not pass platform. Usage: $0 [linux|darwin|windows]\n"
    exit 1
fi

EXE_FILE_DIR=bin
PLATFORM=$1
echo "Building dcos-mongo-tools for $PLATFORM"

if [ -z "$GOPATH" -o -z "$(which go)" ]; then
    echo "Missing GOPATH environment variable or 'go' executable. Please configure a Go build environment."
    syntax
    exit 1
fi

REPO_ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_NAME=$(basename $REPO_ROOT_DIR) # default to name of REPO_ROOT_DIR

if [ -z "$REPO_ROOT_DIR" -o -z "$REPO_NAME" ]; then
    echo "Missing REPO_ROOT_DIR or REPO_NAME environment variables."
    syntax
    exit 1
fi

# Detect Go version to determine if the user has a compatible Go version or not.
GO_VERSION=$(go version | awk '{print $3}')
# Note, UPX only works on binaries produced by Go 1.7+. However, we require Go 1.8+
UPX_BINARY="$(which upx || which upx-ucl || echo '')"
# For dev iteration; upx takes a long time; can set env var
if [ -n "$CLI_BUILD_SKIP_UPX" ]; then
    UPX_BINARY=
fi
case "$GO_VERSION" in
    go1.[8-9]*|go1.1[0-9]*|go[2-9]*) # go1.8+, go2+ (must come before go1.0-go1.7: support e.g. go1.10)
        ;;
    go0.*|go1.[0-7]*) # go0.*, go1.0-go1.7
        echo "Detected Go <=1.7. This is too old, please install Go 1.8+: $(which go) $GO_VERSION"
        exit 1
        ;;
    *) # ???
        echo "Unrecognized go version: $(which go) $GO_VERSION"
        exit 1
        ;;
esac

GOPATH_REPO_ORG=${ORG_PATH:=github.com/percona}
GOPATH_REPO_ORG_DIR="$GOPATH/src/$GOPATH_REPO_ORG"
GOPATH_EXE_DIR="$GOPATH_REPO_ORG_DIR/$REPO_NAME"

# Add symlink from GOPATH which points into the repository directory, if necessary:
SYMLINK_LOCATION="$GOPATH_REPO_ORG_DIR/$REPO_NAME"
if [ ! -h "$SYMLINK_LOCATION" -o "$(readlink $SYMLINK_LOCATION)" != "$REPO_ROOT_DIR" ] && [ ! -d "$SYMLINK_LOCATION" -o "$SYMLINK_LOCATION" != "$REPO_ROOT_DIR" ]; then
    echo "Creating symlink from GOPATH=$SYMLINK_LOCATION to REPOPATH=$REPO_ROOT_DIR"
    rm -rf "$SYMLINK_LOCATION"
    mkdir -p "$GOPATH_REPO_ORG_DIR"
    cd $GOPATH_REPO_ORG_DIR
    ln -s "$REPO_ROOT_DIR" $REPO_NAME
fi

# Run 'make'/'go test' from within GOPATH:
cd $GOPATH_EXE_DIR

# run unit tests
make test

# build
make PLATFORM=$PLATFORM 

if [ -n "$UPX_BINARY" -a "$PLATFORM" != "darwin" ]; then
    for EXE_FILENAME in $EXE_FILE_DIR/*; do
        if [ "$($UPX_BINARY -t $EXE_FILENAME 2>&1 | grep -c 'not packed')" -gt 0 ]; then
            echo "Packing $EXE_FILENAME with upx"
            $UPX_BINARY -q --best $EXE_FILENAME
        else
            echo "Already packed $EXE_FILENAME by UPX, skipping"
        fi
    done
else
    echo "Skipping UPX compression"
fi

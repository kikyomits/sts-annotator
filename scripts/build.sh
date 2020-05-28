#!/usr/bin/env bash

# TODO: update way to build / deploy, use modern CICD

SCRIPT_DIR=$(dirname $0)
ROOT_DIR="$SCRIPT_DIR/.."

VERSION="2"

# Test go
echo "Start go test"
go test $ROOT_DIR

# Build go
echo "Start go build"
env GOOS=linux GOARCH=amd64 go build -o "$ROOT_DIR/build/sts-annotator" $ROOT_DIR/*.go

# Build Docker Image
echo "Start docker build"
docker build -t mkikyotani/sts-annotator:$VERSION $ROOT_DIR/build
docker push mkikyotani/sts-annotator:$VERSION

echo "Build Complete"

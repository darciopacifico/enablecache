#!/bin/bash
GOOS=darwin GOARCH=amd64 go install github.com/darciopacifico/enablecache/./...
GOOS=linux  GOARCH=amd64 go install github.com/darciopacifico/enablecache/./...


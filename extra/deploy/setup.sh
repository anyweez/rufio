#! /bin/bash

# Install golang protobuf compiler
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
protoc --plugin=bin/protoc-gen-go --go_out=. src/github.com/luke-segars/loldata/proto/*.proto

# Install dependencies
echo `go get -u ./...`

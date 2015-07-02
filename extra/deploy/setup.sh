#! /bin/bash

# Install golang protobuf compiler
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get github.com/tools/godep

# Install dependencies
$GOPATH/bin/godep restore


all:
	go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
	protoc --plugin=$(GOPATH)/bin/protoc-gen-go --go_out=. proto/*.proto
	go install github.com/luke-segars/rufio/fetcher
	go install github.com/luke-segars/rufio/taskqueue

clean:
	rm bin/*

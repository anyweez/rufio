
all:
	protoc --plugin=$(GOPATH)/bin/protoc-gen-go --go_out=. proto/*.proto
	# Raw data binaries
	go install github.com/luke-segars/loldata/raw/raw_games
	go install github.com/luke-segars/loldata/raw/raw_leagues
	go install github.com/luke-segars/loldata/raw/raw_summoners
	# Processed data binaries
	go install github.com/luke-segars/loldata/processed/processed_games
	go install github.com/luke-segars/loldata/processed/processed_leagues
	go install github.com/luke-segars/loldata/processed/processed_summoners
	# Task queue
	go install github.com/luke-segars/loldata/taskqueue

clean:
	rm raw_games raw_leagues processed_games
	rm input/*

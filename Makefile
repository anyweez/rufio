
all:
	mkdir -p src/proto
	protoc --plugin=bin/protoc-gen-go --go_out=src/ proto/*.proto
	go install raw/raw_games raw/raw_leagues processed/processed_games processed/processed_summoners taskqueue

inputs:
	./extract_inputs

clean:
	rm raw_games raw_leagues processed_games
	rm input/*

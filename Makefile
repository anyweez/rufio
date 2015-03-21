
all:
	mkdir -p src/proto
	protoc --plugin=bin/protoc-gen-go --go_out=src/ proto/*.proto
	go build raw/raw_games raw/raw_leagues processed/processed_games processed/processed_summoners
	./extract_inputs

clean:
	rm raw_games raw_leagues processed_games
	rm input/*

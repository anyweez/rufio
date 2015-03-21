package main

import (
	gproto "code.google.com/p/goprotobuf/proto"
	"errors"
	"flag"
	"log"
	proto "proto"
	shared "shared"
)

/**
 * This program reads in an arbitrary list of ID's and generates job requests for each of them. The job requests are
 * added to a Beanstalk queue at QUEUE_ADDRESS. Logs processors can pull jobs from the desired queues.
 */

var TARGET_LIST = flag.String("target_ids", "input/summoners", "The list of line-delimited id's that should be created as target id's few new tasks.")
var JOB_TYPE = flag.String("type", "", "Job type: {RETRIEVE_RECENT_GAMES, GENERATE_PROCESSED_GAME}.")
var QUEUE_ADDRESS = flag.String("addr", "localhost:11300", "[host:port] for the queue to populate.")

func main() {
	flag.Parse()

	// Read in all ID's.
	jobType := proto.ProcessedJobRequest_GENERATE_PROCESSED_GAME
	tubeName := ""

	ids, err := shared.LoadIds(*TARGET_LIST)
	if err != nil {
		log.Fatal(err.Error())
	}

	switch *JOB_TYPE {
	case "RETRIEVE_RECENT_GAMES":
		jobType = proto.ProcessedJobRequest_RETRIEVE_RECENT_GAMES
		tubeName = "retrieve_recent_games"
		break
	case "GENERATE_PROCESSED_GAME":
		jobType = proto.ProcessedJobRequest_GENERATE_PROCESSED_GAME
		tubeName = "generate_processed_game"
		break
	}

	if len(tubeName) == 0 {
		log.Fatal(errors.New("Invalid job type provided."))
		return
	}

	submitter, err := NewSubmitter(*QUEUE_ADDRESS, []string{tubeName})
	if err != nil {
		log.Fatal(err.Error())
	}

	// Generate a job proto for each ID and submit to the queue!
	for _, id := range ids {
		submitter.Submit(&proto.ProcessedJobRequest{
			Type:     jobType.Enum(),
			TargetId: gproto.Int64(int64(id)),
		}, tubeName)
	}

	submitter.Stats()
}

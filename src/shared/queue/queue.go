package queue

import (
	gproto "code.google.com/p/goprotobuf/proto"
	"github.com/kr/beanstalk"
	"proto"
)

/**
 * This function returns a channel that continuously pulls from the specified beanstalk queue
 * and returns jobs for a worker to work on.
 *
 * Closing the channel is safe and will stop listening to all tubes.
 * TODO: make closing the channel safe.
 */
func NewQueueListener(address string, tubes []string) (chan proto.ProcessedJobRequest, error) {
	conn, cerr := beanstalk.Dial("tcp", address)
	out := make(chan proto.ProcessedJobRequest)

	if cerr != nil {
		return out, cerr
	}

	// Create a new tube set and kick off a concurrent goroutine to continuously populate it.
	ts := beanstalk.NewTubeSet(conn, tubes)
	go harvestJobs(ts, out)

	return out
}

func harvestJobs(ts *beanstalk.TubeSet, out chan proto.ProcessedJobRequest) {
	defer ts.Conn.Close()

	for {
		id, body, err := ts.Reserve(0)

		if err != nil {
			close(out)
		}

		job := proto.ProcessedJobRequest{}

		gproto.Unmarshal(body, &job)
		job.JobId = id

		// Block until the current task is removed from the channel, then
		// pop another one on.
		out <- job
	}
}

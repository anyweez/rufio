package main

import (
	gproto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/kr/beanstalk"
	proto "proto"
	"time"
)

type Submitter struct {
	Tubes map[string]*beanstalk.Tube
	Conn  *beanstalk.Conn
}

func NewSubmitter(address string, tubes []string) (Submitter, error) {
	sub := Submitter{}
	conn, err := beanstalk.Dial("tcp", address)
	if err != nil {
		return sub, err
	}

	// Create one Tube object for each tube name.
	sub.Tubes = make(map[string]*beanstalk.Tube, 0)
	for _, name := range tubes {
		sub.Tubes[name] = &beanstalk.Tube{
			Name: name,
			Conn: conn,
		}
	}
	return sub, nil
}

func (s *Submitter) Submit(job *proto.ProcessedJobRequest, tube string) error {
	data, err := gproto.Marshal(job)
	_, err = s.Tubes[tube].Put(data, 10, 0, 12*time.Hour)

	//brittneyofthenorth+1loldataset5 means 4441 with qwerty
	return err
}

func (s *Submitter) Stats() {
	for _, tube := range s.Tubes {
		stats, _ := tube.Stats()

		// For each stat, print!
		fmt.Println(tube.Name)
		for k, v := range stats {
			fmt.Println(fmt.Sprintf("  %s: %s", k, v))
		}
		fmt.Println()
	}
}

package loglin

import (
	"fmt"
	logrus "github.com/Sirupsen/logrus"
	"net"
	"os"
	"strings"
	"time"
)

type Fields map[string]interface{}

type LogEvent struct {
	Id           int64
	Name         string
	StickyFields Fields

	// Loggers that rediect output.
	// StdoutLogger   *logrus.Logger
	JsonLogger *logrus.Logger
	UdpDest    *net.UDPConn
}

var (
	STATUS_START    uint = 0
	STATUS_OK       uint = 1
	STATUS_COMPLETE uint = 2
	STATUS_WARNING  uint = 3
	STATUS_ERROR    uint = 4
	STATUS_FATAL    uint = 5
)

func New(name string, sticky Fields) LogEvent {
	// logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stderr)

	le := LogEvent{}
	le.Id = time.Now().UnixNano()
	le.Name = name
	le.StickyFields = sticky
	le.JsonLogger = logrus.New()
	le.JsonLogger.Formatter = &logrus.JSONFormatter{}

	// Connect to an address on my internal network
	// TODO: parameterize this so that it's actually useful to others.
	connectionString := "logger.service.fy:10000"
	addr, err := net.ResolveUDPAddr("udp", connectionString)
	if err != nil {
		fmt.Println("[1 / 2] Cannot connect to logstash on " + connectionString)
	}

	le.UdpDest, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("[2 / 2] Cannot connect to logstash on " + connectionString)
	}

	le.Update(STATUS_START, name, nil)
	return le
}

func updateFields(sticky Fields, fields Fields, eventid int64) Fields {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["_process"] = strings.Replace(os.Args[0], "/", ".", -1)
	fields["_eventid"] = eventid

	// Copy sticky fields over whenever they exist.
	if sticky != nil {
		for k := range sticky {
			fields[k] = sticky[k]
		}
	}

	return fields
}

func (e *LogEvent) Update(status uint, message string, fields Fields) {
	fields = updateFields(e.StickyFields, fields, e.Id)

	switch status {
	case STATUS_START:
		e.Info(fmt.Sprintf("[STATUS_START]  %s", message), fields)
		break
	case STATUS_OK:
		e.Info(fmt.Sprintf("[STATUS_OK]     %s", message), fields)
		break
	case STATUS_COMPLETE:
		e.Info(fmt.Sprintf("[STATUS_COMPL]  %s", message), fields)
		break
	case STATUS_WARNING:
		e.Warn(fmt.Sprintf("[STATUS_WARN]   %s", message), fields)
		break
	case STATUS_ERROR:
		e.Error(fmt.Sprintf("[STATUS_ERROR] %s", message), fields)
		break
	case STATUS_FATAL:
		e.Fatal(fmt.Sprintf("[STATUS_FATAL] %s", message), fields)
		break
	}
}

func (e *LogEvent) Info(message string, fields Fields) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	logrus.WithFields(logrus.Fields(fields)).Info(message)

	data, _ := e.JsonLogger.Formatter.Format(e.JsonLogger.WithFields(logrus.Fields(fields)))
	e.UdpDest.Write(data)
}

func (e *LogEvent) Warn(message string, fields Fields) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	logrus.WithFields(logrus.Fields(fields)).Warn(message)

	data, _ := e.JsonLogger.Formatter.Format(e.JsonLogger.WithFields(logrus.Fields(fields)))
	e.UdpDest.Write(data)
}

func (e *LogEvent) Error(message string, fields Fields) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	logrus.WithFields(logrus.Fields(fields)).Error(message)

	data, _ := e.JsonLogger.Formatter.Format(e.JsonLogger.WithFields(logrus.Fields(fields)))
	e.UdpDest.Write(data)
}

func (e *LogEvent) Fatal(message string, fields Fields) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	logrus.WithFields(logrus.Fields(fields)).Fatal(message)

	data, _ := e.JsonLogger.Formatter.Format(e.JsonLogger.WithFields(logrus.Fields(fields)))
	e.UdpDest.Write(data)
}

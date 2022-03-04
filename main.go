package main

import (
	"os"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var (
	TestRunGroups []*TestRunGroup
)

type TestResults struct {
	ClientRunGroup      *TestRunGroup    `json:"client_run_group"`
	ServerRunGroup      *TestRunGroup    `json:"server_run_group"`
	Results             []*TestRunResult `json:"results"`
	AverageResponseTime int64            `json:"average_response_time_ns"`
}

type TestRunGroup struct {
	RunGroupID       string        `json:"run_group_id"`
	RunCount         int           `json:"run_count"`
	TestRuns         []*TestRun    `json:"test_runs"`
	Description      string        `json:"description"`
	CreatedAt        time.Time     `json:"created_at"`
	Concurrency      int           `json:"concurrency"`
	ClientDelayNS    int64         `json:"client_delay_ns"`
	ServerDelayNS    int64         `json:"server_delay_ns"`
	RemoteAddr       string        `json:"remote_addr,omitempty"`
	UpstreamEndpoint string        `json:"upstream_endpoint,omitempty"`
	UpstreamTimeout  time.Duration `json:"upstream_timeout_ns,omitempty"`
	Runs             chan *TestRun `json:"-"`
	Results          chan *TestRun `json:"-"`
	ClientReportFile string        `json:"-"`
	TestData         []byte        `json:"-"`
}

type TestRun struct {
	RunID                  string    `json:"run_id,omitempty"`
	RunGroupID             string    `json:"run_group_id,omitempty"`
	RemoteAddr             string    `json:"remote_addr,omitempty"`
	RunCount               int       `json:"run_count,omitempty"`
	Time                   time.Time `json:"time,omitempty"`
	Data                   []byte    `json:"data,omitempty"`
	ClientStart            time.Time `json:"client_start,omitempty"`
	ClientEnd              time.Time `json:"client_end,omitempty"`
	ServerStart            time.Time `json:"server_start,omitempty"`
	ServerEnd              time.Time `json:"server_end,omitempty"`
	UpstreamResponseCode   int       `json:"upstream_response_code,omitempty"`
	UpstreamResponseTimeNS int64     `json:"upstream_response_time_ns,omitempty"`
	Error                  string    `json:"error,omitempty"`
}

type TestRunResult struct {
	RunID          string        `json:"run_id"`
	RunGroupID     string        `json:"run_group_id,omitempty"`
	RemoteAddr     string        `json:"remote_addr,omitempty"`
	RunCount       int           `json:"run_count"`
	ClientTime     time.Time     `json:"client_time"`
	ServerTime     time.Time     `json:"server_time"`
	ClientDuration time.Duration `json:"client_duration_ns"`
	ServerDuration time.Duration `json:"server_duration_ns"`
	//ClientServerTimeDiff time.Duration `json:"client_server_time_diff"`
	Error string `json:"error,omitempty"`
}

func (t *TestRunGroup) Create() error {
	l := log.WithFields(log.Fields{
		"action": "TestRunGroup.Create",
	})
	l.Info("start")
	t.RunGroupID = uuid.New().String()
	t.CreatedAt = time.Now()
	TestRunGroups = append(TestRunGroups, t)
	return nil
}

func init() {
	var ll = log.InfoLevel
	var err error
	ll, err = log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

func main() {
	l := log.WithFields(log.Fields{
		"action": "main",
	})
	l.Info("start")
	if len(os.Args) == 1 {
		l.Error("no arguments")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "server":
		server()
	case "client":
		if err := client(); err != nil {
			l.Error(err)
			os.Exit(1)
		}
	default:
		l.Error("unknown action")
		os.Exit(1)
	}
}

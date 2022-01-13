package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func getTestRunGroup(remote string, id string) (*TestRunGroup, error) {
	l := log.WithFields(log.Fields{
		"action": "getTestRunGroup",
	})
	l.Info("start")
	c := &http.Client{}
	req, err := http.NewRequest("GET", remote+"/test-groups/"+id, nil)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	t := &TestRunGroup{}
	jerr := json.NewDecoder(resp.Body).Decode(t)
	if jerr != nil {
		l.Error(jerr)
		return nil, jerr
	}
	return t, nil
}

func (r *TestRun) Run() error {
	if len(r.Data) == 0 {
		r.Data = []byte(uuid.New().String())
	}
	l := log.WithFields(log.Fields{
		"action":     "Run",
		"id":         r.RunID,
		"runGroupID": r.RunGroupID,
		"runCount":   r.RunCount,
		"remote":     r.RemoteAddr,
		"data":       string(r.Data),
	})
	l.Info("start")
	r.ClientStart = time.Now()
	c := &http.Client{}
	src := strconv.Itoa(r.RunCount)
	req, err := http.NewRequest("POST", r.RemoteAddr+"/test-groups/"+r.RunGroupID+"/"+src, bytes.NewBuffer(r.Data))
	if err != nil {
		l.Error(err)
		r.Error = err.Error()
		r.ClientEnd = time.Now()
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		l.Error(err)
		r.Error = err.Error()
		r.ClientEnd = time.Now()
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		l.Error("bad status code")
		sc := strconv.Itoa(resp.StatusCode)
		r.Error = errors.New("bad status code: " + sc).Error()
		r.ClientEnd = time.Now()
		return errors.New("bad status code")
	}
	r.ClientEnd = time.Now()
	return nil
}

func rungroupWorker(runs chan *TestRun, res chan *TestRun) {
	for r := range runs {
		r.Run()
		res <- r
	}
}

func (t *TestRunGroup) CreateWork() error {
	l := log.WithFields(log.Fields{
		"action":          "CreateWorkers",
		"concurrency":     t.Concurrency,
		"runCount":        t.RunCount,
		"remote":          t.RemoteAddr,
		"client_delay_ns": t.ClientDelayNS,
	})
	l.Info("start")
	if t.Concurrency == 0 {
		t.Concurrency = 1
	}
	t.Runs = make(chan *TestRun, t.RunCount)
	t.Results = make(chan *TestRun, t.RunCount)
	for i := 0; i < t.Concurrency; i++ {
		go rungroupWorker(t.Runs, t.Results)
	}
	for i := 1; i <= t.RunCount; i++ {
		tr := &TestRun{
			RunGroupID: t.RunGroupID,
			RunCount:   i,
			RemoteAddr: t.RemoteAddr,
			Time:       time.Now(),
			Data:       t.TestData,
		}
		l.WithField("test_run", tr.RunCount).Info("start")
		t.Runs <- tr
		time.Sleep(time.Duration(t.ClientDelayNS * int64(time.Nanosecond)))
	}
	close(t.Runs)
	return nil
}

func (t *TestRunGroup) CollectResults() error {
	l := log.WithFields(log.Fields{
		"action": "CollectResults",
	})
	l.Info("start")
	for i := 0; i < t.RunCount; i++ {
		r := <-t.Results
		l.WithField("test_run", r.RunCount).Info("received")
		t.TestRuns = append(t.TestRuns, r)
	}
	return nil
}

func (t *TestRunGroup) WriteJSON() error {
	l := log.WithFields(log.Fields{
		"action": "WriteJSON",
		"file":   t.ClientReportFile,
	})
	l.Info("start")
	b, err := json.Marshal(t)
	if err != nil {
		l.Error(err)
		return err
	}
	if werr := ioutil.WriteFile(t.ClientReportFile, b, 0644); werr != nil {
		l.Error(werr)
		return werr
	}
	return nil
}

func (t *TestRunGroup) GetRunByID(id string) *TestRun {
	l := log.WithFields(log.Fields{
		"action":   "GetRunByID",
		"id":       id,
		"runCount": len(t.TestRuns),
	})
	l.Info("start")
	for _, tr := range t.TestRuns {
		if tr.RunID == id {
			l.WithField("test_run", tr.RunID).Info("found")
			return tr
		}
	}
	return nil
}

func (t *TestRunGroup) GetRunByCount(count int) *TestRun {
	l := log.WithFields(log.Fields{
		"action":   "GetRunByID",
		"count":    count,
		"runCount": len(t.TestRuns),
	})
	l.Info("start")
	for _, tr := range t.TestRuns {
		if tr.RunCount == count {
			l.WithField("test_run", tr.RunCount).Info("found")
			return tr
		}
	}
	return nil
}

func (r *TestResults) diff() error {
	l := log.WithFields(log.Fields{
		"action": "diff",
	})
	l.Info("start")
	var avg int64
	var totalRespTimeNS int64
	for _, tr := range r.ClientRunGroup.TestRuns {
		l.Info("start")
		res := &TestRunResult{
			RunGroupID: tr.RunGroupID,
			RunCount:   tr.RunCount,
			ClientTime: tr.Time,
			Error:      tr.Error,
		}
		sr := r.ServerRunGroup.GetRunByCount(tr.RunCount)
		if sr == nil {
			l.Error("server run not found")
			res.Error = "server run not found"
			continue
		}
		l = l.WithField("server_test_run", sr.RunID)
		l.Info("received")
		res.RunID = sr.RunID
		res.ServerTime = sr.Time
		res.ClientDuration = time.Duration(tr.ClientEnd.Sub(tr.ClientStart).Nanoseconds())
		//res.ClientServerTimeDiff = sr.Time.Sub(tr.Time)
		if string(tr.Data) != string(sr.Data) {
			res.Error = "data mismatch"
		}
		if tr.RunCount != sr.RunCount {
			res.Error = "run count mismatch"
		}
		r.Results = append(r.Results, res)
		totalRespTimeNS += res.ClientDuration.Nanoseconds()
		l.WithFields(log.Fields{
			"totalRespTimeNS":    totalRespTimeNS,
			"res.ClientDuration": res.ClientDuration,
		}).Info("totalRespTimeNS")
	}
	avg = int64(totalRespTimeNS) / int64(len(r.Results))
	l.WithField("avg", avg).Info("avg")
	r.AverageResponseTime = avg
	return nil
}

func (r *TestResults) Create() error {
	l := log.WithFields(log.Fields{
		"action": "TestResults.Create",
	})
	l.Info("start")
	rtg, err := getTestRunGroup(r.ClientRunGroup.RemoteAddr, r.ClientRunGroup.RunGroupID)
	if err != nil {
		l.Error(err)
		return err
	}
	r.ServerRunGroup = rtg
	l.WithField("server_run_group", r.ServerRunGroup.RunGroupID).Info("end")
	if err := r.diff(); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (r *TestResults) WriteJSON() error {
	l := log.WithFields(log.Fields{
		"action": "TestResults.WriteJSON",
		"file":   r.ClientRunGroup.ClientReportFile,
	})
	l.Info("start")
	b, err := json.Marshal(r)
	if err != nil {
		l.Error(err)
		return err
	}
	if werr := ioutil.WriteFile(r.ClientRunGroup.ClientReportFile, b, 0644); werr != nil {
		l.Error(werr)
		return werr
	}
	return nil
}

func (t *TestRunGroup) InitClient() error {
	l := log.WithFields(log.Fields{
		"action":     "InitClient",
		"remote":     t.RemoteAddr,
		"reportFile": t.ClientReportFile,
	})
	l.Info("start")
	tg, err := getTestRunGroup(t.RemoteAddr, t.RunGroupID)
	if err != nil {
		l.Error(err)
		return err
	}
	tg.TestData = t.TestData
	tg.RemoteAddr = t.RemoteAddr
	if werr := tg.CreateWork(); werr != nil {
		l.Error(werr)
		return werr
	}
	if cerr := tg.CollectResults(); cerr != nil {
		l.Error(cerr)
		return cerr
	}
	tg.ClientReportFile = t.ClientReportFile
	res := &TestResults{
		ClientRunGroup: tg,
	}
	if werr := res.Create(); werr != nil {
		l.Error(werr)
		return werr
	}
	if jerr := res.WriteJSON(); jerr != nil {
		l.Error(jerr)
		return jerr
	}
	return nil
}

func client() error {
	l := log.WithFields(log.Fields{
		"action": "client",
	})
	l.Info("start")
	var runGroupID string
	var remote string
	var reportFile string
	var data string
	cs := flag.NewFlagSet("client", flag.ExitOnError)
	cs.StringVar(&runGroupID, "g", "", "run group ID")
	cs.StringVar(&remote, "r", "", "remote server")
	cs.StringVar(&reportFile, "f", "", "report file")
	cs.StringVar(&data, "d", "", "data")

	cs.Parse(os.Args[2:])
	if runGroupID == "" {
		l.Error("no run group ID")
		return errors.New("no run group ID")
	}
	if remote == "" {
		l.Error("no remote server")
		return errors.New("no remote server")
	}
	if reportFile == "" {
		l.Error("no report file")
		return errors.New("no report file")
	}
	l = l.WithFields(log.Fields{
		"runGroupID": runGroupID,
		"remote":     remote,
		"reportFile": reportFile,
	})
	l.Info("start")
	tg := &TestRunGroup{
		RunGroupID:       runGroupID,
		RemoteAddr:       remote,
		ClientReportFile: reportFile,
		TestData:         []byte(data),
	}
	if err := tg.InitClient(); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

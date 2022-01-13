package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func handleCreateTestRunGroup(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(log.Fields{
		"action": "handleCreateTestRunGroup",
	})
	l.Info("start")
	t := &TestRunGroup{}
	defer r.Body.Close()
	jerr := json.NewDecoder(r.Body).Decode(t)
	if jerr != nil {
		l.Error(jerr)
		http.Error(w, jerr.Error(), http.StatusBadRequest)
		return
	}
	err := t.Create()
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jd, jerr := json.Marshal(t)
	if jerr != nil {
		l.Error(jerr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(jd)
}

func handleGetTestRunGroups(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(log.Fields{
		"action": "handleGetTestRunGroups",
	})
	l.Info("start")
	jd, jerr := json.Marshal(TestRunGroups)
	if jerr != nil {
		l.Error(jerr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(jd)
}

func handleGetTestRunGroup(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(log.Fields{
		"action": "handleGetTestRunGroup",
	})
	l.Info("start")
	vars := mux.Vars(r)
	runGroupID := vars["runGroupID"]
	for _, t := range TestRunGroups {
		if t.RunGroupID == runGroupID {
			jd, jerr := json.Marshal(t)
			if jerr != nil {
				l.Error(jerr)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(jd)
			return
		}
	}
	l.Error("not found")
	w.WriteHeader(http.StatusNotFound)
}

func handleTestRun(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(log.Fields{
		"action": "handleTestRun",
	})
	l.Info("start")
	vars := mux.Vars(r)
	runGroupID := vars["runGroupID"]
	count, err := strconv.Atoi(vars["count"])
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	bd, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Error(err)
	}
	for _, t := range TestRunGroups {
		if t.RunGroupID == runGroupID {
			l.Info("run group found")
			nrid := uuid.New().String()
			tr := &TestRun{
				RunID:    nrid,
				RunCount: count,
				Time:     time.Now(),
				Data:     bd,
			}
			l = l.WithFields(log.Fields{
				"runID": nrid,
			})
			l.Info("adding run")
			t.TestRuns = append(t.TestRuns, tr)
			time.Sleep(time.Duration(t.ServerDelayNS) * time.Nanosecond)
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	l.Error("not found")
	w.WriteHeader(http.StatusNotFound)
}

func server() {
	l := log.WithFields(log.Fields{
		"action": "server",
	})
	l.Info("start")
	var port string
	cs := flag.NewFlagSet("client", flag.ExitOnError)
	cs.StringVar(&port, "p", "8080", "port number")

	cs.Parse(os.Args[2:])
	r := mux.NewRouter()
	r.HandleFunc("/test-groups/create", handleCreateTestRunGroup).Methods("POST")
	r.HandleFunc("/test-groups/{runGroupID}", handleGetTestRunGroup).Methods("GET")
	r.HandleFunc("/test-groups/{runGroupID}/{count}", handleTestRun).Methods("POST")
	r.HandleFunc("/test-groups", handleGetTestRunGroups).Methods("GET")
	if port == "" {
		port = "8080"
	}
	l.WithField("port", port).Info("starting server")
	http.ListenAndServe(":"+port, r)
}

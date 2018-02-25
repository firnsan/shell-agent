package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"runtime"
)

func StatusMemHandler(w http.ResponseWriter, r *http.Request) {
	var stat runtime.MemStats
	runtime.ReadMemStats(&stat)

	b, err := json.Marshal(stat)
	if err != nil {
		log.Errorf("json.Marshal failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

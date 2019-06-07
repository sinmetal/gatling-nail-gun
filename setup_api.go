package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type SetupAPIRequest struct {
	SQL   string `json:"sql"`
	Digit int    `json:"digit"`
}

func handleSetupAPI(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	form := &SetupAPIRequest{}
	if err := json.Unmarshal(b, form); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("%s\n", string(b))

	pqs, err := NewPlanQueueService(TasksClient)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed NewPlanQueueService.err=%+v", err)
		return
	}

	if err := pqs.AddTask(r.Context(), PlanQueueTask{}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed AddTask.err=%+v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

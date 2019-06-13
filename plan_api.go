package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

func handlePlanAPI(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	form := &PlanQueueTask{}
	if err := json.Unmarshal(b, form); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("%s\n", string(b))

	iter := SpannerClient.Single().WithTimestampBound(spanner.ExactStaleness(time.Second*15)).QueryWithStats(r.Context(), spanner.Statement{
		SQL: form.SQL,
	})
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("failed spanner iter.Next().err=%+v", err)
			return
		}
		var id string
		if err := row.ColumnByName("Id", &id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("failed spanner row.ColumnByName().err=%+v", err)
			return
		}
		fmt.Printf("Id is %s\n", id)
	}
}

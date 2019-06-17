package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

type Tweet struct {
	ID         string `spanner:"Id"`
	Author     string
	Content    string
	Count      int64
	Favos      []string
	Sort       int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	CommitedAt time.Time
}

func HandleFireAPI(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	form := &FireQueueTask{}
	if err := json.Unmarshal(b, form); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("%s\n", string(b))

	sql := fmt.Sprintf(form.SQL, form.Param, form.LastID)
	fmt.Printf("Execute SQL %s\n", sql)
	iter := SpannerClient.Single().WithTimestampBound(spanner.ExactStaleness(time.Second*15)).QueryWithStats(r.Context(), spanner.Statement{
		SQL: sql,
	})
	defer iter.Stop()

	keySets := spanner.KeySets()

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

		keySets = spanner.KeySets(keySets, spanner.Key{id})
	}

	var lastID string
	var count int
	_, err = SpannerClient.ReadWriteTransaction(r.Context(), func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		var ml []*spanner.Mutation
		iter := txn.Read(ctx, "Tweet", keySets, []string{"Id", "Count", "CommitedAt", "UpdatedAt"})
		defer iter.Stop()
		for {
			row, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			var tweet Tweet
			if err := row.ToStruct(&tweet); err != nil {
				return err
			}
			tweet.Count++
			tweet.UpdatedAt = time.Now()
			cols := []string{"Id", "Count", "UpdatedAt", "CommitedAt", "SchemaVersion"}
			ml = append(ml, spanner.Update("Tweet", cols, []interface{}{tweet.ID, tweet.Count, tweet.UpdatedAt, spanner.CommitTimestamp, 1}))
			lastID = tweet.ID
			count++
		}
		return txn.BufferWrite(ml)
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed spanner Update Operation .err=%+v", err)
		return
	}
	fmt.Printf("Processing Count %d\n", count)

	pqs, err := NewFireQueueService(r.Host, TasksClient)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed NewFireQueueService. err=%+v", err)
		return
	}

	fmt.Printf("Last Id is %s\n", lastID)
	if err := pqs.AddTask(r.Context(), &FireQueueTask{
		SQL:    form.SQL,
		Param:  form.Param,
		LastID: lastID,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed FireQueueTask.AddTask. err=%+v", err)
		return
	}
}

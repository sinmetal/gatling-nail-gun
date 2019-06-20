package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

var TweetTableName string = "Tweet"

type Tweet struct {
	ID             string `spanner:"Id"`
	Author         string
	Content        string
	Count          int64
	Favos          []string
	Sort           int
	ShardCreatedAt int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CommitedAt     time.Time
	SchemaVersion  spanner.NullInt64
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
	log.Printf("FIRE API BODY:%s\n", string(b))

	if form.SQL == "" {
		log.Printf("required SQL\n")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if form.SchemaVersion == 0 {
		log.Printf("required SchemaVersion")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if form.Limit == 0 {
		log.Printf("required Limit")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if form.StartID == "" {
		log.Printf("required StartID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	count, lastID, err := Migration(r.Context(), form)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed spanner Update Operation .err=%+v", err)
		return
	}
	fmt.Printf("StartID:%s, Processing Count:%d\n", form.StartID, count)

	if lastID == "" {
		w.WriteHeader(http.StatusOK)
		fmt.Println("Finish!!")
		return
	}

	pqs, err := NewFireQueueService(r.Host, TasksClient)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed NewFireQueueService. err=%+v", err)
		return
	}

	fmt.Printf("Last Id is %s\n", lastID)
	if err := pqs.AddTask(r.Context(), &FireQueueTask{
		SQL:           form.SQL,
		Limit:         form.Limit,
		StartID:       form.StartID,
		LastID:        lastID,
		SchemaVersion: form.SchemaVersion,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed FireQueueTask.AddTask. err=%+v", err)
		return
	}
}

func Migration(ctx context.Context, form *FireQueueTask) (count int, lastID string, err error) {
	startID := form.StartID
	if form.LastID != "" {
		startID = form.LastID
	}
	sql := fmt.Sprintf(form.SQL, startID, form.Limit+1)
	fmt.Printf("Execute SQL %s\n", sql)
	iter := SpannerClient.Single().WithTimestampBound(spanner.ExactStaleness(time.Second*15)).QueryWithStats(ctx, spanner.Statement{
		SQL: sql,
	})
	defer iter.Stop()

	keySets := spanner.KeySets()
	var selectCount int
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return count, lastID, err
		}
		var id string
		if err := row.ColumnByName("Id", &id); err != nil {
			return count, lastID, err
		}
		if strings.HasPrefix(id, form.StartID) == false {
			fmt.Printf("%s has not prefix. prefix = %s\n", id, form.StartID)
			break
		}

		keySets = spanner.KeySets(keySets, spanner.Key{id})
		selectCount++
	}
	fmt.Printf("Select Count is %d\n", selectCount)
	if selectCount < 1 {
		return 0, lastID, nil
	}

	_, err = SpannerClient.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		var ml []*spanner.Mutation
		iter := txn.Read(ctx, TweetTableName, keySets, []string{"Id", "Count", "CommitedAt", "UpdatedAt", "SchemaVersion"})
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
			if count >= form.Limit {
				lastID = tweet.ID
				break
			}
			count++
			if tweet.SchemaVersion.Valid && tweet.SchemaVersion.Int64 >= form.SchemaVersion {
				fmt.Printf("%s goes through because SchemaVersion is data=%d, form=%d\n", tweet.ID, tweet.SchemaVersion.Int64, form.SchemaVersion)
				continue
			}
			tweet.Count++
			tweet.UpdatedAt = time.Now()
			cols := []string{"Id", "Count", "UpdatedAt", "CommitedAt", "SchemaVersion"}
			ml = append(ml, spanner.Update(TweetTableName, cols, []interface{}{tweet.ID, tweet.Count, tweet.UpdatedAt, spanner.CommitTimestamp, form.SchemaVersion}))
		}
		if len(ml) < 1 {
			return nil
		}
		return txn.BufferWrite(ml)
	})
	if err != nil {
		return 0, lastID, err
	}

	return count, lastID, nil
}

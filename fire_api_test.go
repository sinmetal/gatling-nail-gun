package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
)

func TestHandleFireAPI(t *testing.T) {
	t.SkipNow()

	hf := http.HandlerFunc(HandleFireAPI)
	server := httptest.NewServer(hf)
	defer server.Close()

	form := FireQueueTask{
		SQL:           `SELECT Id FROM TweetTest WHERE STARTS_WITH(Id, \"%v\") AND Id > \"%v\" AND Id < \"%v\" ORDER BY Id Limit %v`,
		SchemaVersion: 2,
		StartID:       "0",
		LastID:        "",
		Limit:         10,
	}
	b, err := json.Marshal(form)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(server.URL, "application/json; charset=utf8", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if e, g := http.StatusOK, resp.StatusCode; e != g {
		t.Errorf("StatusCode expected %v; got %v", e, g)
	}
}

// TestHandleFireAPI_ScenarioTest is シナリオテスト
// 全データが更新されるかをチェックする
func TestMigration(t *testing.T) {
	TweetTableName = "TweetTest"

	if err := createTestData(context.Background()); err != nil {
		t.Fatal(err)
	}

	form := &FireQueueTask{
		SQL:           `SELECT Id FROM TweetTest WHERE Id >= "%v" ORDER BY Id Limit %v`,
		SchemaVersion: 1,
		StartID:       "00",
		LastID:        "",
		Limit:         10,
	}

	{
		count, lastID, err := Migration(context.Background(), form)
		if err != nil {
			t.Fatal(err)
		}
		if e, g := 10, count; e != g {
			t.Errorf("count expected %d, got %d\n", e, g)
		}
		if e, g := "009", lastID; e != g {
			t.Errorf("lastID expected %v, got %v\n", e, g)
		}

		form.LastID = lastID
		count, lastID, err = Migration(context.Background(), form)
		if err != nil {
			t.Fatal(err)
		}
		if e, g := 1, count; e != g {
			t.Errorf("count expected %d, got %d\n", e, g)
		}
		if e, g := "", lastID; e != g {
			t.Errorf("lastID expected %v, got %v\n", e, g)
		}
	}

	{
		form.StartID = "01"
		form.LastID = ""
		count, lastID, err := Migration(context.Background(), form)
		if err != nil {
			t.Fatal(err)
		}
		if e, g := 10, count; e != g {
			t.Errorf("count expected %d, got %d\n", e, g)
		}
		if e, g := "019", lastID; e != g {
			t.Errorf("lastID expected %v, got %v\n", e, g)
		}

		form.LastID = lastID
		count, lastID, err = Migration(context.Background(), form)
		if err != nil {
			t.Fatal(err)
		}
		if e, g := 1, count; e != g {
			t.Errorf("count expected %d, got %d\n", e, g)
		}
		if e, g := "", lastID; e != g {
			t.Errorf("lastID expected %v, got %v\n", e, g)
		}
	}

	{
		form.StartID = "zz"
		form.LastID = ""
		count, lastID, err := Migration(context.Background(), form)
		if err != nil {
			t.Fatal(err)
		}
		if e, g := 0, count; e != g {
			t.Errorf("count expected %d, got %d\n", e, g)
		}
		if e, g := "", lastID; e != g {
			t.Errorf("lastID expected %v, got %v\n", e, g)
		}
	}
}

func createTestData(ctx context.Context) error {
	var ms []*spanner.Mutation

	for i := 0; i < 11; i++ {
		m, err := spanner.InsertOrUpdateStruct("TweetTest", &Tweet{
			ID:            fmt.Sprintf("00%d", i),
			Favos:         []string{},
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			CommitedAt:    spanner.CommitTimestamp,
			SchemaVersion: spanner.NullInt64{Int64: 0},
		})
		if err != nil {
			return err
		}
		ms = append(ms, m)
	}
	for i := 0; i < 11; i++ {
		m, err := spanner.InsertOrUpdateStruct("TweetTest", &Tweet{
			ID:            fmt.Sprintf("01%d", i),
			Favos:         []string{},
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			CommitedAt:    spanner.CommitTimestamp,
			SchemaVersion: spanner.NullInt64{Int64: 0},
		})
		if err != nil {
			return err
		}
		ms = append(ms, m)
	}

	_, err := SpannerClient.Apply(ctx, ms)
	if err != nil {
		return err
	}

	return nil
}

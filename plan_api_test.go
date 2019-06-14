package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlePlanAPI(t *testing.T) {
	hf := http.HandlerFunc(HandlePlanAPI)
	server := httptest.NewServer(hf)
	defer server.Close()

	form := PlanQueueTask{
		SQL:   `SELECT Id FROM Tweet WHERE STARTS_WITH(Id, '%v') ORDER BY Id Limit 1000`,
		Param: "0",
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

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
		SQL:    `SELECT Id FROM Tweet WHERE STARTS_WITH(Id, '%v') AND Id > '%v' ORDER BY Id Limit 1000`,
		Param:  "0",
		LastID: "00002fd0-1152-49c4-9275-4e9bb7d0c0d0c6e70da0-904d-4858-ae61-5d847cc78894",
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

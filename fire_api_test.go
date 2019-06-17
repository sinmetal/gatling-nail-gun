package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleFireAPI(t *testing.T) {
	hf := http.HandlerFunc(HandleFireAPI)
	server := httptest.NewServer(hf)
	defer server.Close()

	form := FireQueueTask{
		SQL:    `SELECT Id FROM Tweet WHERE STARTS_WITH(Id, '%v') AND Id > '%v' ORDER BY Id Limit 1000`,
		Param:  "0",
		LastID: "",
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

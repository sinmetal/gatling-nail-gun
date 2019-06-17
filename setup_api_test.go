package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleSetupAPI(t *testing.T) {
	t.SkipNow()

	hf := http.HandlerFunc(HandleSetupAPI)
	server := httptest.NewServer(hf)
	defer server.Close()

	form := SetupAPIRequest{
		SQL: `SELECT Id FROM Tweet WHERE STARTS_WITH(Id, '%v') AND Id > '%v' AND Id < '%v' ORDER BY Id Limit 1000`,
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

func TestGenerateUUIDPrefix(t *testing.T) {
	fmt.Printf("UUID:%+v\n", GenerateUUIDPrefix())
}

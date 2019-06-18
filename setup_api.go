package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type SetupAPIRequest struct {
	SQL           string `json:"sql"`
	SchemaVersion int64  `json:"schemaVersion"`
	Limit         int    `json:"limit"`
	Digit         int    `json:"digit"` // TODO UUIDの桁数を指定しようかと思っているが未実装
}

func HandleSetupAPI(w http.ResponseWriter, r *http.Request) {
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

	pqs, err := NewFireQueueService(r.Host, TasksClient)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed NewFireQueueService.err=%+v", err)
		return
	}

	prefix := GenerateUUIDPrefix()
	for _, p := range prefix {
		log.Printf("SQL is %s, Start:%s\n", form.SQL, p)
		if err := pqs.AddTask(r.Context(), &FireQueueTask{
			SQL:           form.SQL,
			SchemaVersion: form.SchemaVersion,
			Limit:         form.Limit,
			StartID:       p,
			LastID:        "",
		}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("failed AddTask.err=%+v", err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("DONE"))
	if err != nil {
		log.Printf("failed response.write.err=%+v", err)
	}
}

// GenerateUUIDPrefix is UUIDの先頭文字になりえる2桁の文字列一覧を返す
// for example 00, 01, 02..., a0, a1, a2 ...
func GenerateUUIDPrefix() []string {
	var runeList []string
	for i := 0; i < 10; i++ {
		runeList = append(runeList, fmt.Sprintf("%d", i))
	}
	for i := 0; i < 6; i++ {
		r := rune('a' + i)
		runeList = append(runeList, fmt.Sprintf("%v", string(r)))
	}

	var results []string
	for _, i := range runeList {
		for _, j := range runeList {
			results = append(results, fmt.Sprintf("%s%s", i, j))
		}
	}

	return results
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type SetupAPIRequest struct {
	SQL   string `json:"sql"`
	Digit int    `json:"digit"` // TODO UUIDの桁数を指定しようかと思っているが未実装
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

	pqs, err := NewPlanQueueService(r.Host, TasksClient)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed NewPlanQueueService.err=%+v", err)
		return
	}

	prefix := GenerateUUIDPrefix()
	for _, p := range prefix {
		log.Printf("SQL is %s, Param is %s\n", form.SQL, p)
		if err := pqs.AddTask(r.Context(), &PlanQueueTask{
			SQL:    form.SQL,
			Param:  p,
			LastID: "",
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
	// TODO 2段だと多いので、とりあえず1段に...
	//for _, i := range runeList {
	//	for _, j := range runeList {
	//		results = append(results, fmt.Sprintf("%s%s", i, j))
	//	}
	//}
	for _, i := range runeList {
		results = append(results, fmt.Sprintf("%s", i))
	}

	return results
}

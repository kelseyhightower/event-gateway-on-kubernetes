package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type CloudEvent struct {
	EventType          string      `json:"eventType"`
	EventID            string      `json:"eventID"`
	CloudEventsVersion string      `json:"cloudEventsversion"`
	ContentType        string      `json:"contentType"`
	Source             string      `json:"source"`
	EventTime          string      `json:"eventTime"`
	Data               interface{} `json:"data"`
}

func main() {
	log.Println("Starting HTTP server...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}
		r.Body.Close()

		var ce CloudEvent

		if err := json.Unmarshal(data, &ce); err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		log.Printf("Handling event %s from %s ...", ce.EventID, ce.Source)

		fmt.Println(ce.Data)
	})

	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}

// Copyright 2018 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

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

		log.Printf("Handling event %s (%s) ...", ce.EventID, ce.EventType)

		m, err := toMap(ce.Data)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		fmt.Println(m["message"])
	})

	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}

func toMap(v interface{}) (map[string]string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var m map[string]string

	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return m, nil
}

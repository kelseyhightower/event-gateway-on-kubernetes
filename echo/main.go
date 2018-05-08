// Copyright 2018 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package main

import (
	"encoding/json"
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

type HTTPEvent struct {
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Host    string            `json:"host"`
	Query   map[string]string `json:"query"`
	Params  map[string]string `json:"params"`
	Body    string            `json:"body"`
}

type HTTPResponse struct {
	Body       string            `json:"body"`
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers,omitempty"`
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

		log.Printf("Handling HTTP event %s ...", ce.EventID)

		e, err := httpEvent(ce.Data)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		response := HTTPResponse{
			Body:       string(e.Body),
			StatusCode: 200,
		}

		data, err = json.MarshalIndent(&response, "", " ")
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		w.Write(data)
	})

	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}

func httpEvent(v interface{}) (*HTTPEvent, error) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Println("can't marshal HTTP event")
		return nil, err
	}

	var e HTTPEvent

	if err := json.Unmarshal(data, &e); err != nil {
		log.Println("can't unmarshal HTTP event")
		return nil, err
	}

	return &e, nil
}

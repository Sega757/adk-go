// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package controllers contains the controllers for the ADK REST API.
package controllers

import (
	"encoding/json"
	"log"
	"net/http"
)

// TODO: Move to an internal package, controllers doesn't have to be public API.

// EncodeJSONResponse uses json.Marshal to serialize an interface to JSON before writing headers and body.
// Performing JSON serialization prior to calling WriteHeader prevents returning HTTP 200 OK or writing
// partial responses when serialization fails.
func EncodeJSONResponse(i any, status int, w http.ResponseWriter) {
	wHeader := w.Header()

	if i == nil {
		wHeader.Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(status)
		return
	}

	data, err := json.Marshal(i)
	if err != nil {
		log.Printf("EncodeJSONResponse failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	wHeader.Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		log.Printf("EncodeJSONResponse write failed: %v", err)
	}
}

type errorHandler func(http.ResponseWriter, *http.Request) error

// NewErrorHandler writes the error code returned from the http handler.
func NewErrorHandler(fn errorHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			if statusErr, ok := err.(statusError); ok {
				if statusErr.Status() == http.StatusInternalServerError {
					log.Printf("Internal server error: %v", statusErr.Error())
					http.Error(w, "internal server error", http.StatusInternalServerError)
				} else {
					http.Error(w, statusErr.Error(), statusErr.Status())
				}
			} else {
				log.Printf("Internal server error: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
	}
}

// Unimplemented returns 501 - Status Not Implemented error
func Unimplemented(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusNotImplemented)
}

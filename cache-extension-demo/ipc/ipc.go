// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package ipc

import (
	"cache-extension-demo/extension"
	"cache-extension-demo/plugins"
	"net/http"

	"github.com/gorilla/mux"
)

// Start begins running the sidecar
func Start(port string) {
	go startHTTPServer(port)
}

// Method that responds back with the cached values
func startHTTPServer(port string) {
	router := mux.NewRouter()
	router.Path("/{cacheType}").Queries("name", "{name}").Methods("GET").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			value := extension.RouteCache(vars["cacheType"], vars["name"])

			if len(value) != 0 {
				_, _ = w.Write([]byte(value))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("No data found"))
			}
		})
	router.Path("/{cacheType}/{name}").Queries("value", "{value}").Methods("PUT").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			response := extension.PutCache(vars["cacheType"], vars["name"], vars["value"])
			if len(response) != 0 {
				_, _ = w.Write([]byte(response))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Can't store value to cache"))
			}
		})

	println(plugins.PrintPrefix, "Starting Httpserver on port ", port)
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		println(plugins.PrintPrefix, "Error serving: "+err.Error())
		panic(err)
	}
}

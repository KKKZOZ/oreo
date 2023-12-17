package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type MemoryDatabase struct {
	Address string
	Port    int
	records map[string]string
	server  http.Server
}

func NewMemoryDatabase(address string, port int) *MemoryDatabase {
	return &MemoryDatabase{
		Address: address,
		Port:    port,
		records: make(map[string]string),
	}
}

func (m *MemoryDatabase) serveGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	if value, ok := m.records[key]; ok {
		fmt.Fprintf(w, value)
	} else {
		// response with 404
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Key not found")
	}
}

func (m *MemoryDatabase) servePut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request")
		return
	}
	// Access form values
	value := r.FormValue("value")
	if value == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Value is empty")
		return
	}

	m.records[key] = value
	fmt.Fprintf(w, "OK")
}

func (m *MemoryDatabase) serveDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	if _, ok := m.records[key]; ok {
		delete(m.records, key)
		fmt.Fprintf(w, "OK")
	} else {
		// response with 404
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Key not found")
	}
}

func (m *MemoryDatabase) start() {

	router := mux.NewRouter()
	router.HandleFunc("/get/{key}", m.serveGet).Methods("GET")
	router.HandleFunc("/put/{key}", m.servePut).Methods("POST")
	router.HandleFunc("/delete/{key}", m.serveDelete).Methods("DELETE")

	m.server = http.Server{
		Addr:    fmt.Sprintf("%s:%d", m.Address, m.Port),
		Handler: router,
	}

	log.Fatal(m.server.ListenAndServe())
}

func (m *MemoryDatabase) stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()
	m.server.Shutdown(ctx)
}

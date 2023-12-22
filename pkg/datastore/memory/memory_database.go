package memory

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type MemoryDatabase struct {
	mu      sync.Mutex
	Address string
	Port    int
	records map[string]string
	server  http.Server
	MsgChan chan string
}

func NewMemoryDatabase(address string, port int) *MemoryDatabase {
	return &MemoryDatabase{
		Address: address,
		Port:    port,
		records: make(map[string]string),
		MsgChan: make(chan string),
	}
}

func (m *MemoryDatabase) serveGet(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	vars := mux.Vars(r)
	key := vars["key"]
	if value, ok := m.records[key]; ok {
		fmt.Fprint(w, value)
	} else {
		// response with 404
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Key not found")
	}
}

func (m *MemoryDatabase) servePut(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

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

func (m *MemoryDatabase) serveHeartbeat(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func (m *MemoryDatabase) Start() error {

	router := mux.NewRouter()
	router.HandleFunc("/get/{key}", m.serveGet).Methods("GET")
	router.HandleFunc("/put/{key}", m.servePut).Methods("POST")
	router.HandleFunc("/delete/{key}", m.serveDelete).Methods("DELETE")
	router.HandleFunc("/heartbeat", m.serveHeartbeat).Methods("GET")

	m.server = http.Server{
		Addr:    fmt.Sprintf("%s:%d", m.Address, m.Port),
		Handler: router,
	}
	return m.server.ListenAndServe()
}

func (m *MemoryDatabase) Stop() {
	ctx, _ := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	m.server.Shutdown(ctx)
	go func() { m.MsgChan <- "Memory database stopped" }()
}

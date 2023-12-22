package timeoracle

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/kkkzoz/vanilla-icecream/locker"
)

type SimpleTimeOracle struct {
	Address string
	Port    int
	baseURL string
	server  http.Server
	MsgChan chan string
	locker  locker.Locker
}

func NewSimpleTimeOracle(address string, port int, locker locker.Locker) *SimpleTimeOracle {
	return &SimpleTimeOracle{
		Address: address,
		Port:    port,
		baseURL: fmt.Sprintf("http://%s:%d", address, port),
		MsgChan: make(chan string),
		locker:  locker,
	}
}

func (s *SimpleTimeOracle) GetTime() time.Time {
	return time.Now()
}

func (s *SimpleTimeOracle) serveTime(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", s.GetTime())
}

func (s *SimpleTimeOracle) serveLock(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request")
		return
	}
	// Access form values
	key := r.FormValue("key")
	id := r.FormValue("id")
	durationStr := r.FormValue("duration")

	// validate
	if key == "" || id == "" || durationStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Missing parameters")
		return
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid duration")
		return
	}
	err = s.locker.Lock(key, id, time.Duration(duration)*time.Millisecond)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Lock failed")
		return
	}
	fmt.Fprintf(w, "OK")
}

func (s *SimpleTimeOracle) serveUnlock(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
		return
	}
	// Access form values
	key := r.FormValue("key")
	id := r.FormValue("id")

	// validate
	if key == "" || id == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Missing parameters")
		return
	}

	err = s.locker.Unlock(key, id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprintf(w, "OK")
}

func (s *SimpleTimeOracle) Start() error {

	router := mux.NewRouter()
	router.HandleFunc("/time", s.serveTime).Methods("GET")
	router.HandleFunc("/lock", s.serveLock).Methods("GET")
	router.HandleFunc("/unlock", s.serveUnlock).Methods("GET")
	s.server = http.Server{
		Addr:    s.baseURL,
		Handler: router,
	}

	return s.server.ListenAndServe()
}

func (s *SimpleTimeOracle) WaitForStartUp(timeout time.Duration) error {
	ch := make(chan bool)
	go func() {
		for {
			_, err := http.Get(s.baseURL + "/time")
			if err == nil {
				ch <- true
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("server did not reply after %v", timeout)
	}
}

func (s *SimpleTimeOracle) Stop() error {
	ctx, _ := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	s.server.Shutdown(ctx)
	s.MsgChan <- "Simple Time Oracle stopped"
	return nil
}

package timeoracle

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/oreo-dtx-lab/oreo/pkg/locker"
	"github.com/oreo-dtx-lab/oreo/pkg/logger"
)

// SimpleTimeOracle represents a simple time oracle that provides time-related information.
// It contains the address, port, base URL, server, message channel, and locker.
type SimpleTimeOracle struct {
	// Address is the IP address of the time oracle.
	Address string
	// Port is the port number of the time oracle.
	Port int
	// baseURL is the base URL of the time oracle.
	baseURL string
	// server is the HTTP server used by the time oracle.
	server http.Server
	// MsgChan is the channel used for sending and receiving messages.
	MsgChan chan string
	// locker is used for synchronizing access to the time oracle.
	locker locker.Locker
}

// NewSimpleTimeOracle creates a new instance of SimpleTimeOracle.
func NewSimpleTimeOracle(address string, port int, locker locker.Locker) *SimpleTimeOracle {
	return &SimpleTimeOracle{
		Address: address,
		Port:    port,
		baseURL: fmt.Sprintf("http://%s:%d", address, port),
		MsgChan: make(chan string),
		locker:  locker,
	}
}

// GetTime returns the current time as reported by the SimpleTimeOracle.
func (s *SimpleTimeOracle) GetTime() time.Time {
	return time.Now()
}

func (s *SimpleTimeOracle) serveTime(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "%s", s.GetTime())
	logger.CheckAndLogError("Failed to write time response", err)
}

func (s *SimpleTimeOracle) serveLock(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err = fmt.Fprintf(w, "Bad request")
		logger.CheckAndLogError("Failed to write response", err)
		return
	}
	// Access form values
	key := r.FormValue("key")
	id := r.FormValue("id")
	durationStr := r.FormValue("duration")

	// validate
	if key == "" || id == "" || durationStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, err := fmt.Fprintf(w, "Missing parameters")
		logger.CheckAndLogError("Failed to write response", err)
		return
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err = fmt.Fprintf(w, "Invalid duration")
		logger.CheckAndLogError("Failed to write response", err)
		return
	}
	err = s.locker.Lock(key, id, time.Duration(duration)*time.Millisecond)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err = fmt.Fprintf(w, "Lock failed")
		logger.CheckAndLogError("Failed to write response", err)
		return
	}
	_, err = fmt.Fprintf(w, "OK")
	logger.CheckAndLogError("Failed to write response", err)
}

func (s *SimpleTimeOracle) serveUnlock(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err = fmt.Fprint(w, err.Error())
		logger.CheckAndLogError("Failed to parse form", err)
		return
	}
	// Access form values
	key := r.FormValue("key")
	id := r.FormValue("id")

	// validate
	if key == "" || id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, err = fmt.Fprintf(w, "Missing parameters")
		logger.CheckAndLogError("Failed to write response", err)
		return
	}

	err = s.locker.Unlock(key, id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err = fmt.Fprint(w, err.Error())
		logger.CheckAndLogError("Failed to write response", err)
		return
	}
	_, err = fmt.Fprintf(w, "OK")
	logger.CheckAndLogError("Failed to write response", err)
}

// Start starts the SimpleTimeOracle server.
// It initializes the router and sets up the necessary routes for serving time, locking, and unlocking.
// It then starts the server and listens for incoming requests.
// Returns an error if there was a problem starting the server.
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

// WaitForStartUp waits for the server to start up by continuously sending HTTP requests to the "/time" endpoint until a successful response is received or the timeout is reached.
// It returns an error if the server does not reply within the specified timeout duration.
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

// Stop stops the SimpleTimeOracle server gracefully.
// It shuts down the server and sends a message to the MsgChan indicating that the Simple Time Oracle has stopped.
// It returns an error if there was an issue shutting down the server.
func (s *SimpleTimeOracle) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	_ = s.server.Shutdown(ctx)
	go func() { s.MsgChan <- "Simple Time Oracle stopped" }()
	return nil
}

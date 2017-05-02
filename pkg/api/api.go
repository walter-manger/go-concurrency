package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/walter-manger/go-concurrency/pkg/hasher"
)

// HasherAPI represents the API for interacting with Hasher
type HasherAPI struct {
	RequestChannel chan bool
	hasher         *hasher.Hasher
	port           string
	requestTimes   []time.Duration
}

// NewHasherAPI returns a pointer to a new instance of HasherAPI
func NewHasherAPI(port string) *HasherAPI {
	hasher := hasher.NewHasher()
	return &HasherAPI{
		RequestChannel: make(chan bool),
		hasher:         hasher,
		port:           port,
		requestTimes:   []time.Duration{},
	}
}

// HasherJobs returns Hasher's WgJobs
func (hapi *HasherAPI) HasherJobs() *sync.WaitGroup {
	return hapi.hasher.WgJobs
}

// Start sets up the API for routing on the desired port
func (hapi *HasherAPI) Start() {
	p := hapi.port
	go func() {
		log.Printf("Hasher API accepting connections on port %s\n", p)
		http.HandleFunc("/hash", hapi.createHandler(hapi.postHashHandler, "POST"))
		http.HandleFunc("/hash/", hapi.createHandler(hapi.getHashHandler, "GET"))
		http.HandleFunc("/stats", hapi.createHandler(hapi.getStatsHandler, "GET"))
		log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%s", p), nil))
	}()
}

func (hapi *HasherAPI) createHandler(handler http.HandlerFunc, method string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		select {
		case <-hapi.RequestChannel:
			http.Error(w, "Not accepting new connections", http.StatusLocked)
			return
		default:
		}

		if r.Method != method {
			http.Error(w, fmt.Sprintf("This endpoint only supports %s requests", method), http.StatusMethodNotAllowed)
			return
		}

		handler(w, r)
	}
}

func (hapi *HasherAPI) postHashHandler(w http.ResponseWriter, r *http.Request) {

	pw := r.FormValue("password")

	if pw == "" {
		http.Error(w, "password is a required field", http.StatusUnprocessableEntity)
		return
	}

	job := hapi.hasher.RunHash(pw)
	fmt.Fprintf(w, fmt.Sprintf("%d", job))
}

func (hapi *HasherAPI) getHashHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/hash/"):]
	s := time.Now()

	jobID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "jobID was an incorrect format", http.StatusUnprocessableEntity)
		return
	}

	val, err := hapi.hasher.GetPassword(jobID)
	if err != nil {
		http.Error(w, "JobID is not a valid ID", http.StatusUnprocessableEntity)
		return
	}

	for val == "" {
		time.Sleep(250 * time.Millisecond)
		val, err = hapi.hasher.GetPassword(jobID)
		if err != nil {
			http.Error(w, "JobID is not a valid ID", http.StatusUnprocessableEntity)
			return
		}
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", val))
	hapi.requestTimes = append(hapi.requestTimes, time.Since(s))
	return
}

func (hapi *HasherAPI) getStatsHandler(w http.ResponseWriter, r *http.Request) {

	type stats struct {
		Total   int   `json:"total"`
		Average int64 `json:"average"`
	}

	var sum time.Duration

	sum = 0
	for _, v := range hapi.requestTimes {
		fmt.Println(v)
		sum += v
	}

	milli := time.Millisecond
	conv := sum.Nanoseconds()
	avg := conv / 2

	s := stats{
		Total:   hapi.hasher.GetJobCount(),
		Average: int64(avg / int64(milli)),
	}

	j, err := json.Marshal(s)
	if err != nil {
		http.Error(w, "Could not marshal stats to json", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		http.Error(w, "Could not write json", http.StatusInternalServerError)
	}
}

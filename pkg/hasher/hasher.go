package hasher

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"log"
	"sync"
	"time"
)

// State represents the state of a Hasher instance.
type State struct {
	JobCount        int
	HashedPasswords map[int]string
}

// Inc initializes the next state by incrementing the JobCount
// and setting up the HashedPassword store.
func (s *State) Inc() int {
	s.JobCount++
	s.HashedPasswords[s.JobCount] = ""
	return s.JobCount
}

// Hasher represents a hashing service
type Hasher struct {
	mu     *sync.Mutex
	state  State
	WgJobs *sync.WaitGroup
}

// NewHasher initializes and returns a new Hasher instance.
func NewHasher() *Hasher {
	return &Hasher{
		mu: &sync.Mutex{},
		state: State{
			JobCount:        0,
			HashedPasswords: map[int]string{},
		},
		WgJobs: &sync.WaitGroup{},
	}
}

// GetPassword safely reads and returns a stored password from
// the hasher state.
func (h *Hasher) GetPassword(jobID int) (string, error) {
	h.mu.Lock()
	val, ok := h.state.HashedPasswords[jobID]
	h.mu.Unlock()

	if !ok {
		return "", errors.New("Invalid jobID")
	}

	return val, nil
}

// GetJobCount safely reads and returns the stored jobcount from
// the hasher state.
func (h *Hasher) GetJobCount() int {
	h.mu.Lock()
	c := h.state.JobCount
	h.mu.Unlock()

	return c
}

// RunHash runs the hashing function against the password
// provided.
func (h *Hasher) RunHash(password string) int {
	h.mu.Lock()
	h.state.JobCount++
	jobID := h.state.JobCount
	h.state.HashedPasswords[jobID] = ""
	h.mu.Unlock()

	go h.hash(jobID, password)
	return jobID
}

func (h *Hasher) hash(jobID int, in string) {
	log.Printf("Starting hash jobID: %d\n", jobID)

	h.WgJobs.Add(1)

	c := sha512.New()

	_, err := c.Write([]byte(in))
	if err != nil {
		log.Fatalf("Could not hash password for jobID: %d", jobID)
		return
	}

	time.Sleep(5 * time.Second)
	pw := base64.StdEncoding.EncodeToString(c.Sum(nil))

	h.mu.Lock()
	h.state.HashedPasswords[jobID] = pw
	h.mu.Unlock()

	log.Printf("Finished hash for jobID: %d\n", jobID)
	h.WgJobs.Done()
}

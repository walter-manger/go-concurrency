package main

import (
	"crypto/sha512"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var wgHashRequests *sync.WaitGroup
var wgPasswordRequests *sync.WaitGroup
var jobIDs []string
var jobPasswords map[string]string

func init() {
	wgHashRequests = &sync.WaitGroup{}
	wgPasswordRequests = &sync.WaitGroup{}
	jobIDs = []string{}
	jobPasswords = map[string]string{}
	rand.Seed(time.Now().Unix())
}

func main() {

	addr := flag.String("addr", "http://localhost:8080", "The address to send hash requests to")

	numReq := flag.Int("numReq", 10, "The number of requests to run against the address")

	flag.Parse()

	log.Printf("Running the Hasher CLI on '%s', %d times\n\n", *addr, *numReq)

	wgHashRequests.Add(*numReq)
	for i := 0; i < *numReq; i++ {
		go makeHashRequest(addr, i)
	}

	// Let's wait until we at least have 2 jobs complete before
	// requesting the password
	for len(jobIDs) < 2 {
		time.Sleep(500 * time.Millisecond)
	}

	wgPasswordRequests.Add(*numReq)
	for i := 0; i < *numReq; i++ {
		go makePasswordRequest(addr)
	}

	wgHashRequests.Wait()
	wgPasswordRequests.Wait()

	makeStatsRequest(addr)
}

func makeHashRequest(addr *string, i int) {

	r := rand.Int63n(10-1) + 1
	log.Printf("Waiting %d Seconds To Request", r)

	time.Sleep(time.Duration(r) * time.Second)

	p := fmt.Sprintf("angryMonkey%d", i)

	res, err := http.PostForm(fmt.Sprintf("%s/hash", *addr), url.Values{"password": {p}})
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Fatal(res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Job Response: %s", string(body))
	jobID := string(body)
	jobIDs = append(jobIDs, jobID)
	jobPasswords[jobID] = p

	wgHashRequests.Done()
}

func makePasswordRequest(addr *string) {

	r := rand.Int63n(700-200) + 200
	time.Sleep(time.Duration(r) * time.Millisecond)

	l := len(jobIDs) - 1
	i := rand.Intn(l) + 0
	jobID := jobIDs[i]
	log.Printf("Requesting JobID: %s\n", jobID)
	res, err := http.Get(fmt.Sprintf("%s/hash/%s", *addr, jobID))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Fatal(res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	pw := string(body)
	log.Printf("Password Response: %s", pw)
	checkPasswordHash(jobPasswords[jobID], pw)
	wgPasswordRequests.Done()
}

func makeStatsRequest(addr *string) {
	res, err := http.Get(fmt.Sprintf("%s/stats", *addr))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Fatal(res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Stats Response: %s", string(body))
}

func checkPasswordHash(password string, response string) {

	c := sha512.New()
	_, err := c.Write([]byte(password))
	if err != nil {
		log.Fatal("Could not hash password for check")
		return
	}

	expected := base64.StdEncoding.EncodeToString(c.Sum(nil))

	if expected != response {
		log.Fatal("********* Password Hash Didn't Match")
	}

	log.Println("Password Hash Check Passed!")
}

# Go Concurrency (Hasher)

This project is an exploration of the concurrency techniques used in Golang.

## Structure ##

``` text
- cmd
  - hashercli
    - main.go
  - hashersrv
    - main.go
- pkg
  - hasher
    - hasher.go
  - api
    - api.go
```

## Starting the Hasher Service

`go run cmd/hashersrv/main.go`

Will start the Hasher Service and listen on port 8080 by default.

### Need Help? 

``` text
go run cmd/hashersrv/main.go --help
2017/05/02 15:04:40 Starting Hasher Service...

  -addr string
    	The port to listen on (default "8080")
exit status 2
```

From here, you can interact with the service using these endpoints:

## Endpoints

### POST /hash

Starts a Hasher job and returns the jobID. _All Hasher jobs take 5 seconds to perform_.

`curl --data "password=somePassword" http://localhost:8080/hash`

Example Output:

`21`

### GET /hash/{jobID}

Waits (blocks) for the Hasher job to finish and returns the SHA512 + base64 Encoded password. 

`curl http://localhost:8080/hash/21`

Example Output:

``` text
auSfSFKDloNDlfwJo68gFLWDtbIqzMkf8gGnd8sw9dEpMmciJAh73g9M+BzRzR6F5vQgRVoaQRwSvZaOsTVHLw==
```

### GET /stats

Returns a JSON describing statistics for the running Hasher service.

`curl http://localhost:8080/stats`

Example Output:

`{"total":2,"average": 221}`

Where `total` is the number of hash requests since the Hasher Service started, and `average` is the average time in milliseconds that a `/hash/{jobID}` took to complete. 

## An Alternative to curl

Alternatively to running `curl`, you can interact with the Hasher Service using the Hasher CLI tool. 

`go run cmd/hashercli/main.go`

### Need Help?

``` text
go run cmd/hashercli/main.go --help
  -addr string
    	The address to send hash requests to (default "http://localhost:8080")
  -numReq int
    	The number of requests to run against the address (default 10)
exit status 2
```

The Hasher CLI tool will make requests to the Hasher Service at random intervals with _somewhat_ randomized passwords. 

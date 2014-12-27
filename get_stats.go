package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

func statsForAllHosts(w http.ResponseWriter, r *http.Request) {
	c := redisPool.Get()
	defer c.Close()

	reply, err := redis.Values(c.Do("KEYS", "host:*"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "oops", http.StatusInternalServerError)
	}

	// Prepare map for final results
	allHostsResult := make(map[string]string)

	// Build the results map
	// TODO: Redis performance can probably be improved here with pipelining or other techniques
	for _, key := range reply {

		// data comes back from redis untyped, so we need to explictly cast
		keyBytes := key.([]byte)

		// extract host from key
		host := string(keyBytes[5:len(keyBytes)])

		// get total count for this host
		reply2, err := c.Do("GET", string(keyBytes))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			http.Error(w, "oops", http.StatusInternalServerError)
			return
		}

		// store in map
		allHostsResult[host] = string(reply2.([]byte))
	}

	js, err := json.Marshal(allHostsResult)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func bucketDataForHost(w http.ResponseWriter, r *http.Request) {
	c := redisPool.Get()
	defer c.Close()

	// Extract hostname from URL path
	host := r.URL.Path[12:len(r.URL.Path)]

	// Look up all minute buckets for this host
	reply, err := redis.Values(c.Do("KEYS", host+":*"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "oops", http.StatusInternalServerError)
		return
	}

	// Prepare map for final results
	values := make(map[string]string)

	// Build the results map
	// TODO: Redis performance can probably be improved here with pipelining or other techniques
	for _, key := range reply {

		// data comes back from redis untyped, so we need to explictly cast
		keyBytes := key.([]byte)

		// extract minute bucket from key
		minuteBucket := string(keyBytes[len(host)+1 : len(keyBytes)])

		// get count for this bucket
		reply2, err := c.Do("GET", string(keyBytes))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			http.Error(w, "oops", http.StatusInternalServerError)
			return
		}

		// store in map
		values[minuteBucket] = string(reply2.([]byte))
	}

	// send back data as JSON
	js, err := json.Marshal(values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

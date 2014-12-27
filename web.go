package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/bmizerany/lpx"
	"github.com/garyburd/redigo/redis"
	"github.com/kr/logfmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type routerLog struct {
	host string
}

var (
	redisPool redis.Pool
)

func (r *routerLog) HandleLogfmt(key, val []byte) error {
	if string(key) == "host" {
		r.host = string(val)
	}
	return nil
}

func main() {
	redisUrl, _ := url.Parse(os.Getenv("REDISCLOUD_URL"))

	redisPool = redis.Pool{
		MaxIdle:   50,
		MaxActive: 500, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisUrl.Host)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}

	defer redisPool.Close()

	http.HandleFunc("/log", checkAuth(os.Getenv("AUTH_SECRET"), processLogs))
	http.HandleFunc("/stats/hosts", checkAuth(os.Getenv("AUTH_SECRET"), statsForAllHosts))
	http.HandleFunc("/stats/host/", checkAuth(os.Getenv("AUTH_SECRET"), bucketDataForHost))
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func processLogs(w http.ResponseWriter, r *http.Request) {
	c := redisPool.Get()
	defer c.Close()

	lp := lpx.NewReader(bufio.NewReader(r.Body))
	for lp.Next() {
		if string(lp.Header().Procid) == "router" {
			rl := new(routerLog)
			if err := logfmt.Unmarshal(lp.Bytes(), rl); err != nil {
				// oops
			} else {
				timeBucket := timestamp2Bucket(lp.Header().Time)
				_, err := c.Do("INCR", fmt.Sprintf("%v:%v", rl.host, timeBucket))
				if err != nil {
					fmt.Printf("Error running INCR on Redis: %v\n", err)
				}
				_, err = c.Do("INCR", "host:"+rl.host)
				if err != nil {
					fmt.Printf("Error running INCR on Redis: %v\n", err)
				}
				fmt.Printf("%v @ %v: +1\n", rl.host, timeBucket)
			}
		}
	}
}

func statsForAllHosts(w http.ResponseWriter, r *http.Request) {
	c := redisPool.Get()
	defer c.Close()

	reply, err := redis.Values(c.Do("KEYS", "host:*"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "oops", http.StatusInternalServerError)
	}
	values := make([]string, len(reply))
	for i,key := range reply {
		values[i] = string(key.([]byte))
	}
	js, err := json.Marshal(values)
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
		// keys and values come back from redis untyped, so we need to explictily cast
		keyBytes := key.([]byte)

		// extract minute bucket from key
		minuteBucket := string(keyBytes[len(host)+1:len(keyBytes)])

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

	// send back result
	js, err := json.Marshal(values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)


}

func timestamp2Bucket(b []byte) int64 {
	t, err := time.Parse(time.RFC3339, string(b))
	if err != nil {
		fmt.Println(err)
	}
	return (t.Unix() / 60) * 60
}

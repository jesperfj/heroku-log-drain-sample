package main

import (
	"fmt"
	"bufio"
	"time"
	"net/http"
	"github.com/bmizerany/lpx"
	"github.com/kr/logfmt"
)

// This struct and the method below takes care of capturing the data we need
// from each log line. We pass it to Keith Rarick's logfmt parser and it
// handles parsing for us.
type routerLog struct {
	host string
}

func (r *routerLog) HandleLogfmt(key, val []byte) error {
	if string(key) == "host" {
		r.host = string(val)
	}
	return nil
}

// This is called every time we receive log lines from an app
func processLogs(w http.ResponseWriter, r *http.Request) {
	c := redisPool.Get()
	defer c.Close()

	lp := lpx.NewReader(bufio.NewReader(r.Body))
	// a single request may contain multiple log lines. Loop over each of them
	for lp.Next() {
		// we only care about logs from the heroku router
		if string(lp.Header().Procid) == "router" {
			rl := new(routerLog)
			if err := logfmt.Unmarshal(lp.Bytes(), rl); err != nil {
				fmt.Printf("Error parsing log line: %v\n", err)
			} else {
				timeBucket, err := timestamp2Bucket(lp.Header().Time)
				if err != nil {
					fmt.Printf("Error parsing time: %v", err)
					continue
				}
				_, err = c.Do("INCR", fmt.Sprintf("%v:%v", rl.host, timeBucket))
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

// Heroku log lines are formatted according to RFC5424 which is a subset
// of RFC3339 (RFC5424 is more restrictive).
// Reference: https://devcenter.heroku.com/articles/logging#log-format
func timestamp2Bucket(b []byte) (int64, error) {
	t, err := time.Parse(time.RFC3339, string(b))
	if err != nil {
		return 0, err
	}
	return (t.Unix() / 60) * 60, nil
}

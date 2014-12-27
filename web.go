package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"net/http"
	"net/url"
	"os"
)

var (
	redisPool redis.Pool
)

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

package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/soveran/redisurl"
	"net/http"
	"os"
)

var (
	redisPool redis.Pool
)

func main() {
	redisPool = redis.Pool{
		MaxIdle:   50,
		MaxActive: 500, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redisurl.ConnectToURL(os.Getenv("REDISCLOUD_URL"))
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}

	defer redisPool.Close()

	auth := checkAuth(os.Getenv("AUTH_SECRET"))

	http.HandleFunc("/log", auth, processLogs))
	http.HandleFunc("/stats/hosts", auth, statsForAllHosts))
	http.HandleFunc("/stats/host/", auth, bucketDataForHost))
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func checkAuth(correctPass string, pass http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		_, password, ok := r.BasicAuth()

		if !ok {
			http.Error(w, "authtorization required", http.StatusBadRequest)
			return
		}

		if password != correctPass {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		pass(w, r)
	}
}

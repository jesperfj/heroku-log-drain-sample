package main

import (
	"fmt"
	"net/http"
	"os"
	"bufio"
	"github.com/bmizerany/lpx"
)

func main() {
	http.HandleFunc("/", readMsg)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func readMsg(w http.ResponseWriter, r *http.Request) {
	lp := lpx.NewReader(bufio.NewReader(r.Body))
	for lp.Next() {
		if string(lp.Header().Name) == "router" {
			fmt.Println("Got router line with "+string(len(lp.Bytes()))+" bytes")
		}
	}
}

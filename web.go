package main

import (
	"bufio"
	"fmt"
	"github.com/bmizerany/lpx"
	"io/ioutil"
	"net/http"
	"os"
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
	if os.Getenv("DEBUG") == "true" {
		arr, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(arr))
	} else {
		lp := lpx.NewReader(bufio.NewReader(r.Body))
		for lp.Next() {
			fmt.Println(lp.Header().Name)
			if string(lp.Header().Name) == "router" {
				fmt.Println("Got router line with " + string(len(lp.Bytes())) + " bytes")
			}
		}
	}
}

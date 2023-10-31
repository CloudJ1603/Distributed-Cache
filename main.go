package main

import (
	"fmt"
	"distributed_cache/cacheFlex"
	"log"
	"net/http"
)

type server int

func (h *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	w.Write([]byte("Hello World!"))
}

func main() {
	var s server
	// http.ListenAndServe("localhost:8080", &s)
	if err := http.ListenAndServe("localhost:9999", &s); err != nil {
		log.Fatal("Server error:", err)
	}

}

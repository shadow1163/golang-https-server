package main

import (
	"flag"
	"net/http"

	"github.com/shadow1163/logger"
)

var (
	addr = flag.String("addr", ":8880", "http service address")
	log  = logger.NewLogger()
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Info(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "vuejsdemo.html")
}

func main() {
	flag.Parse()
	http.HandleFunc("/", serveHome)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Error("ListenAndServe: ", err)
	}
}

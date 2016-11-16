package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/srr/", srrHandler)
	if err := http.ListenAndServe(":20051", nil); err != nil {
		log.Println(err)
	}
}

func srrHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		handlePost(w, r)
	default:
		io.WriteString(w, "POSTのみ受け付けます")
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v\n", r)
	log.Println(r.Form)
	log.Println(r.Form["events"])
	for _, event := range r.Form["events"] {
		for _, eve := range event {
			log.Printf("%v\n", eve)
		}
	}

	// Message EventのTextとLocationを扱う。
	// Textはキーワードリターン用
	// Locationはストレージ保存用

	// のち、PUB/SUBにする。
}

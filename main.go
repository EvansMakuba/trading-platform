package main

import (
	"log"
	"net/http"
)

func main(){
	// serve the single page application (frontend)
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	// Example PI endpoint for your signals
	http.HandleFunc("/api/signals", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "success", "signal":"BUY", "asset": "Vol 25"}`))
	})

	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", nil)

	if err != nil{
		log.Fatal(err)
	}
}

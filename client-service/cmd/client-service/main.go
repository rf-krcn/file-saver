package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = 80

func main() {

	staticDir := "/build"

	fileServer := http.FileServer(http.Dir(staticDir))

	http.Handle("/", http.StripPrefix("/", fileServer))

	log.Printf("Server started on :%d...\n", port)
	err := http.ListenAndServe(":"+fmt.Sprint(port), nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}

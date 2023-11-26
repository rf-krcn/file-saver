package main

import (
	"fmt"
	"log"
	"net/http"

	web "github.com/AbderraoufKhorchani/file-saver/broker-service/web"
)

const webPort = "80"

func main() {

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: web.Routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}

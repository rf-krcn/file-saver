package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AbderraoufKhorchani/file-saver/broker-service/cmd/api"
)

const webPort = "80"

func main() {

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: api.Routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AbderraoufKhorchani/file-saver/auth-service/cmd/api"
	"github.com/AbderraoufKhorchani/file-saver/auth-service/data"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const webPort = "80"

var counts int64

func main() {
	log.Println("Starting authentication service")

	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	data.New(conn)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: api.Routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
func connectToDB() *gorm.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds....")
		time.Sleep(2 * time.Second)
		continue
	}
}

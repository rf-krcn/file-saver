package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/AbderraoufKhorchani/file-saver/file-service/cmd/api"
	"github.com/AbderraoufKhorchani/file-saver/file-service/data"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var counts int64

func main() {

	log.Println("Starting files saving service")

	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	data.New(conn)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	fileService := &api.FileService{} // Use a pointer to the service instance
	srv := grpc.NewServer()

	// Register your service with the gRPC server
	api.RegisterFileServiceServer(srv, fileService)

	log.Println("File Service listening on :50051")

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
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

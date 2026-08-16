package main

import (
	"log"
	"net/http"
	"os"

	"pumpstation/internal/pumpstation"
)

func main() {
	address := os.Getenv("PUMPSTATION_ADDR")
	if address == "" {
		address = ":8080"
	}
	service := pumpstation.NewFixtureService()
	server := &http.Server{Addr: address, Handler: pumpstation.NewHTTPHandler(service)}
	log.Printf("水厂泵站设备档案服务已启动: %s", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

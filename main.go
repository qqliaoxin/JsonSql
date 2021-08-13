package main

import (
	"log"
	"net/http"

	"github.com/qqliaoxin/jsonsql/handler"
	"github.com/qqliaoxin/jsonsql/logger"
)

func main() {
	http.HandleFunc("/get", handler.GetHandler)
	addr := ":8080"
	logger.SetLevel(logger.DEBUG)
	logger.Info("server listen on " + addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

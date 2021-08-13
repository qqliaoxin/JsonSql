package main

import (
	"log"
	"net/http"

	"jsonSql/handler"
	"jsonSql/logger"
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

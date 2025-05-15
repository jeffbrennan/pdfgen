package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jeffbrennan/pdfgen/internal/server"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/generate-pdf", server.GeneratePDFHandler).Methods("POST")
	r.HandleFunc("/stream-logs", server.StreamLogsHandler)

	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/").Handler(fs)

	fmt.Println("Starting server on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

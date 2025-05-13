package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gorilla/mux"
)

var (
	logChannel = make(chan string)
	logMu      sync.Mutex
)

func publishLog(message string) {
	logMu.Lock()
	defer logMu.Unlock()
	select {
	case logChannel <- message:
	default:
	}
}

func streamLogsHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		case msg := <-logChannel:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func generatePDFHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	response, err := pdfgen(url)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		http.Error(w, fmt.Sprintf("PDF generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	fileName := filepath.Base(response.pdfPath)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.WriteHeader(http.StatusOK)
	w.Write(response.pdfBytes)

	cleanupRepo(response.parts)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/generate-pdf", generatePDFHandler).Methods("POST")
	r.HandleFunc("/stream-logs", streamLogsHandler)

	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/").Handler(fs)

	fmt.Println("Starting server on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

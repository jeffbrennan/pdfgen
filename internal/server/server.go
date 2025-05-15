package server

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/jeffbrennan/pdfgen/internal/generators"
	"github.com/jeffbrennan/pdfgen/internal/logging"
	"github.com/jeffbrennan/pdfgen/internal/utils"
)

func StreamLogsHandler(w http.ResponseWriter, r *http.Request) {
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
		case msg := <-logging.LogChannel:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func GeneratePDFHandler(w http.ResponseWriter, r *http.Request) {
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

	response, err := generators.HandlePdfGeneration(url)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		http.Error(w, fmt.Sprintf("PDF generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	fileName := filepath.Base(response.PdfPath)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.WriteHeader(http.StatusOK)
	w.Write(response.PdfBytes)

	utils.CleanupDir(response.Parts)
}

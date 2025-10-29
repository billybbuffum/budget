package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/billybbuffum/budget/internal/application"
)

type ImportHandler struct {
	importService *application.ImportService
}

func NewImportHandler(importService *application.ImportService) *ImportHandler {
	return &ImportHandler{importService: importService}
}

const (
	maxUploadSize = 10 << 20 // 10 MB
)

// ImportTransactions handles OFX/QFX file upload and import
func (h *ImportHandler) ImportTransactions(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with size limit
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large (max 10MB)", http.StatusBadRequest)
		return
	}

	// Get account_id from form
	accountID := r.FormValue("account_id")
	if accountID == "" {
		http.Error(w, "account_id is required", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".ofx" && ext != ".qfx" {
		http.Error(w, "invalid file type, must be .ofx or .qfx", http.StatusBadRequest)
		return
	}

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file content", http.StatusInternalServerError)
		return
	}

	// Reset file reader for parsing
	reader := strings.NewReader(string(fileContent))

	// Validate OFX file
	if err := h.importService.ValidateOFXFile(reader); err != nil {
		http.Error(w, fmt.Sprintf("invalid OFX file: %v", err), http.StatusBadRequest)
		return
	}

	// Reset reader for import
	reader.Seek(0, io.SeekStart)

	// Import transactions
	result, err := h.importService.ImportFromOFX(r.Context(), accountID, reader)
	if err != nil {
		http.Error(w, fmt.Sprintf("import failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return import result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

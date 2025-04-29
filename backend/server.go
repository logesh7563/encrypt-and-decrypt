package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

const (
	// HTTPPort is the port for the HTTP server
	HTTPPort = "8083"

	// MaxUploadSize is the maximum size of uploaded images (10MB)
	MaxUploadSize = 10 * 1024 * 1024

	// UploadPath is the directory for storing uploaded images
	UploadPath = "./assets/uploads"

	// ProcessedPath is the directory for storing processed images
	ProcessedPath = "./assets/processed"
)

// ImageProcessingRequest represents the request body for image processing
type ImageProcessingRequest struct {
	Operation string  `json:"operation"`
	Angle     float64 `json:"angle,omitempty"`
	Radius    float64 `json:"radius,omitempty"`
	Key       string  `json:"key"`
}

// ImageResponse represents the response for image operations
type ImageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// StartServer initializes and starts the HTTP server
func StartServer() {
	// Create upload and processed directories if they don't exist
	if err := os.MkdirAll(UploadPath, 0755); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}
	if err := os.MkdirAll(ProcessedPath, 0755); err != nil {
		log.Fatal("Failed to create processed directory:", err)
	}

	// Create router
	router := mux.NewRouter()

	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/upload", handleUpload).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/process", handleProcess).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/download", handleDownload).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/transmit", handleTransmit).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/decrypt", handleDecrypt).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/encrypt", handleEncrypt).Methods("POST", "OPTIONS")

	// Serve static files from the frontend directory
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../frontend")))

	// Add middleware for CORS and logging
	handler := corsMiddleware(loggingMiddleware(router))

	// Start TCP server in a goroutine
	go StartTCPServer()

	// Start HTTP server
	log.Printf("HTTP server starting on port %s", HTTPPort)
	if err := http.ListenAndServe(":"+HTTPPort, handler); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only add CORS headers for API routes
		if strings.HasPrefix(r.URL.Path, "/api") {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// handleUpload handles image upload requests
func handleUpload(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error getting file from form: %v", err)
		http.Error(w, "Error getting file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Printf("Error creating upload directory: %v", err)
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	// Create a new file in the upload directory
	filename := filepath.Join("uploads", handler.Filename)
	dst, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the created file
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error copying file: %v", err)
		http.Error(w, "Error copying file", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "File uploaded successfully",
		"data":    handler.Filename,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleProcess handles image processing requests
func handleProcess(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get filename from query parameter
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Get operation from request body
	var req struct {
		Operation string `json:"operation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Open the uploaded image
	filepath := "uploads/" + filename
	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		http.Error(w, "Failed to open image", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		http.Error(w, "Failed to decode image", http.StatusInternalServerError)
		return
	}

	// Process the image based on the operation
	var processedImg image.Image
	switch req.Operation {
	case "grayscale":
		processedImg = ConvertToGrayscale(img)
	case "flip":
		processedImg = FlipVertical(img)
	case "rotate":
		processedImg = RotateArbitrary(img, 90) // Default 90-degree rotation
	case "blur":
		processedImg = ApplyGaussianBlur(img, 2.0) // Default blur radius
	default:
		http.Error(w, "Invalid operation", http.StatusBadRequest)
		return
	}

	// Create processed directory if it doesn't exist
	if err := os.MkdirAll("processed", 0755); err != nil {
		log.Printf("Error creating processed directory: %v", err)
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	// Save the processed image
	processedFilename := "processed_" + filename
	processedPath := "processed/" + processedFilename
	processedFile, err := os.Create(processedPath)
	if err != nil {
		log.Printf("Error creating processed file: %v", err)
		http.Error(w, "Error creating processed file", http.StatusInternalServerError)
		return
	}
	defer processedFile.Close()

	// Encode the processed image
	if err := jpeg.Encode(processedFile, processedImg, nil); err != nil {
		log.Printf("Error encoding processed image: %v", err)
		http.Error(w, "Error encoding processed image", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "Image processed successfully",
		"data":    processedFilename,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// padKey pads the key to the required length (32 bytes for AES-256)
func padKey(key string) []byte {
	// Use SHA-256 for consistent key derivation
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return hasher.Sum(nil)
}

// handleDownload handles image download requests
func handleDownload(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get filename and key from query parameters
	filename := r.URL.Query().Get("data")
	key := r.URL.Query().Get("key")

	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Check if the file exists in the processed directory
	filepath := "processed/" + filename
	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		http.Error(w, "Failed to open image", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}

	// If a key is provided, encrypt the content
	if key != "" {
		// Pad the key to the required length
		paddedKey := padKey(key)

		// Convert the image data to base64
		base64Data := base64.StdEncoding.EncodeToString(fileContent)

		// Encrypt the base64 data
		encryptedData, err := EncryptToBase64([]byte(base64Data), string(paddedKey))
		if err != nil {
			log.Printf("Error encrypting file: %v", err)
			http.Error(w, "Failed to encrypt image", http.StatusInternalServerError)
			return
		}

		// Set the encrypted data as the response
		fileContent = []byte(encryptedData)

		// Set appropriate headers for encrypted data
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", "attachment; filename=encrypted_"+filename)
	} else {
		// Set headers for regular image download
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	}

	// Write the content to the response
	if _, err := w.Write(fileContent); err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Failed to send image", http.StatusInternalServerError)
		return
	}
}

// handleTransmit handles image transmission requests
func handleTransmit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EncryptedData string `json:"encryptedData"`
		ServerAddr    string `json:"serverAddr"`
		ImageID       string `json:"imageID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode encrypted data
	encryptedData, err := base64.StdEncoding.DecodeString(req.EncryptedData)
	if err != nil {
		sendError(w, "Invalid encrypted data", http.StatusBadRequest)
		return
	}

	// Send image via TCP
	err = SendImageViaTCP(req.ImageID, encryptedData, req.ServerAddr)
	if err != nil {
		sendError(w, "Failed to transmit image", http.StatusInternalServerError)
		return
	}

	sendJSON(w, ImageResponse{
		Success: true,
		Message: "Image transmitted successfully",
	})
}

// handleDecrypt handles decryption of encrypted files
func handleDecrypt(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for browser compatibility
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse the multipart form
	err := r.ParseMultipartForm(MaxUploadSize)
	if err != nil {
		sendError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the file and key from the form
	file, header, err := r.FormFile("file")
	if err != nil {
		sendError(w, "No file received: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	key := r.FormValue("key")
	if key == "" {
		sendError(w, "Decryption key is required", http.StatusBadRequest)
		return
	}

	// Check file extension and provide a warning but continue
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".enc") {
		log.Printf("Warning: File %s doesn't have .enc extension", header.Filename)
	}

	// Read the encrypted file
	encryptedData, err := io.ReadAll(file)
	if err != nil {
		sendError(w, "Failed to read encrypted file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Received file: %s, size: %d bytes, attempting to decrypt", header.Filename, len(encryptedData))

	// Try to decrypt the data
	decryptedData, err := DecryptData(encryptedData, key)
	if err != nil {
		log.Printf("Decryption error: %v", err)
		sendError(w, "Failed to decrypt data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Decryption successful, decrypted size: %d bytes", len(decryptedData))

	// Try to determine the content type
	contentType := http.DetectContentType(decryptedData)
	log.Printf("Detected content type: %s", contentType)

	// Set the appropriate content type and write the decrypted data
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(decryptedData)))
	w.WriteHeader(http.StatusOK)
	w.Write(decryptedData)
}

// handleEncrypt handles encryption of files
func handleEncrypt(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse multipart form with the defined max size
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		http.Error(w, fmt.Sprintf("Could not parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Get the file from form data
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get the encryption key
	key := r.FormValue("key")
	if key == "" {
		http.Error(w, "No encryption key provided", http.StatusBadRequest)
		return
	}

	// Read the file
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file: %v", err), http.StatusInternalServerError)
		return
	}

	// Log the size of data being encrypted for debugging
	log.Printf("Encrypting file: %s, size: %d bytes", handler.Filename, len(fileData))

	// Encrypt the data using our secure EncryptData function
	encryptedData, err := EncryptData(fileData, key)
	if err != nil {
		http.Error(w, fmt.Sprintf("Encryption failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Encryption successful. Encrypted size: %d bytes", len(encryptedData))

	// Set headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.enc", handler.Filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(encryptedData)))

	// Write encrypted data to response
	if _, err := w.Write(encryptedData); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Helper function to check if a string is base64 encoded
func isBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// sendError sends an error response
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ImageResponse{
		Success: false,
		Message: message,
	})
}

// sendJSON sends a JSON response
func sendJSON(w http.ResponseWriter, response ImageResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	StartServer()
}

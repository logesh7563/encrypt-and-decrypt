package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	// TCPPort is the port for the TCP server
	TCPPort = "8084"

	// Message types
	ImageDataRequest    = byte(1) // Request for an image
	ImageDataResponse   = byte(2) // Response with image data
	ImageDataTransfer   = byte(3) // Sending image data to server for storage
	ConfirmationMessage = byte(4) // Confirmation of receipt
)

var (
	// Storage for the last encrypted image data with mutex for concurrent access
	encryptedImageStore      = make(map[string][]byte)
	encryptedImageStoreMutex sync.RWMutex
)

// StartTCPServer starts the TCP server for image transmission
func StartTCPServer() {
	listener, err := net.Listen("tcp", ":"+TCPPort)
	if err != nil {
		log.Fatal("Failed to start TCP server:", err)
	}
	defer listener.Close()

	log.Printf("TCP server started on port %s", TCPPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

// handleConnection handles incoming TCP connections
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read image data
	buffer := make([]byte, 1024*1024) // 1MB buffer
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println("Failed to read data:", err)
		return
	}

	// Process received data
	data := buffer[:n]
	log.Printf("Received %d bytes of data", len(data))
}

// handleImageRequest sends requested encrypted image back to client
func handleImageRequest(conn net.Conn) {
	// Read image ID (fixed-length for simplicity)
	idLenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, idLenBuf); err != nil {
		log.Printf("Error reading ID length: %v", err)
		return
	}

	idLen := binary.BigEndian.Uint32(idLenBuf)
	idBuf := make([]byte, idLen)
	if _, err := io.ReadFull(conn, idBuf); err != nil {
		log.Printf("Error reading image ID: %v", err)
		return
	}

	imageID := string(idBuf)

	// Read the image from storage
	encryptedImageStoreMutex.RLock()
	imageData, exists := encryptedImageStore[imageID]
	encryptedImageStoreMutex.RUnlock()

	// Prepare header for response
	header := []byte{ImageDataResponse}

	if !exists {
		// Send empty data if image doesn't exist
		sizeBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(sizeBytes, 0)

		response := append(header, sizeBytes...)
		if _, err := conn.Write(response); err != nil {
			log.Printf("Error sending empty response: %v", err)
		}
		return
	}

	// Send the image data
	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(len(imageData)))

	// Combine header, size, and data
	response := append(header, sizeBytes...)
	response = append(response, imageData...)

	if _, err := conn.Write(response); err != nil {
		log.Printf("Error sending image data: %v", err)
		return
	}

	log.Printf("Sent image '%s' (%d bytes) to client", imageID, len(imageData))
}

// handleImageTransfer receives and stores encrypted image from client
func handleImageTransfer(conn net.Conn) {
	// Read image ID length
	idLenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, idLenBuf); err != nil {
		log.Printf("Error reading ID length: %v", err)
		return
	}

	idLen := binary.BigEndian.Uint32(idLenBuf)
	idBuf := make([]byte, idLen)
	if _, err := io.ReadFull(conn, idBuf); err != nil {
		log.Printf("Error reading image ID: %v", err)
		return
	}

	imageID := string(idBuf)

	// Read image data size
	sizeBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, sizeBuf); err != nil {
		log.Printf("Error reading data size: %v", err)
		return
	}

	dataSize := binary.BigEndian.Uint32(sizeBuf)

	// Read the encrypted image data
	data := make([]byte, dataSize)
	if _, err := io.ReadFull(conn, data); err != nil {
		log.Printf("Error reading image data: %v", err)
		return
	}

	// Store the image data
	encryptedImageStoreMutex.Lock()
	encryptedImageStore[imageID] = data
	encryptedImageStoreMutex.Unlock()

	// Send confirmation
	conn.Write([]byte{ConfirmationMessage})

	log.Printf("Received and stored encrypted image '%s' (%d bytes)", imageID, dataSize)
}

// SendImageViaTCP sends an encrypted image to a TCP server
func SendImageViaTCP(imageID string, encryptedData []byte, serverAddr string) error {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send image data
	_, err = conn.Write(encryptedData)
	return err
}

// RequestImageViaTCP requests and receives an encrypted image from a TCP server
func RequestImageViaTCP(imageID string, serverAddr string, outputPath string) error {
	// Connect to the server
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Set timeouts
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// Create request message
	var buf bytes.Buffer

	// Add message type
	buf.WriteByte(ImageDataRequest)

	// Add image ID length and ID
	idBytes := []byte(imageID)
	idLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idLenBytes, uint32(len(idBytes)))
	buf.Write(idLenBytes)
	buf.Write(idBytes)

	// Send the request
	if _, err := conn.Write(buf.Bytes()); err != nil {
		return err
	}

	// Read response type
	responseBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, responseBuf); err != nil {
		return err
	}

	if responseBuf[0] != ImageDataResponse {
		return err
	}

	// Read data size
	sizeBytes := make([]byte, 4)
	if _, err := io.ReadFull(conn, sizeBytes); err != nil {
		return err
	}

	dataSize := binary.BigEndian.Uint32(sizeBytes)
	if dataSize == 0 {
		return err
	}

	// Read the image data
	data := make([]byte, dataSize)
	if _, err := io.ReadFull(conn, data); err != nil {
		return err
	}

	// Save to file if outputPath is specified
	if outputPath != "" {
		return os.WriteFile(outputPath, data, 0644)
	}

	return nil
}

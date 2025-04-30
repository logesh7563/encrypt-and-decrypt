package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
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

	log.Printf("New connection from %s", conn.RemoteAddr().String())

	// Set a reasonable timeout
	conn.SetDeadline(time.Now().Add(1 * time.Minute))

	// Read message type first
	msgTypeBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, msgTypeBuf); err != nil {
		log.Println("Failed to read message type:", err)
		return
	}

	// Handle the message based on its type
	switch msgTypeBuf[0] {
	case ImageDataRequest:
		log.Printf("Received image request from %s", conn.RemoteAddr().String())
		handleImageRequest(conn)

	case ImageDataTransfer:
		log.Printf("Received image transfer from %s", conn.RemoteAddr().String())
		handleImageTransfer(conn)

	default:
		log.Printf("Unknown message type %d from %s", msgTypeBuf[0], conn.RemoteAddr().String())
	}
}

// handleImageRequest sends requested encrypted image back to client
func handleImageRequest(conn net.Conn) error {
	// Read image ID length (4 bytes)
	idLenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, idLenBuf); err != nil {
		return fmt.Errorf("failed to read ID length: %v", err)
	}
	idLen := binary.BigEndian.Uint32(idLenBuf)

	// Read image ID
	idBuf := make([]byte, idLen)
	if _, err := io.ReadFull(conn, idBuf); err != nil {
		return fmt.Errorf("failed to read image ID: %v", err)
	}
	imageID := string(idBuf)

	// Retrieve the encrypted image data from storage
	encryptedImageStoreMutex.RLock()
	imageData, exists := encryptedImageStore[imageID]
	encryptedImageStoreMutex.RUnlock()

	if !exists {
		return fmt.Errorf("image with ID %s not found", imageID)
	}

	// Send response message type
	if _, err := conn.Write([]byte{ImageDataResponse}); err != nil {
		return fmt.Errorf("failed to send response type: %v", err)
	}

	// Send data length
	dataLenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(dataLenBuf, uint32(len(imageData)))
	if _, err := conn.Write(dataLenBuf); err != nil {
		return fmt.Errorf("failed to send data length: %v", err)
	}

	// Send image data
	if _, err := conn.Write(imageData); err != nil {
		return fmt.Errorf("failed to send image data: %v", err)
	}

	return nil
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

	// Create buffer for the complete message
	var buf bytes.Buffer

	// Add message type
	buf.WriteByte(ImageDataTransfer)

	// Add image ID length and ID
	idBytes := []byte(imageID)
	idLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idLenBytes, uint32(len(idBytes)))
	buf.Write(idLenBytes)
	buf.Write(idBytes)

	// Add data size and data
	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(len(encryptedData)))
	buf.Write(sizeBytes)
	buf.Write(encryptedData)

	// Send the complete message
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	// Wait for confirmation
	confirmBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, confirmBuf); err != nil {
		return err
	}

	if confirmBuf[0] != ConfirmationMessage {
		return err
	}

	return nil
}

// RequestImageViaTCP requests and receives an encrypted image from a TCP server
func RequestImageViaTCP(serverAddr, imageID string) ([]byte, error) {
	log.Printf("RequestImageViaTCP: Requesting image '%s' from %s", imageID, serverAddr)

	// Connect to the TCP server
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Set a reasonable timeout
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// Prepare image ID data
	idBytes := []byte(imageID)
	idLen := len(idBytes)
	log.Printf("RequestImageViaTCP: Image ID length: %d bytes", idLen)

	// Send message type
	if _, err := conn.Write([]byte{ImageDataRequest}); err != nil {
		return nil, fmt.Errorf("failed to send message type: %v", err)
	}

	// Send image ID length (4 bytes)
	idLenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(idLenBuf, uint32(idLen))
	if _, err := conn.Write(idLenBuf); err != nil {
		return nil, fmt.Errorf("failed to send ID length: %v", err)
	}

	// Send image ID
	if _, err := conn.Write(idBytes); err != nil {
		return nil, fmt.Errorf("failed to send image ID: %v", err)
	}

	// Read the response message type
	msgTypeBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, msgTypeBuf); err != nil {
		return nil, fmt.Errorf("failed to read response type: %v", err)
	}

	if msgTypeBuf[0] != ImageDataResponse {
		return nil, fmt.Errorf("unexpected response type: %d", msgTypeBuf[0])
	}

	// Read data length (4 bytes)
	dataLenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, dataLenBuf); err != nil {
		return nil, fmt.Errorf("failed to read data length: %v", err)
	}
	dataLen := binary.BigEndian.Uint32(dataLenBuf)
	log.Printf("RequestImageViaTCP: Receiving %d bytes of image data", dataLen)

	// Validate data length to prevent potential issues
	if dataLen == 0 {
		return nil, fmt.Errorf("received zero-length data")
	}

	if dataLen > 100*1024*1024 { // 100 MB limit
		return nil, fmt.Errorf("data length too large: %d bytes", dataLen)
	}

	// Read image data
	imageData := make([]byte, dataLen)
	bytesRead, err := io.ReadFull(conn, imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %v (read %d of %d bytes)",
			err, bytesRead, dataLen)
	}

	log.Printf("RequestImageViaTCP: Successfully received %d bytes", bytesRead)

	return imageData, nil
}

# Secure Image Processing and Transmission Web App

A web-based platform for secure image processing, encryption, and transmission.

## Features

- Upload images for processing
- Apply various image processing operations:
  - Flip vertically
  - Rotate by arbitrary angle
  - Rotate using three shear matrices
  - Convert to grayscale
  - Apply box blur
  - Apply Gaussian blur
  - Edge detection using Sobel operator
- Encrypt processed images using AES-256
- Download or transmit encrypted images securely
- Support for TCP and gRPC transmission

## Project Structure

```
- /frontend - Web interface files
- /backend - Go server implementation
- /assets - Temporary storage for uploaded and processed images
```

## Requirements

- Go 1.21+
- Modern web browser (Chrome, Firefox, Edge, etc.)

## Setup Instructions

### Install dependencies and run the server

```bash
# Initialize Go modules
go mod tidy

# Run the server
go run backend/server.go
```

### Accessing the Web Interface

After starting the server, open your browser and navigate to:

```
http://localhost:8080
```

## Usage Guide

1. Open the web interface
2. Upload an image
3. Select the desired image processing operation
4. Enter an encryption key (this will be required for decryption)
5. Process and encrypt the image
6. Download the encrypted image or transmit it via TCP/gRPC

## Security Notes

- All image encryption uses AES-256 in GCM mode
- Encryption keys should be kept secure
- Transmitted images are encrypted by default
- Processed images are stored temporarily and then deleted 
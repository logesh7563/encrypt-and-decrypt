package main

import (
	"log"
	"path/filepath"
	"strings"
)

// Extract name and extension
ext := filepath.Ext(filename)
if ext == "" {
	log.Printf("Warning: No file extension found in filename: %s", filename)
}
nameOnly := strings.TrimSuffix(filename, ext) 
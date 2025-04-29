// Global variables to store image data
let originalImageData = null;
let processedImageData = null;

/**
 * Handles image upload from the file input
 */
async function handleUpload() {
    const fileInput = document.getElementById('imageInput');
    const statusElement = document.getElementById('uploadStatus');
    
    try {
        // Check if a file was selected
        if (!fileInput.files || fileInput.files.length === 0) {
            statusElement.textContent = "Please select an image file first.";
            statusElement.className = "status error";
            return;
        }

        statusElement.textContent = "Loading image...";
        statusElement.className = "status loading";
        
        const file = fileInput.files[0];
        
        // Verify it's an image
        if (!file.type.startsWith('image/')) {
            statusElement.textContent = "Please select a valid image file.";
            statusElement.className = "status error";
            return;
        }
        
        // Read the file as a data URL
        const reader = new FileReader();
        
        // Create a promise to handle the file reading
        const imageLoaded = new Promise((resolve, reject) => {
            reader.onload = (event) => resolve(event.target.result);
            reader.onerror = (error) => reject(error);
        });
        
        reader.readAsDataURL(file);
        
        // Wait for the image to load
        originalImageData = await imageLoaded;
        
        // Display the original image
        const originalImage = document.getElementById('originalImage');
        originalImage.src = originalImageData;
        
        statusElement.textContent = "Image uploaded successfully!";
        statusElement.className = "status success";
    } catch (error) {
        console.error("Upload error:", error);
        statusElement.textContent = `Upload failed: ${error.message}`;
        statusElement.className = "status error";
    }
}

/**
 * Processes the image with the selected operation
 */
async function handleProcess() {
    const statusElement = document.getElementById('processingStatus');
    if (!statusElement) {
        console.error("Missing element: processingStatus");
        return;
    }
    const operationSelect = document.getElementById('operationSelect');
    const encryptionKey = document.getElementById('encryptionKey').value;
    
    try {
        // Check if an image has been uploaded
        if (!originalImageData) {
            statusElement.textContent = "Please upload an image first.";
            statusElement.className = "status error";
            return;
        }
        
        // Check if encryption key is provided
        if (!encryptionKey) {
            statusElement.textContent = "Please enter an encryption key.";
            statusElement.className = "status error";
            return;
        }
        
        statusElement.textContent = "Processing image...";
        statusElement.className = "status loading";
        
        const operation = operationSelect.value;
        
        // Process the image on the client side (in a real app, this would be sent to the server)
        const processedImage = await processImage(originalImageData, operation);
        
        // Display the processed image
        document.getElementById('processedImage').src = processedImage;
        
        // Store the processed data
        processedImageData = processedImage;
        
        statusElement.textContent = "Image processed successfully!";
        statusElement.className = "status success";
    } catch (error) {
        console.error("Processing error:", error);
        statusElement.textContent = `Processing failed: ${error.message}`;
        statusElement.className = "status error";
    }
}

/**
 * Client-side image processing (simplified for demo)
 * In a real app, this would call the server for processing
 */
async function processImage(imageData, operation) {
    return new Promise((resolve) => {
        const img = new Image();
        img.onload = () => {
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            
            // Set canvas dimensions
            canvas.width = img.width;
            canvas.height = img.height;
            
            // Apply the selected operation
            switch (operation) {
                case 'flip':
                    ctx.translate(0, canvas.height);
                    ctx.scale(1, -1);
                    break;
                    
                case 'rotate':
                    ctx.translate(canvas.width/2, canvas.height/2);
                    ctx.rotate(Math.PI/2); // 90 degrees
                    ctx.translate(-canvas.height/2, -canvas.width/2);
                    canvas.width = img.height;
                    canvas.height = img.width;
                    break;
                    
                case 'grayscale':
                    ctx.filter = 'grayscale(100%)';
                    break;
                    
                case 'blur':
                    ctx.filter = 'blur(5px)';
                    break;
            }
            
            // Draw image with the applied filters/transformations
            if (operation === 'rotate') {
                ctx.drawImage(img, 0, 0, img.height, img.width);
            } else {
                ctx.drawImage(img, 0, 0, img.width, img.height);
            }
            
            // Reset filters
            ctx.filter = 'none';
            
            // Return the processed image
            resolve(canvas.toDataURL('image/png'));
        };
        
        img.src = imageData;
    });
}

/**
 * Handles download of the encrypted image
 */
async function handleDownload() {
    const statusElement = document.getElementById('processingStatus');
    const encryptionKey = document.getElementById('encryptionKey').value;
    
    try {
        if (!processedImageData) {
            statusElement.textContent = "Please process an image first.";
            statusElement.className = "status error";
            return;
        }
        
        if (!encryptionKey) {
            statusElement.textContent = "Please enter an encryption key.";
            statusElement.className = "status error";
            return;
        }
        
        statusElement.textContent = "Encrypting and preparing download...";
        statusElement.className = "status loading";
        
        // Convert base64 data URL to blob
        const response = await fetch(processedImageData);
        const blob = await response.blob();
        
        // Create form data
        const formData = new FormData();
        formData.append('file', blob, 'image.png');
        formData.append('key', encryptionKey);
        
        // Send to server for encryption - USING THE ENCRYPT ENDPOINT DIRECTLY
        const encryptResponse = await fetch('http://localhost:8085/api/encrypt', {
            method: 'POST',
            body: formData
        });
        
        if (!encryptResponse.ok) {
            const errorText = await encryptResponse.text();
            throw new Error(`Server error: ${errorText || encryptResponse.status}`);
        }
        
        // Get the encrypted data as a blob
        const encryptedBlob = await encryptResponse.blob();
        
        // Create a download link
        const downloadUrl = URL.createObjectURL(encryptedBlob);
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = 'encrypted_image.enc';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        // Clean up the URL object
        URL.revokeObjectURL(downloadUrl);
        
        statusElement.textContent = "Image encrypted and downloaded successfully!";
        statusElement.className = "status success";
    } catch (error) {
        console.error("Download error:", error);
        let errorMessage = `Download failed: ${error.message}`;
        if (error.name === 'TypeError' && error.message.includes('Failed to fetch')) {
            errorMessage += '\nPlease check if the server is running and accessible.';
        }
        statusElement.textContent = errorMessage;
        statusElement.className = "status error";
    }
}

/**
 * Handles transmission of the encrypted image to a server
 */
async function handleTransmit() {
    const statusElement = document.getElementById('transmissionStatus');
    const serverAddress = document.getElementById('serverAddress').value || 'localhost:8084';
    const imageId = document.getElementById('imageId').value || `img_${Date.now()}`;
    const encryptionKey = document.getElementById('encryptionKey').value;
    
    try {
        if (!processedImageData) {
            statusElement.textContent = "Please process an image first.";
            statusElement.className = "status error";
            return;
        }
        
        if (!encryptionKey) {
            statusElement.textContent = "Please enter an encryption key.";
            statusElement.className = "status error";
            return;
        }
        
        statusElement.textContent = "Transmitting image...";
        statusElement.className = "status loading";
        
        // Create the protocol (http or https)
        const protocol = window.location.protocol === 'https:' ? 'https' : 'http';
        const serverUrl = `${protocol}://${serverAddress}/api/upload`;
        
        // Convert base64 data URL to blob
        const response = await fetch(processedImageData);
        const blob = await response.blob();
        
        // Create form data
        const formData = new FormData();
        formData.append('image', blob, 'image.png');
        formData.append('key', encryptionKey);
        formData.append('id', imageId);
        
        // Send to server
        const uploadResponse = await fetch(serverUrl, {
            method: 'POST',
            body: formData
        });
        
        if (!uploadResponse.ok) {
            throw new Error(`HTTP error! status: ${uploadResponse.status}`);
        }
        
        const result = await uploadResponse.json();
        
        statusElement.textContent = `Image transmitted successfully! ID: ${result.id || imageId}`;
        statusElement.className = "status success";
    } catch (error) {
        console.error("Transmission error:", error);
        statusElement.textContent = `Transmission failed: ${error.message}`;
        statusElement.className = "status error";
    }
}

document.addEventListener('DOMContentLoaded', function() {
    const processBtn = document.getElementById('processBtn');
    if (processBtn) {
        processBtn.addEventListener('click', handleProcess);
    }
    // Other event bindings...
});
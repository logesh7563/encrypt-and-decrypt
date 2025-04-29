// Global variables to store image data
let originalImageData = null;
let processedImageData = null;

/**
 * Handles image upload from the file input
 */
async function handleUpload() {
    // Try to get the file input from either encrypt or decrypt page
    const fileInput = document.getElementById('imageInput') || document.getElementById('encryptedImageInput');
    const statusElement = document.getElementById('uploadStatus');
    const imageElement = document.getElementById('originalImage') || document.getElementById('encryptedImage');
    
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
        
        // Verify it's an image or .enc file
        if (!file.type.startsWith('image/') && !file.name.endsWith('.enc')) {
            statusElement.textContent = "Please select a valid image or .enc file.";
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
        if (imageElement) {
            imageElement.src = originalImageData;
        }
        
        statusElement.textContent = "File uploaded successfully!";
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
        const encryptResponse = await fetch('http://localhost:8083/api/encrypt', {
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
        statusElement.textContent = `Download failed: ${error.message}`;
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

/**
 * Handles decryption of the uploaded image
 */
async function handleDecrypt() {
    const statusElement = document.getElementById('decryptionStatus');
    const decryptionKey = document.getElementById('decryptionKey').value;
    const sourceSelect = document.getElementById('sourceSelect');
    
    try {
        if (!originalImageData) {
            statusElement.textContent = "Please upload an image first.";
            statusElement.className = "status error";
            return;
        }
        
        if (!decryptionKey) {
            statusElement.textContent = "Please enter a decryption key.";
            statusElement.className = "status error";
            return;
        }
        
        statusElement.textContent = "Decrypting image...";
        statusElement.className = "status loading";
        
        // Convert base64 data URL to blob
        const response = await fetch(originalImageData);
        const blob = await response.blob();
        
        // Create form data
        const formData = new FormData();
        formData.append('file', blob, 'image.enc');
        formData.append('key', decryptionKey);
        
        // Send to server for decryption
        const decryptResponse = await fetch('http://localhost:8083/api/decrypt', {
            method: 'POST',
            body: formData
        });
        
        if (!decryptResponse.ok) {
            throw new Error(`HTTP error! status: ${decryptResponse.status}`);
        }
        
        // Get the decrypted image as a blob
        const decryptedBlob = await decryptResponse.blob();
        const decryptedUrl = URL.createObjectURL(decryptedBlob);
        
        // Display the decrypted image
        const decryptedImage = document.getElementById('decryptedImage');
        if (decryptedImage) {
            decryptedImage.src = decryptedUrl;
        }
        
        statusElement.textContent = "Image decrypted successfully!";
        statusElement.className = "status success";
    } catch (error) {
        console.error("Decryption error:", error);
        statusElement.textContent = `Decryption failed: ${error.message}`;
        statusElement.className = "status error";
    }
}

/**
 * Handles download of the decrypted image
 */
async function handleDownloadDecrypted() {
    const statusElement = document.getElementById('decryptionStatus');
    const decryptedImage = document.getElementById('decryptedImage');
    
    try {
        if (!decryptedImage || !decryptedImage.src) {
            statusElement.textContent = "Please decrypt an image first.";
            statusElement.className = "status error";
            return;
        }
        
        statusElement.textContent = "Preparing download...";
        statusElement.className = "status loading";
        
        // Create a download link
        const link = document.createElement('a');
        link.href = decryptedImage.src;
        link.download = 'decrypted_image.png';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        statusElement.textContent = "Image downloaded successfully!";
        statusElement.className = "status success";
    } catch (error) {
        console.error("Download error:", error);
        statusElement.textContent = `Download failed: ${error.message}`;
        statusElement.className = "status error";
    }
}

/**
 * Resets the decrypt page to its initial state
 */
function handleReset() {
    // Reset file input
    const fileInput = document.getElementById('encryptedImageInput');
    if (fileInput) {
        fileInput.value = '';
    }
    
    // Reset decryption key
    const decryptionKey = document.getElementById('decryptionKey');
    if (decryptionKey) {
        decryptionKey.value = '';
    }
    
    // Reset images
    const encryptedImage = document.getElementById('encryptedImage');
    if (encryptedImage) {
        encryptedImage.src = '';
    }
    
    const decryptedImage = document.getElementById('decryptedImage');
    if (decryptedImage) {
        decryptedImage.src = '';
    }
    
    // Reset status messages
    const uploadStatus = document.getElementById('uploadStatus');
    if (uploadStatus) {
        uploadStatus.textContent = '';
        uploadStatus.className = 'status';
    }
    
    const decryptionStatus = document.getElementById('decryptionStatus');
    if (decryptionStatus) {
        decryptionStatus.textContent = '';
        decryptionStatus.className = 'status';
    }
    
    // Reset global variables
    originalImageData = null;
    processedImageData = null;
}

document.addEventListener('DOMContentLoaded', function() {
    // Encrypt page buttons
    const processBtn = document.getElementById('processBtn');
    if (processBtn) {
        processBtn.addEventListener('click', handleProcess);
    }

    // Decrypt page buttons
    const uploadBtn = document.getElementById('uploadBtn');
    if (uploadBtn) {
        uploadBtn.addEventListener('click', handleUpload);
    }

    const decryptBtn = document.getElementById('decryptBtn');
    if (decryptBtn) {
        decryptBtn.addEventListener('click', handleDecrypt);
    }

    const downloadBtn = document.getElementById('downloadBtn');
    if (downloadBtn) {
        // Check if we're on the decrypt page
        if (document.getElementById('decryptedImage')) {
            downloadBtn.addEventListener('click', handleDownloadDecrypted);
        } else {
            downloadBtn.addEventListener('click', handleDownload);
        }
    }

    const processAgainBtn = document.getElementById('processAgainBtn');
    if (processAgainBtn) {
        processAgainBtn.addEventListener('click', handleReset);
    }

    // Show/hide server options based on source selection
    const sourceSelect = document.getElementById('sourceSelect');
    if (sourceSelect) {
        sourceSelect.addEventListener('change', function() {
            const serverOptions = document.getElementById('serverOptions');
            if (serverOptions) {
                serverOptions.style.display = this.value === 'server' ? 'block' : 'none';
            }
        });
    }
});
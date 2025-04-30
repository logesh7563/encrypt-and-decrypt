// Global variables to store image data
let originalImageData = null;
let processedImageData = null;

// Operation parameters configuration
const operationParams = {
    rotate: {
        angle: { type: 'range', min: 0, max: 360, default: 45, label: 'Rotation Angle (degrees)' }
    },
    shearRotate: {
        angle: { type: 'range', min: 0, max: 360, default: 45, label: 'Rotation Angle (degrees)' }
    },
    boxBlur: {
        kernelSize: { type: 'range', min: 3, max: 19, step: 2, default: 3, label: 'Kernel Size' }
    },
    gaussian: {
        sigma: { type: 'range', min: 0.1, max: 10, step: 0.1, default: 1.5, label: 'Sigma' },
        kernelSize: { type: 'range', min: 3, max: 19, step: 2, default: 3, label: 'Kernel Size' }
    },
    sobelEdge: {
        threshold: { type: 'range', min: 0, max: 255, default: 128, label: 'Threshold' }
    }
};

// Function to update operation parameters UI
function updateOperationParams() {
    const operation = document.getElementById('operationSelect').value;
    const paramsContainer = document.getElementById('operationParams');
    
    // Clear existing parameters
    paramsContainer.innerHTML = '';
    
    // If operation has parameters, create UI for them
    if (operationParams[operation]) {
        Object.entries(operationParams[operation]).forEach(([paramName, config]) => {
            const paramGroup = document.createElement('div');
            paramGroup.className = 'param-group';
            
            const label = document.createElement('label');
            label.htmlFor = `param_${paramName}`;
            label.textContent = config.label;
            paramGroup.appendChild(label);
            
            const input = document.createElement('input');
            input.type = config.type;
            input.id = `param_${paramName}`;
            input.min = config.min;
            input.max = config.max;
            input.step = config.step || 1;
            input.value = config.default;
            
            // Add value display for range inputs
            if (config.type === 'range') {
                const valueDisplay = document.createElement('div');
                valueDisplay.className = 'param-value';
                valueDisplay.textContent = config.default + (paramName === 'angle' ? '°' : '');
                
                input.oninput = () => {
                    valueDisplay.textContent = input.value + (paramName === 'angle' ? '°' : '');
                };
                
                paramGroup.appendChild(input);
                paramGroup.appendChild(valueDisplay);
            } else {
                paramGroup.appendChild(input);
            }
            
            paramsContainer.appendChild(paramGroup);
        });
    }
}

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
            
            // Get operation parameters
            const params = {};
            if (operationParams[operation]) {
                Object.keys(operationParams[operation]).forEach(paramName => {
                    const input = document.getElementById(`param_${paramName}`);
                    if (input) {
                        params[paramName] = parseFloat(input.value);
                    }
                });
            }
            
            // Apply the selected operation
            switch (operation) {
                case 'flip':
                    // Upside down - flip vertically
                    ctx.translate(0, canvas.height);
                    ctx.scale(1, -1);
                    break;
                    
                case 'rotate':
                    // Rotate by angle
                    ctx.translate(canvas.width/2, canvas.height/2);
                    ctx.rotate((params.angle || 45) * Math.PI / 180);
                    ctx.translate(-canvas.width/2, -canvas.height/2);
                    break;

                case 'shearRotate':
                    // Three shear matrix rotation - will be implemented server-side
                    ctx.translate(canvas.width/2, canvas.height/2);
                    ctx.rotate((params.angle || 45) * Math.PI / 180);
                    ctx.translate(-canvas.width/2, -canvas.height/2);
                    break;
                    
                case 'grayscale':
                    ctx.filter = 'grayscale(100%)';
                    break;
                    
                case 'boxBlur':
                    ctx.filter = `blur(${params.kernelSize || 3}px)`;
                    break;

                case 'gaussian':
                    ctx.filter = `blur(${params.sigma || 1.5}px)`;
                    break;

                case 'sobelEdge':
                    ctx.filter = `grayscale(100%) contrast(${params.threshold || 128}%)`;
                    break;
            }
            
            // Draw image with the applied filters/transformations
            ctx.drawImage(img, 0, 0, img.width, img.height);
            
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
 * Handles image transmission
 */
async function handleTransmit() {
    const transmissionStatus = document.getElementById('transmissionStatus');
    const serverAddr = document.getElementById('serverAddress').value;
    const imgId = document.getElementById('imageId').value;
    const key = document.getElementById('encryptionKey').value;

    // Input validation
    if (!processedImageData) {
        showStatus(transmissionStatus, 'Please process an image first.', 'error');
        return;
    }

    if (!serverAddr) {
        showStatus(transmissionStatus, 'Please enter a server address.', 'error');
        return;
    }
    
    // Validate server address format
    if (!isValidServerAddress(serverAddr)) {
        showStatus(transmissionStatus, 'Invalid server address format. Use format: hostname:port or IP:port', 'error');
        return;
    }

    if (!imgId) {
        showStatus(transmissionStatus, 'Please enter an image ID.', 'error');
        return;
    }

    if (!key) {
        showStatus(transmissionStatus, 'Please enter an encryption key.', 'error');
        return;
    }

    try {
        // 1) Encrypt processedImageData via your HTTP /api/encrypt endpoint
        showStatus(transmissionStatus, 'Encrypting image for transmission...', 'info');
        console.log(`Encrypting with key: "${key}"`); // Debug log the key being used
        const pngResp = await fetch(processedImageData);
        const pngBlob = await pngResp.blob();
        const encForm = new FormData();
        encForm.append('file', pngBlob, 'image.png');
        encForm.append('key', key);

        const encResp = await fetch('http://localhost:8083/api/encrypt', {
            method: 'POST',
            body: encForm
        });
        if (!encResp.ok) {
            const errText = await encResp.text();
            throw new Error(`Encryption failed: ${errText || encResp.statusText}`);
        }
        const encryptedBuffer = await encResp.arrayBuffer();
        const encryptedBytes = new Uint8Array(encryptedBuffer);
        
        // Convert to base64 without any encoding issues
        let binaryStr = '';
        encryptedBytes.forEach(b => binaryStr += String.fromCharCode(b));
        const base64Data = btoa(binaryStr);
        
        console.log(`Encrypted data: ${encryptedBytes.length} bytes, Base64 length: ${base64Data.length}`);

        // 2) Transmit that encrypted payload to your TCP server with timeout
        showStatus(transmissionStatus, `Transmitting encrypted image to ${serverAddr}...`, 'info');
        
        // Use a Promise with timeout for the API call
        const timeout = 30000; // 30 seconds timeout
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), timeout);
        
        try {
            const txResp = await fetch('http://localhost:8083/api/transmit', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    encryptedData: base64Data,
                    serverAddr: serverAddr,
                    imageID: imgId,
                    key: key  // Include the key for consistency
                }),
                signal: controller.signal
            });
            
            clearTimeout(timeoutId);
            
            if (!txResp.ok) {
                const errorText = await txResp.text();
                throw new Error(`Server error (${txResp.status}): ${errorText}`);
            }
            
            const txJson = await txResp.json();
            
            // Save the image ID and server address in localStorage for easy decryption later
            try {
                localStorage.setItem('lastImageId', imgId);
                localStorage.setItem('lastServerAddr', serverAddr);
                localStorage.setItem('lastEncryptionKey', key);
            } catch (e) {
                console.warn("Couldn't save transmission details to localStorage", e);
            }
            
            showStatus(transmissionStatus, txJson.message, txJson.success ? 'success' : 'error');
        } catch (fetchError) {
            clearTimeout(timeoutId);
            
            if (fetchError.name === 'AbortError') {
                throw new Error(`Connection to ${serverAddr} timed out. Please check the server address and ensure the server is running.`);
            } else {
                throw fetchError;
            }
        }
    } catch (error) {
        console.error('Transmission error:', error);
        showStatus(transmissionStatus, 'Failed to transmit image: ' + error.message, 'error');
    }
}

/**
 * Validates a server address format (host:port)
 * @param {string} address - The server address to validate
 * @returns {boolean} - True if valid format
 */
function isValidServerAddress(address) {
    // Simple regex to validate host:port format
    const regex = /^([a-zA-Z0-9.-]+|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):\d{1,5}$/;
    return regex.test(address);
}

/**
 * Handles decryption of the uploaded image
 */
async function handleDecrypt() {
    const statusElement = document.getElementById('decryptionStatus');
    const decryptionKey = document.getElementById('decryptionKey').value;
    
    try {
        // Check for decryption key
        if (!decryptionKey) {
            statusElement.textContent = "Please enter a decryption key.";
            statusElement.className = "status error";
            return;
        }

        // Check if retrieving from server
        const sourceSelect = document.getElementById('sourceSelect');
        if (sourceSelect && sourceSelect.value === 'server') {
            const serverAddr = document.getElementById('serverAddress').value;
            const imgId = document.getElementById('imageId').value;
            if (!serverAddr || !imgId) {
                statusElement.textContent = "Please enter server address and image ID.";
                statusElement.className = "status error";
                return;
            }
            
            statusElement.textContent = "Retrieving and decrypting image...";
            statusElement.className = "status loading";
            
            try {
                // First try to get the raw encrypted data from the server
                const imageResponse = await fetch('http://localhost:8083/api/request-image', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ serverAddr, imageID: imgId })
                });
                
                if (!imageResponse.ok) {
                    throw new Error(`Failed to retrieve image from server: ${imageResponse.status}`);
                }
                
                const imageResult = await imageResponse.json();
                if (!imageResult.success || !imageResult.encryptedData) {
                    throw new Error(imageResult.message || "Failed to retrieve encrypted image data");
                }
                
                // Now we have the encrypted data as base64, convert it to binary form
                const binary = atob(imageResult.encryptedData);
                const bytes = new Uint8Array(binary.length);
                for (let i = 0; i < binary.length; i++) {
                    bytes[i] = binary.charCodeAt(i);
                }
                
                // Create a blob and form data for decryption
                const blob = new Blob([bytes]);
                const formData = new FormData();
                formData.append('file', blob, 'image.enc');
                formData.append('key', decryptionKey);
                
                // Use the server's decrypt endpoint directly instead of request-decrypt
                const decryptResponse = await fetch('http://localhost:8083/api/decrypt', {
                    method: 'POST',
                    body: formData
                });
                
                if (!decryptResponse.ok) {
                    let errorText = await decryptResponse.text();
                    try {
                        const errorJson = JSON.parse(errorText);
                        if (errorJson.message) errorText = errorJson.message;
                    } catch {}
                    throw new Error(`Decryption failed: ${errorText}`);
                }
                
                // Get the decrypted image as a blob and display it
                const decryptedBlob = await decryptResponse.blob();
                const url = URL.createObjectURL(decryptedBlob);
                document.getElementById('decryptedImage').src = url;
                statusElement.textContent = "Image retrieved and decrypted successfully!";
                statusElement.className = "status success";
                
            } catch (error) {
                console.error('Server decryption error:', error);
                statusElement.textContent = `Server decryption failed: ${error.message}`;
                statusElement.className = "status error";
            }
            return;
        }

        // Local decryption flow (unchanged)
        // Check if an image has been uploaded (only for local decryption)
        if (!originalImageData) {
            statusElement.textContent = "Please upload an image first.";
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
    const decryptionKey = document.getElementById('decryptionKey').value;
    
    try {
        if (!decryptedImage || !decryptedImage.src) {
            statusElement.textContent = "Please decrypt an image first.";
            statusElement.className = "status error";
            return;
        }
        
        if (!decryptionKey) {
            statusElement.textContent = "Please enter the decryption key.";
            statusElement.className = "status error";
            return;
        }
        
        statusElement.textContent = "Preparing download...";
        statusElement.className = "status loading";
        
        // Get the image data from the src
        const response = await fetch(decryptedImage.src);
        const blob = await response.blob();
        
        // Create a FormData object to send to the server
        const formData = new FormData();
        formData.append('file', blob, 'image.enc');
        formData.append('key', decryptionKey);
        
        // Use the specialized endpoint for downloading properly formatted images
        const downloadResponse = await fetch('http://localhost:8083/api/get-decrypted-image', {
            method: 'POST',
            body: formData
        });
        
        if (!downloadResponse.ok) {
            throw new Error(`Failed to format image for download: ${downloadResponse.status}`);
        }
        
        // Get the properly formatted image as a blob
        const formattedImageBlob = await downloadResponse.blob();
        
        // Create a download link with the proper file extension
        const downloadUrl = URL.createObjectURL(formattedImageBlob);
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = 'decrypted_image.png';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        // Clean up the URL object
        URL.revokeObjectURL(downloadUrl);
        
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

    // Add event listener for operation select change
    const operationSelect = document.getElementById('operationSelect');
    if (operationSelect) {
        operationSelect.addEventListener('change', updateOperationParams);
        // Initialize parameters for default selection
        updateOperationParams();
    }
});

// Add this function to app.js
function showStatus(element, message, type) {
    console.log('Status:', type, message);
    if (element) {
        element.textContent = message;
        element.className = 'status ' + type;
        
        // Clear status after delay
        const currentMessage = message;
        setTimeout(() => {
            if (element.textContent === currentMessage) {
                element.textContent = '';
                element.className = 'status';
            }
        }, 5000);
    }
}
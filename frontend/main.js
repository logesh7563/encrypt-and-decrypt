// Global variables
let currentImage = null;
let processedImageData = null;

// DOM elements
const imageInput = document.getElementById('imageInput');
const uploadBtn = document.getElementById('uploadBtn');
const uploadStatus = document.getElementById('uploadStatus');
const operationSelect = document.getElementById('operationSelect');
const angleInput = document.getElementById('angleInput');
const radiusInput = document.getElementById('radiusInput');
const processBtn = document.getElementById('processBtn');
const processingStatus = document.getElementById('processingStatus');
const originalImage = document.getElementById('originalImage');
const processedImage = document.getElementById('processedImage');
const downloadBtn = document.getElementById('downloadBtn');
const transmitBtn = document.getElementById('transmitBtn');
const serverAddress = document.getElementById('serverAddress');
const imageId = document.getElementById('imageId');
const transmissionStatus = document.getElementById('transmissionStatus');
const encryptionKey = document.getElementById('encryptionKey');

// Event listeners
imageInput.addEventListener('change', handleImageSelect);
uploadBtn.addEventListener('click', handleUpload);
operationSelect.addEventListener('change', handleOperationChange);
processBtn.addEventListener('click', handleProcess);
downloadBtn.addEventListener('click', handleDownload);
transmitBtn.addEventListener('click', handleTransmit);

// Handle image selection
function handleImageSelect(event) {
    const file = event.target.files[0];
    if (file) {
        console.log('File selected:', file.name, 'Size:', file.size, 'Type:', file.type);
        const reader = new FileReader();
        reader.onload = function(e) {
            originalImage.src = e.target.result;
            currentImage = file;
            console.log('Image preview loaded');
        };
        reader.readAsDataURL(file);
    }
}

// Handle image upload
async function handleUpload() {
    if (!currentImage) {
        showStatus(uploadStatus, 'Please select an image first', 'error');
        return;
    }

    console.log('Starting upload for file:', currentImage.name);
    const formData = new FormData();
    formData.append('image', currentImage);

    try {
        console.log('Sending upload request...');
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 30000); // 30 second timeout

        const response = await fetch('/api/upload', {
            method: 'POST',
            body: formData,
            signal: controller.signal
        });

        clearTimeout(timeoutId);

        console.log('Response status:', response.status);
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
        }

        const data = await response.json();
        console.log('Upload response data:', data);
        
        if (data.success) {
            console.log('Upload successful');
            showStatus(uploadStatus, data.message, 'success');
            processBtn.disabled = false;
        } else {
            console.log('Upload failed:', data.message);
            showStatus(uploadStatus, data.message || 'Upload failed', 'error');
        }
    } catch (error) {
        console.error('Upload error:', error);
        if (error.name === 'AbortError') {
            showStatus(uploadStatus, 'Upload timed out. Please try again.', 'error');
        } else {
            showStatus(uploadStatus, 'Failed to upload image: ' + error.message, 'error');
        }
    }
}

// Handle operation selection change
function handleOperationChange() {
    const operation = operationSelect.value;
    console.log('Operation selected:', operation);
    
    // Show/hide angle input for rotation operations
    if (operation === 'rotate' || operation === 'rotate_shear') {
        angleInput.style.display = 'block';
    } else {
        angleInput.style.display = 'none';
    }
    
    // Show/hide radius input for blur operations
    if (operation === 'box_blur' || operation === 'gaussian_blur') {
        radiusInput.style.display = 'block';
    } else {
        radiusInput.style.display = 'none';
    }
}

// Handle image processing
async function handleProcess() {
    if (!currentImage) {
        showStatus(processingStatus, 'Please upload an image first', 'error');
        return;
    }

    const operation = operationSelect.value;
    const key = encryptionKey.value;

    if (!key) {
        showStatus(processingStatus, 'Please enter an encryption key', 'error');
        return;
    }

    const params = new URLSearchParams();
    params.append('filename', currentImage.name);

    const requestBody = {
        operation: operation,
        key: key
    };

    // Add operation-specific parameters
    if (operation === 'rotate' || operation === 'rotate_shear') {
        requestBody.angle = parseFloat(document.getElementById('angle').value);
    }
    if (operation === 'box_blur' || operation === 'gaussian_blur') {
        requestBody.radius = parseFloat(document.getElementById('radius').value);
    }

    try {
        const response = await fetch(`/api/process?${params.toString()}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestBody)
        });

        const data = await response.json();
        if (data.success) {
            showStatus(processingStatus, data.message, 'success');
            processedImageData = data.data;
            downloadBtn.disabled = false;
            transmitBtn.disabled = false;
        } else {
            showStatus(processingStatus, data.message, 'error');
        }
    } catch (error) {
        showStatus(processingStatus, 'Failed to process image', 'error');
    }
}

// Handle image download
function handleDownload() {
    if (!processedImageData) {
        showStatus(processingStatus, 'No processed image available', 'error');
        return;
    }

    const key = encryptionKey.value;
    if (!key) {
        showStatus(processingStatus, 'Please enter an encryption key', 'error');
        return;
    }

    window.location.href = `/api/download?data=${encodeURIComponent(processedImageData)}&key=${encodeURIComponent(key)}`;
}

// Handle image transmission
async function handleTransmit() {
    if (!processedImageData) {
        showStatus(transmissionStatus, 'No processed image available', 'error');
        return;
    }

    const serverAddr = serverAddress.value;
    const imgId = imageId.value;
    const key = encryptionKey.value;

    if (!serverAddr || !imgId || !key) {
        showStatus(transmissionStatus, 'Please fill in all required fields', 'error');
        return;
    }

    try {
        // Encrypt image using server encrypt endpoint
        showStatus(transmissionStatus, 'Encrypting image for transmission...', 'info');
        // Convert processedImageData (dataURL) to blob
        const resp = await fetch(processedImageData);
        const imgBlob = await resp.blob();

        const encForm = new FormData();
        encForm.append('file', imgBlob, 'image.png');
        encForm.append('key', key);

        const encResp = await fetch('/api/encrypt', { method: 'POST', body: encForm });
        if (!encResp.ok) {
            const errText = await encResp.text();
            throw new Error(`Encryption for transmission failed: ${errText || encResp.statusText}`);
        }
        // Get raw encrypted binary
        const encryptedBuffer = await encResp.arrayBuffer();
        const encryptedBytes = new Uint8Array(encryptedBuffer);
        let binaryStr = '';
        encryptedBytes.forEach(b => binaryStr += String.fromCharCode(b));
        const base64Data = btoa(binaryStr);

        // Transmit encrypted image
        showStatus(transmissionStatus, 'Transmitting encrypted image...', 'info');
        const transmitResp = await fetch('/api/transmit', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ encryptedData: base64Data, serverAddr, imageID: imgId })
        });
        if (!transmitResp.ok) {
            const errText = await transmitResp.text();
            throw new Error(`Transmission failed: ${errText || transmitResp.statusText}`);
        }
        const data = await transmitResp.json();
        if (data.success) {
            showStatus(transmissionStatus, data.message, 'success');
            if (confirm('Transmission successful! Do you want to decrypt this image from the server?')) {
                localStorage.setItem('serverAddress', serverAddr);
                localStorage.setItem('imageId', imgId);
                localStorage.setItem('decryptKey', key);
                localStorage.setItem('fromTransmission', 'true');
                window.location.href = 'decrypt.html';
            }
        } else {
            showStatus(transmissionStatus, data.message, 'error');
        }
    } catch (error) {
        console.error('Transmission error:', error);
        showStatus(transmissionStatus, 'Failed to transmit image: ' + error.message, 'error');
    }
}

// Helper function to show status messages
function showStatus(element, message, type) {
    console.log('Status:', type, message);
    element.textContent = message;
    element.className = type;
    
    // Clear status after delay, but only if it's the same message
    const currentMessage = element.textContent;
    setTimeout(() => {
        if (element.textContent === currentMessage) {
            element.textContent = '';
            element.className = '';
        }
    }, 5000);
}

// Initialize UI state
console.log('Initializing UI state');
processBtn.disabled = true;
downloadBtn.disabled = true;
transmitBtn.disabled = true;
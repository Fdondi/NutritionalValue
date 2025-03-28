// Simplified app.js - handles image selection, display, weight input, and submission
console.log('Simple app.js loaded');

document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM loaded - setting up image selection');
    
    // Get elements
    const scanButton = document.getElementById('scanButton');
    const imageInput = document.getElementById('imageInput');
    const previewImage = document.getElementById('previewImage');
    const previewSection = document.getElementById('preview');
    const weightInputSection = document.getElementById('weightInputSection');
    const totalWeightInput = document.getElementById('totalWeight');
    const submitButton = document.getElementById('submitButton');
    const resultsSection = document.getElementById('resultsSection');
    const nutritionResults = document.getElementById('nutritionResults');
    const confirmButton = document.getElementById('confirmButton');
    const cancelButton = document.getElementById('cancelButton');
    
    // Store current image data and nutrition info
    let currentImageData = null;
    let currentNutritionInfo = null;
    
    // WebSocket connection
    let ws = null;
    
    // Initialize WebSocket connection
    function connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        console.log(`Connecting to WebSocket at ${wsUrl}`);
        ws = new WebSocket(wsUrl);
        
        ws.onopen = () => {
            console.log('WebSocket connection established');
        };
        
        ws.onmessage = (event) => {
            console.log('WebSocket message received:', event.data);
            try {
                const message = JSON.parse(event.data);
                handleWebSocketMessage(message);
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        };
        
        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
        
        ws.onclose = () => {
            console.log('WebSocket connection closed');
            // Try to reconnect after a delay
            setTimeout(connectWebSocket, 3000);
        };
    }
    
    // Handle WebSocket messages
    function handleWebSocketMessage(message) {
        console.log('Processing message type:', message.type);
        
        if (message.type === 'scan_result') {
            console.log('Received scan results:', message.data);
            
            // Store the nutrition info for later use
            currentNutritionInfo = message.data;
            
            // Reset submit button
            submitButton.disabled = false;
            submitButton.textContent = 'Submit';
            
            // Hide weight input section
            weightInputSection.classList.add('hidden');
            
            // Display the results
            displayNutritionResults(message.data);
            
            // Show results section
            resultsSection.classList.remove('hidden');
        } else if (message.type === 'scan_saved') {
            alert('Nutrition information saved successfully!');
            
            // Reset the form
            resetForm();
        } else if (message.type === 'error') {
            console.error('Server error:', message.message);
            alert(`Error: ${message.message}`);
            
            // Reset submit button if there was an error
            if (submitButton.disabled) {
                submitButton.disabled = false;
                submitButton.textContent = 'Submit';
            }
        }
    }
    
    // Display nutrition results
    function displayNutritionResults(data) {
        console.log('Displaying nutrition results:', data);
        
        // Check if we have the expected properties
        // The server is returning lowercase property names
        const calories = data.calories;
        const protein = data.protein;
        const carbs = data.carbs;
        const fat = data.fat;
        const fiber = data.fiber;
        const sugar = data.sugar;
        const totalWeight = data.total_weight;
        
        const html = `
            <table class="nutrition-table">
                <tr>
                    <th>Nutrient</th>
                    <th>Per 100g</th>
                    <th>Total (${totalWeight}g)</th>
                </tr>
                <tr>
                    <td>Calories</td>
                    <td>${calories.toFixed(1)}</td>
                    <td>${(calories * totalWeight / 100).toFixed(1)}</td>
                </tr>
                <tr>
                    <td>Carbs</td>
                    <td>${carbs.toFixed(1)}g</td>
                    <td>${(carbs * totalWeight / 100).toFixed(1)}g</td>
                </tr>
                ${sugar !== null && sugar !== undefined ? `
                    <tr>
                        <td>Sugar</td>
                        <td>${sugar.toFixed(1)}g</td>
                        <td>${(sugar * totalWeight / 100).toFixed(1)}g</td>
                    </tr>` : ''}
                <tr>
                    <td>Fat</td>
                    <td>${fat.toFixed(1)}g</td>
                    <td>${(fat * totalWeight / 100).toFixed(1)}g</td>
                </tr>                <tr>
                    <td>Protein</td>
                    <td>${protein.toFixed(1)}g</td>
                    <td>${(protein * totalWeight / 100).toFixed(1)}g</td>
                </tr>
                ${fiber !== null && fiber !== undefined ? `
                <tr>
                    <td>Fiber</td>
                    <td>${fiber.toFixed(1)}g</td>
                    <td>${(fiber * totalWeight / 100).toFixed(1)}g</td>
                </tr>` : ''}
            </table>
            <p>Please confirm if the nutrition information is correct.</p>
        `;
        
        nutritionResults.innerHTML = html;
    }
    
    // Reset the form
    function resetForm() {
        // Clear image preview
        previewImage.src = '';
        previewSection.classList.add('hidden');
        
        // Hide results section
        resultsSection.classList.add('hidden');
        
        // Clear weight input
        totalWeightInput.value = '';
        
        // Reset current data
        currentImageData = null;
        currentNutritionInfo = null;
    }
    
    // Send message to server
    function sendMessage(type, data) {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            console.error('WebSocket not connected');
            alert('Not connected to server. Please refresh the page.');
            return;
        }
        
        const message = {
            type: type,
            data: data
        };
        
        console.log('Sending message:', message);
        ws.send(JSON.stringify(message));
    }
    
    // Set up scan button click
    if (scanButton) {
        scanButton.addEventListener('click', () => {
            console.log('Scan button clicked');
            if (imageInput) {
                imageInput.click();
            }
        });
    }
    
    // Set up image input change
    if (imageInput) {
        imageInput.addEventListener('change', (e) => {
            console.log('Image selected');
            if (e.target.files && e.target.files[0]) {
                const file = e.target.files[0];
                
                // Create a FileReader to read the image
                const reader = new FileReader();
                reader.onload = (event) => {
                    console.log('Image loaded');
                    // Store the image data (remove the data:image/jpeg;base64, prefix)
                    currentImageData = event.target.result.split(',')[1];
                    
                    // Display the preview
                    previewImage.src = event.target.result;
                    previewSection.classList.remove('hidden');
                    
                    // Show weight input section
                    weightInputSection.classList.remove('hidden');
                    
                    // Hide results section if visible
                    resultsSection.classList.add('hidden');
                    
                    // Focus on the weight input
                    totalWeightInput.focus();
                };
                
                // Read the file as a data URL
                reader.readAsDataURL(file);
            }
        });
    }
    
    // Set up submit button click
    if (submitButton) {
        submitButton.addEventListener('click', () => {
            console.log('Submit button clicked');
            
            // Validate inputs
            if (!currentImageData) {
                alert('Please select an image first');
                return;
            }
            
            const weight = parseFloat(totalWeightInput.value);
            if (isNaN(weight) || weight <= 0) {
                alert('Please enter a valid weight');
                totalWeightInput.focus();
                return;
            }
            
            // Send data to server
            sendMessage('scan', {
                image: currentImageData,
                totalWeight: weight
            });
            
            // Show loading state or feedback
            submitButton.disabled = true;
            submitButton.textContent = 'Processing...';
        });
    }
    
    // Set up confirm button click
    if (confirmButton) {
        confirmButton.addEventListener('click', () => {
            console.log('Confirm button clicked');
            
            if (!currentNutritionInfo) {
                alert('No nutrition information to confirm');
                return;
            }
            
            // Send confirmation to server
            sendMessage('confirm_scan', {
                id: currentNutritionInfo.id,
                total_weight: currentNutritionInfo.total_weight,
                calories: currentNutritionInfo.calories,
                protein: currentNutritionInfo.protein,
                carbs: currentNutritionInfo.carbs,
                fat: currentNutritionInfo.fat,
                fiber: currentNutritionInfo.fiber,
                sugar: currentNutritionInfo.sugar
            });
            
            // Disable confirm button
            confirmButton.disabled = true;
            confirmButton.textContent = 'Saving...';
        });
    }
    
    // Set up cancel button click
    if (cancelButton) {
        cancelButton.addEventListener('click', () => {
            console.log('Cancel button clicked');
            
            // Hide results section
            resultsSection.classList.add('hidden');
            
            // Show weight input section again
            weightInputSection.classList.remove('hidden');
            
            // Focus on weight input
            totalWeightInput.focus();
        });
    }
    
    // Connect to WebSocket when page loads
    connectWebSocket();
});







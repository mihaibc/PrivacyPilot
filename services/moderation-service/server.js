// ./services/moderation-service/server.js

const express = require('express');
const { requestAiTask, TASK_TYPES } = require('./lib/aiCoordinatorClient'); // Import the client

const app = express();

// --- Configuration ---
const PORT = process.env.PORT || 8082; // Default port for Moderation service

// --- Middleware ---
app.use(express.json()); // Parse JSON request bodies

// --- Routes ---

// Health Check Endpoint
app.get('/health', (req, res) => {
    res.status(200).json({ status: 'OK', service: 'Moderation Service' });
});

// Moderation Endpoint (Refactored and Corrected)
app.post('/moderate', async (req, res, next) => { // Make handler async
    const { text, imageUrl } = req.body;

    if (!text && !imageUrl) {
        console.warn('Moderation request received without text or imageUrl');
        return res.status(400).json({ error: 'Invalid request body: text or imageUrl is required.' });
    }

    console.log(`Received moderation request for: ${text ? 'text' : ''}${text && imageUrl ? ' and ' : ''}${imageUrl ? 'imageUrl' : ''}`);

    try {
        let taskType;
        let payload;

        // Determine task type and payload based on input
        // Prioritize image if both are present, or handle separately
        if (imageUrl) {
            taskType = TASK_TYPES.MODERATE_IMAGE;
            payload = { imageUrl: imageUrl, textContext: text }; // Send text as context if available
            console.log(`Moderation Service: Requesting '${taskType}' from AI Coordinator.`);
        } else { // Only text is present
            taskType = TASK_TYPES.MODERATE_TEXT;
            payload = { text: text };
            console.log(`Moderation Service: Requesting '${taskType}' from AI Coordinator.`);
        }

        // --- Call AI Coordinator ---
        // Pass taskType, payload, and an empty config object {} as the third argument
        const moderationResult = await requestAiTask(taskType, payload, {});
        // --------------------------

        // The client now returns only the 'result' part of the coordinator's response
        // Ensure the structure matches what the coordinator/adapter will return
        // Example expected structure:
        // { is_acceptable: true, flags: [], details: "...", confidence_score: 0.95 }

        console.log('Moderation Service: Successfully processed moderation via AI Coordinator.');
        res.status(200).json(moderationResult); // Return the result directly

    } catch (error) {
        // Log the specific error from the coordinator call
        console.error(`Moderation Service: Error during AI Coordinator call: ${error.message}`);
        // Pass error to the global error handler for consistent response format
        next(error); // Use next(error) for async errors in Express
    }
});


// --- Global Error Handler (Basic) ---
// Catches errors passed via next(error)
app.use((err, req, res, next) => {
    // Log the error internally
    console.error("Unhandled error:", err.message);
    // Send a generic error message to the client to avoid leaking details
    res.status(500).json({ error: 'An internal error occurred while processing the moderation request.' });
});

// --- Start Server ---
const server = app.listen(PORT, () => {
    console.log(`Moderation Service listening on port ${PORT}`);
});

// --- Graceful Shutdown Logic ---
const gracefulShutdown = (signal) => {
    console.log(`${signal} signal received: closing HTTP server`);
    server.close(() => {
        console.log('HTTP server closed');
        // Add any other cleanup logic here (e.g., close database connections)
        process.exit(0); // Exit gracefully
    });
};

// Listen for termination signals
process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
process.on('SIGINT', () => gracefulShutdown('SIGINT')); // Handle Ctrl+C

module.exports = server; // Export for testing purposes
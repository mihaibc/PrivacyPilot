const express = require('express');
const { requestAiTask, TASK_TYPES } = require('./lib/aiCoordinatorClient'); // Import the client

const app = express();

// --- Configuration ---
const PORT = process.env.PORT || 8082;

// --- Middleware ---
app.use(express.json());

// --- Routes ---

// Health Check Endpoint
app.get('/health', (req, res) => {
    res.status(200).json({ status: 'OK', service: 'Moderation Service' });
});

// Moderation Endpoint (Refactored)
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
        const moderationResult = await requestAiTask(taskType, payload);
        // The client now returns only the 'result' part of the coordinator's response
        // Ensure the structure matches what the coordinator/adapter will return
        // Example expected structure based on previous Go definition:
        // { is_acceptable: true, flags: [], details: "...", confidence_score: 0.95 }

        console.log('Moderation Service: Successfully processed moderation via AI Coordinator.');
        res.status(200).json(moderationResult); // Return the result directly

    } catch (error) {
        console.error(`Moderation Service: Error during AI Coordinator call: ${error.message}`);
        // Pass error to the global error handler
        next(error); // Use next(error) for async errors
    }
});


// --- Global Error Handler (Basic) ---
app.use((err, req, res, next) => {
    console.error("Unhandled error:", err.message); // Log just the message for cleaner output
    // Provide a more generic error message to the client
    res.status(500).json({ error: 'An internal error occurred while processing the moderation request.' });
});

// --- Start Server & Graceful Shutdown ---
// (Keep the server start and graceful shutdown logic from the previous version)
const server = app.listen(PORT, () => {
    console.log(`Moderation Service listening on port ${PORT}`);
});

process.on('SIGTERM', () => { /* ... */ });
process.on('SIGINT', () => { /* ... */ });

module.exports = server; // Export for testing
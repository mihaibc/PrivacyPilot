const AI_COORDINATOR_URL = process.env.AI_COORDINATOR_URL;
const REQUEST_TIMEOUT_MS = 20000; // 20 seconds timeout for AI tasks

if (!AI_COORDINATOR_URL) {
    console.error('FATAL ERROR: AI_COORDINATOR_URL environment variable is not set.');
    process.exit(1);
}

// Define task types (should match definitions in Go client/coordinator)
const TASK_TYPES = {
    MODERATE_TEXT: 'moderate_text',
    MODERATE_IMAGE: 'moderate_image',
    // Add other types as needed
};

/**
 * Sends a task request to the AI Coordinator service.
 * @param {string} taskType - The type of task (e.g., TASK_TYPES.MODERATE_TEXT).
 * @param {object} payload - The specific data for the task.
 * @param {object} [config] - Optional configuration for the task.
 * @returns {Promise<object>} - The result part of the AI Coordinator's response.
 * @throws {Error} - If the request fails or the coordinator returns an error.
 */
async function requestAiTask(taskType, payload, config = {}) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);

    const requestBody = {
        task_type: taskType,
        payload: payload,
        ...(Object.keys(config).length > 0 && { config }), // Add config only if not empty
    };

    const url = `${AI_COORDINATOR_URL}/process`; // Assuming /process endpoint
    console.log(`Sending task '${taskType}' to AI Coordinator at ${url}`);

    try {
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
            },
            body: JSON.stringify(requestBody),
            signal: controller.signal, // Add abort signal for timeout
        });

        clearTimeout(timeoutId); // Clear timeout if fetch completes

        const responseBody = await response.json(); // Attempt to parse JSON regardless of status

        if (!response.ok || !responseBody.success) {
            const errorMessage = responseBody?.error || `AI Coordinator returned status ${response.status}`;
            console.error(`AI Coordinator task failed: ${errorMessage}`, responseBody);
            throw new Error(`AI Coordinator task '${taskType}' failed: ${errorMessage}`);
        }

        console.log(`Successfully received result for task '${taskType}' from AI Coordinator.`);
        return responseBody.result; // Return only the result part on success

    } catch (error) {
        clearTimeout(timeoutId); // Ensure timeout is cleared on error
        if (error.name === 'AbortError') {
            console.error(`AI Coordinator request timed out after ${REQUEST_TIMEOUT_MS}ms`);
            throw new Error(`AI Coordinator request timed out`);
        }
        console.error(`Error communicating with AI Coordinator: ${error.message}`);
        // Rethrow a generic error or the specific error
        throw new Error(`Failed to communicate with AI Coordinator: ${error.message}`);
    }
}

module.exports = {
    requestAiTask,
    TASK_TYPES,
};
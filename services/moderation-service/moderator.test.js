const request = require('supertest');
const app = require('./server'); // Import the Express app

// --- Mock the AI Coordinator Client ---
// Mock the entire module
jest.mock('./lib/aiCoordinatorClient', () => ({
    // Mock the exported function 'requestAiTask'
    requestAiTask: jest.fn(),
    // Also mock exported constants if they are used directly in tests (optional here)
    TASK_TYPES: {
        MODERATE_TEXT: 'moderate_text',
        MODERATE_IMAGE: 'moderate_image',
    },
}));
// Import the mocked client *after* jest.mock
const { requestAiTask, TASK_TYPES } = require('./lib/aiCoordinatorClient');
// ------------------------------------


// Ensure the server doesn't stay open after tests
afterAll((done) => {
    app.close(done);
});

// Reset mocks before each test
beforeEach(() => {
    requestAiTask.mockClear(); // Clear call history and reset mock implementation
});


describe('Moderation Service API', () => {
    // Test Health Check Endpoint (remains the same)
    describe('GET /health', () => {
        it('should respond with 200 OK and service status', async () => {
            const response = await request(app).get('/health');
            expect(response.statusCode).toBe(200);
            // ... rest of assertions
            expect(response.body).toEqual({
                status: 'OK',
                service: 'Moderation Service',
            });
        });
    });

    // Test Moderation Endpoint (Refactored Tests)
    describe('POST /moderate', () => {
        it('should return 400 if text and imageUrl are missing', async () => {
            const response = await request(app)
                .post('/moderate')
                .send({});
            expect(response.statusCode).toBe(400);
            expect(response.body.error).toContain('text or imageUrl is required');
            expect(requestAiTask).not.toHaveBeenCalled(); // Ensure client wasn't called
        });

        it('should call AI Coordinator with MODERATE_TEXT and return its result for text input', async () => {
            const mockCoordResult = { // Example successful result from coordinator
                is_acceptable: true,
                flags: [],
                details: "AI Coordinator processed text successfully.",
                confidence_score: 0.99
            };
            // Configure the mock function to return the mock result
            requestAiTask.mockResolvedValueOnce(mockCoordResult);

            const inputText = "This is some input text.";
            const response = await request(app)
                .post('/moderate')
                .send({ text: inputText });

            // Check response from our service
            expect(response.statusCode).toBe(200);
            expect(response.body).toEqual(mockCoordResult); // Should match coordinator result

            // Check if the mock client was called correctly
            expect(requestAiTask).toHaveBeenCalledTimes(1);
            expect(requestAiTask).toHaveBeenCalledWith(
                TASK_TYPES.MODERATE_TEXT, // Check task type
                { text: inputText },      // Check payload
                expect.any(Object)       // Config object might be empty or not, check type
            );
        });

        it('should call AI Coordinator with MODERATE_IMAGE and return its result for image input', async () => {
            const mockCoordResult = {
                is_acceptable: false,
                flags: ['unsafe_image'],
                details: "AI Coordinator flagged image.",
                confidence_score: 0.85
            };
            requestAiTask.mockResolvedValueOnce(mockCoordResult);

            const inputImageUrl = "http://example.com/image.jpg";
            const response = await request(app)
                .post('/moderate')
                .send({ imageUrl: inputImageUrl });

            expect(response.statusCode).toBe(200);
            expect(response.body).toEqual(mockCoordResult);

            expect(requestAiTask).toHaveBeenCalledTimes(1);
            expect(requestAiTask).toHaveBeenCalledWith(
                TASK_TYPES.MODERATE_IMAGE,
                { imageUrl: inputImageUrl, textContext: undefined }, // Check payload
                expect.any(Object)
            );
        });

        it('should prioritize MODERATE_IMAGE and include text context if both inputs are present', async () => {
            const mockCoordResult = { is_acceptable: true, flags: [], details: "Image OK with context", confidence_score: 0.9 };
            requestAiTask.mockResolvedValueOnce(mockCoordResult);

            const inputText = "Context for image";
            const inputImageUrl = "http://example.com/image.jpg";
            const response = await request(app)
                .post('/moderate')
                .send({ text: inputText, imageUrl: inputImageUrl });

             expect(response.statusCode).toBe(200);
             expect(response.body).toEqual(mockCoordResult);

             expect(requestAiTask).toHaveBeenCalledTimes(1);
             expect(requestAiTask).toHaveBeenCalledWith(
                 TASK_TYPES.MODERATE_IMAGE,
                 { imageUrl: inputImageUrl, textContext: inputText }, // textContext should be included
                 expect.any(Object)
             );
        });


        it('should return 500 if AI Coordinator call fails', async () => {
            const errorMessage = "AI Coordinator task failed: AI Model Error";
            // Configure the mock function to throw an error
            requestAiTask.mockRejectedValueOnce(new Error(errorMessage));

            const response = await request(app)
                .post('/moderate')
                .send({ text: "This will fail" });

            expect(response.statusCode).toBe(500);
            expect(response.body).toHaveProperty('error');
            expect(response.body.error).toContain('internal error occurred'); // Check for generic error message

            expect(requestAiTask).toHaveBeenCalledTimes(1);
        });
    });
});
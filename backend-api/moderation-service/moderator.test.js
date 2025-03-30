const request = require('supertest');
const express = require('express');
const bodyParser = require('body-parser');
const { moderateContent } = require('./moderator');

const app = express();
app.use(bodyParser.json());
app.post('/moderate', (req, res) => {
    const { text } = req.body;
    const moderatedText = moderateContent(text);
    res.json({ moderatedText });
});

describe('Moderation Service', () => {
    test('should censor badword', async () => {
        const response = await request(app)
            .post('/moderate')
            .send({ text: "This is a badword test" });
        expect(response.body.moderatedText).toBe("This is a **** test");
    });

    test('should return original text if no bad words', async () => {
        const response = await request(app)
            .post('/moderate')
            .send({ text: "This is a clean text" });
        expect(response.body.moderatedText).toBe("This is a clean text");
    });
});

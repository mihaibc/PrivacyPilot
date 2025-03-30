const express = require('express');
const bodyParser = require('body-parser');
const { moderateContent } = require('./moderator');

const app = express();
app.use(bodyParser.json());

app.post('/moderate', (req, res) => {
    const { text } = req.body;
    if (!text) {
        return res.status(400).json({ error: 'Text is required' });
    }
    const result = moderateContent(text);
    res.json({ moderatedText: result });
});

const PORT = process.env.PORT || 8082;
app.listen(PORT, () => {
    console.log(`Moderation Service running on port ${PORT}`);
});

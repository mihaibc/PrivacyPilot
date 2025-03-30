// moderateContent applies a simple rule-based filter.
// In production, this could be replaced with AI or more complex logic.
function moderateContent(text) {
    const badWords = ["badword"];
    let moderatedText = text;
    badWords.forEach(word => {
        const regex = new RegExp(word, 'gi');
        moderatedText = moderatedText.replace(regex, '****');
    });
    return moderatedText;
}

module.exports = { moderateContent };

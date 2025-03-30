# Guidelines for Creating Issues and Pull Requests

## Creating Issues
To ensure effective communication, please follow these guidelines when opening a new issue:

### Issue Template
- **Title:** Clearly summarize the issue.
- **Description:** Provide a detailed description, including:
  - **Expected behavior:** Clearly explain what should happen.
  - **Actual behavior:** Describe the issue encountered.
  - **Steps to reproduce:** Provide detailed steps.
  - **Environment details:** Mention versions, OS, or any specific configurations.

### Example Issue:
```
Title: Error on anonymization endpoint when processing large text inputs

Description:
- **Expected:** Endpoint successfully anonymizes large texts (>10,000 characters).
- **Actual:** HTTP 500 error occurs when text input exceeds 10,000 characters.
- **Steps to reproduce:**
  1. Start PrivacyPilot locally.
  2. POST a large JSON payload to `/anonymize` endpoint.
- **Environment:** Docker Compose local deployment, Node.js v20.
```

## Creating Pull Requests
Follow these guidelines to streamline the review process:

### Pull Request Template
- **Title:** Concisely describe your PR.
- **Description:** Clearly mention:
  - The issue it resolves (link to issue, e.g., `Fixes #IssueNumber`).
  - Summary of changes made.
  - Any relevant testing done.

### Example PR:
```
Title: Fix anonymization endpoint for large payloads (#45)

Description:
Fixes #45

- Updated payload parsing logic.
- Increased JSON body parser limits in Express.
- Added unit tests to verify functionality.

Tested locally and via integration tests.
```

## Best Practices
- Ensure your PR addresses only one issue or feature at a time.
- Run existing tests before submission.
- Keep discussions professional and respectful.

---

🚀 **Thank you for improving PrivacyPilot!**
# Contribution Guidelines

Thank you for your interest in contributing to PrivacyPilot! Your help is greatly appreciated.

## How to Contribute
You can contribute in the following ways:
- **Reporting Bugs:** Clearly describe the issue, expected vs. actual behavior, and reproduction steps.
- **Suggesting Features:** Propose new features or enhancements via issues.
- **Improving Documentation:** Clarify, update, or enhance documentation.
- **Code Contributions:** Implement new features, fix bugs, or optimize existing functionality.

## Contribution Workflow

### Step 1: Fork the Repository
- Fork the repository to your own GitHub account.

### Step 2: Create an Issue
- Before working on your changes, **create an issue** in the main repository.
- Clearly describe the purpose of your issue (bug fix, enhancement, etc.).

### Step 3: Create a Branch
- Clone your forked repository and create a descriptive branch:
```bash
git clone https://github.com/<your-username>/PrivacyPilot.git
cd PrivacyPilot
git checkout -b feature/<short-description>
```

### Step 4: Make Your Changes
- Follow existing coding styles and conventions.
- Keep your changes focused on a single issue.

### Step 5: Commit Your Changes
- Write clear and concise commit messages.
```bash
git commit -m "feat: implement feature X (#IssueNumber)"
```

### Step 6: Push and Open a Pull Request
- Push your changes to your forked repository:
```bash
git push origin feature/<short-description>
```
- Open a pull request against the `main` branch of the **source repository**.
- **Link the issue** you've created previously in your pull request description:
```
Fixes #IssueNumber
```

## Review Process
- Your pull request will be reviewed by maintainers.
- Please address any feedback promptly.
- Once approved, maintainers will merge your changes.

## Important Notes
- Ensure that all PRs reference a clearly defined issue.
- PRs without a linked issue may not be reviewed promptly.

---

🙏 **Thank you for contributing to PrivacyPilot!**

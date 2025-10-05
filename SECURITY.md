# Security Policy

## Supported versions

Selene is currently in active development. We aim to keep the `main` branch secure and patched. When a tagged release is cut, the most recent minor release will continue receiving security fixes.

## Reporting a vulnerability

If you discover a security vulnerability, please open an issue with the following information:

- A detailed description of the issue.
- Steps to reproduce and, if possible, a minimal proof-of-concept.
- The potential impact and any suggested mitigations.

We will acknowledge receipt within **3 business days** and provide regular updates while we investigate. Once a fix is available, we will coordinate a disclosure timeline with you before publishing an advisory.

Please do not create public GitHub issues for security-sensitive reports.

## Security best practices for contributors

- Run `make vulncheck` locally or in CI before submitting pull requests.
- Keep dependencies locked in `selene.lock` and prefer tagged releases for vendored code.
- Avoid committing secrets, credentials, or private keys.

Thank you for helping keep Selene secure!

# Security Policy

## Supported Versions

This repository currently supports the latest `main` branch.

## Reporting a Vulnerability

Do not open a public issue for a suspected vulnerability.

Report security issues privately through GitHub security advisories or direct maintainer contact.

When reporting, include:

- affected area
- reproduction steps
- expected impact
- suggested mitigation if known

## Repository Safety Rules

- Never commit secrets, tokens, API keys, certificates, or private keys.
- Keep local caches, runtime logs, and worktree scratch directories out of version control.
- Treat `external/` as reference material only; it must not be published with the main project.
- Run CI and secret scanning before publishing changes.

# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and the project uses Semantic Versioning.

## [0.1.0] - 2026-04-17

### Added

- FastAPI service with `/convert`, `/health`, and `/ready` endpoints.
- Streamlit UI for uploading Markdown, previewing rendered HTML, and downloading output.
- Docker image packaging for running the API and UI together.
- GitHub Actions workflow for building and publishing Docker images to GitHub Container Registry on every GitHub release.

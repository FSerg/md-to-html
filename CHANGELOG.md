# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and the project uses Semantic Versioning.

## [0.1.2] - 2026-04-18

### Added

- Streamlit UI now supports converting Markdown pasted from the clipboard in addition to uploaded `.md` files.

## [0.1.1] - 2026-04-17

### Fixed

- Removed GitHub Actions cache export from the Docker release workflow after GitHub Actions Cache returned `502` during image publication.

## [0.1.0] - 2026-04-17

### Added

- FastAPI service with `/convert`, `/health`, and `/ready` endpoints.
- Streamlit UI for uploading Markdown, previewing rendered HTML, and downloading output.
- Docker image packaging for running the API and UI together.
- GitHub Actions workflow for building and publishing Docker images to GitHub Container Registry on every GitHub release.

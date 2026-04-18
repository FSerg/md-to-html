# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and the project uses Semantic Versioning.

## [0.2.2] - 2026-04-18

### Changed

- Web UI redesigned into a single-column classic layout with tabbed file/text input, updated result card, and API snippet.
- Added a local release build script and Make targets for current-platform and cross-platform release artifacts.
- README updated with the current build, release, and run instructions.

## [0.2.1] - 2026-04-18

### Fixed

- GitHub release workflow now lowercases the GHCR image name before publishing, which fixes releases for repositories with uppercase owner names.

## [0.2.0] - 2026-04-18

### Changed

- **BREAKING**: project fully rewritten in Go (goldmark + templUI); Python implementation moved to `archive/`.
- **BREAKING**: heading anchors now use ASCII transliteration (`## Установка` → `id="ustanovka"`).
- **BREAKING**: heading HTML markup simplified; `<div class="markdown-heading">` is no longer emitted.
- Removed the GitHub Markdown API dependency; conversion now works fully offline.
- Replaced the two-process runtime (uvicorn + Streamlit) with a single binary.
- Preview and download links are now one-shot, UUID-backed, and expire after one hour.

### Added

- Syntax highlighting via chroma with inline styles for self-contained HTML output.
- Footnote support in addition to baseline GFM features.
- Cross-platform release binaries for `linux/amd64`, `linux/arm64`, and `darwin/arm64`.

### Removed

- `READY_CHECK_GITHUB` environment variable.
- Streamlit UI on dedicated port `:8501`.

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

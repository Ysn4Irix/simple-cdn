---
type: project
created: 2026-07-18
updated: 2026-07-18
---

# Technical Decisions

- Component metadata uses SemVer while the toolkit release keeps CalVer.
- `manifest.json` and `manifest.lock.json` must remain synchronized with component frontmatter.
- [cdn-architecture] Simple Static CDN built as a single-node Go service with authenticated REST API, lightweight Web UI, HTTP Byte-Range, ETags, Gzip/Brotli, signed URLs, hotlink protection, and rate limiting.

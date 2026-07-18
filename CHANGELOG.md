# Changelog

Format: [Keep a Changelog](https://keepachangelog.com/); versioning: [SemVer](https://semver.org/).

## [0.1.0] - 2026-07-17

Initial release — a four-language real-time transcription & translation grid.

### Added
- **Rust DSP**: a from-scratch recursive radix-2 Cooley-Tukey FFT (and inverse),
  a spectral noise gate, dominant-frequency detection, and voiced-segment
  tokenisation over raw PCM. axum `/process`. FFT peak-detection + round-trip
  tests.
- **Go SFU**: a hand-rolled RFC 6455 WebSocket (handshake + frame codec, no
  library) and a selective-forwarding unit with a worker pool that runs
  DSP→STT→MT and broadcasts transcripts, plus a synthetic speaker. Tests.
- **Python grid**: a deterministic mock STT (DSP tokens → words) and a
  phrase-table MT (en→es/fr/de) with optional Whisper / neural-MT adapters,
  FastAPI `/transcribe` `/translate` `/pipeline`, CLI, and tests.
- **Next.js client**: Web Audio mic capture, PCM framing over WebSocket, and a
  live multi-speaker caption feed with a translation-language selector.
- Shared JSON protocol, docker-compose, per-service Dockerfiles, multi-language
  CI, Makefile, offline verifier, MIT license.

# babelgrid 🎙️

**A real-time audio transcription & translation grid.** A Go selective-
forwarding unit ingests audio from many speakers over WebSocket, a Rust DSP
layer cleans each frame with a **from-scratch FFT** and extracts features, a
Python grid transcribes and translates, and everyone sees live cross-translated
captions.

Four languages, each on the layer it's built for, over one JSON protocol:

```
 mic ─▶ TS client ─ws─▶ Go SFU ─▶ Rust DSP ─▶ Python grid ─▶ transcript+translations
                         │ (worker pool)   (FFT, gate)    (STT + MT)        │
                         └────────── broadcast to every listener ◀──────────┘
```

| Layer | Language | Owns |
| --- | --- | --- |
| **Client** | TypeScript | Mic capture (Web Audio), PCM framing, live caption feed |
| **SFU** | Go | From-scratch WebSocket, worker-pool fan-out to DSP + grid |
| **DSP** | Rust | From-scratch FFT, spectral noise gate, voiced-segment tokens |
| **Grid** | Python | Mock STT (tokens→words) + phrase-table MT (en→es/fr/de) |

Runs **offline** — no Opus codec (PCM over JSON), no external WebSocket library
(RFC 6455 by hand), no FFT crate (Cooley-Tukey from scratch), no ASR weights (a
deterministic mock STT). Real backends drop in via env vars.

## Quickstart — the pieces, offline

```bash
cd dsp-rust    && cargo test          # FFT + noise-gate + feature tests
cd sfu-go      && go test ./...        # WebSocket framing + SFU tests
cd grid-python && pip install -e ".[dev]" && python -m babelgrid_grid.cli demo
```

```
transcript: 'the quick meeting review pull'
  es: el rápido reunión revisar extraer
  fr: le rapide réunion réviser tirer
  de: die schnelle besprechung überprüfen ziehen
```

Offline end-to-end check:

```bash
python scripts/verify.py     # RESULT: N passed, 0 failed
```

## Quickstart — the whole grid

```bash
docker compose up --build
# Client: http://localhost:3000   (click "start mic" and speak)
# SFU:    http://localhost:8080/healthz
# DSP:    http://localhost:8092/healthz
# Grid:   http://localhost:8000/healthz
```

The SFU runs with `BABELGRID_SYNTH=1`, so a synthetic speaker exercises the
pipeline even without a mic.

## The interesting engineering

- **From-scratch FFT (Rust)** — recursive radix-2 Cooley-Tukey + inverse,
  spectral noise gate, dominant-frequency detection, and voiced-segment
  tokenisation. `dsp-rust/src/fft.rs`, `dsp.rs`
- **From-scratch WebSocket (Go)** — RFC 6455 handshake + frame codec
  (FIN/opcode, 7/16/64-bit lengths, client mask XOR), no third-party library.
  `sfu-go/internal/ws/`
- **Worker-pool SFU (Go)** — audio frames fan out to a pool that runs
  DSP→STT→MT and broadcasts transcripts; a saturated pool drops frames rather
  than stalling ingestion. `sfu-go/internal/sfu/`
- **Transcription grid (Python)** — deterministic token→word STT and a
  phrase-table translator, with real Whisper/MT adapters behind env flags.
  `grid-python/babelgrid_grid/`

## Testing

```bash
make test                     # rust + go + python
cd dsp-rust    && cargo test
cd sfu-go      && go test ./...
cd grid-python && pytest -q
cd client-ts   && npm run build
```

## Layout

```
proto/            shared JSON audio + transcript protocol
client-ts/        Next.js mic client + live caption feed
sfu-go/           Go selective-forwarding unit (hand-rolled WebSocket + pool)
dsp-rust/         Rust DSP: from-scratch FFT, noise gate, features (axum)
grid-python/      mock STT + phrase-table MT + FastAPI
scripts/verify.py offline end-to-end check
docs/ARCHITECTURE.md
```

## License

MIT © 2026 Akhil Vase

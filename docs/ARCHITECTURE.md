# babelgrid architecture

An audio routing + ML pipeline mesh: many speakers in, cross-translated captions
out, under a second. Each language owns its layer; one JSON contract
(`proto/protocol.md`) connects them.

```
  browser mic (Web Audio)
        │ PCM frames over WebSocket
        ▼
┌──────────────────────────┐  worker pool
│ SFU · Go                 │───────────────┐
│ RFC6455 + fan-out        │               │ POST /process
└───────┬──────────────────┘               ▼
        │ broadcast              ┌──────────────────────────┐
        │ transcripts            │ DSP · Rust               │
        │                        │ FFT, noise gate, tokens  │
        │                        └───────────┬──────────────┘
        │                        POST /pipeline (tokens)
        │                                    ▼
        │                        ┌──────────────────────────┐
        │                        │ Grid · Python            │
        │                        │ mock STT + phrase-table MT│
        │                        └───────────┬──────────────┘
        ▼                                    │
┌──────────────────────────┐  {text, translations}
│ every listener (TS)      │◀────────────────┘
└──────────────────────────┘
```

## Why each language

| Layer | Language | Reason |
| --- | --- | --- |
| Client | **TypeScript** | Web Audio mic capture and a live caption UI. |
| SFU | **Go** | Many concurrent WebSocket callers + a cheap worker pool. |
| DSP | **Rust** | Tight numeric loops (FFT, gating) at native speed, no GC. |
| Grid | **Python** | Where ASR/MT live; here an auditable mock + real adapters. |

## Flow

1. The browser captures mic audio, converts Float32 samples to 16-bit PCM, and
   streams ~85 ms frames to the Go SFU over a hand-rolled WebSocket.
2. The SFU pushes each frame into a **worker pool** (the "split out to worker
   nodes" step). A saturated pool drops frames rather than stalling ingestion.
3. A worker sends the frame to the Rust DSP `/process`, which runs a from-scratch
   FFT, applies a spectral **noise gate**, measures the dominant frequency, and
   emits one stable **token** per voiced segment. Silence is dropped here.
4. The worker sends the tokens to the Python grid `/pipeline`, which maps tokens
   to words (mock STT) and translates the phrase into each target language
   (phrase-table MT).
5. The SFU broadcasts `{speaker, text, translations, dominant_hz}` to every
   listener in the room; each client shows the translation for its selected
   language.

## The DSP

`dsp-rust` implements a recursive radix-2 Cooley-Tukey FFT and its inverse from
scratch. `process()` normalises PCM to `[-1,1]`, computes RMS, runs a spectral
noise gate (zero bins below a fraction of the peak magnitude, then inverse FFT),
finds the dominant bin for `dominant_hz`, and windows the signal into ~10 ms
frames to detect voiced segments — each segment hashed to a token the STT layer
maps to a word. FFT correctness is unit-tested (peak-bin detection and
FFT→IFFT round-trip).

## The WebSocket

`sfu-go/internal/ws` implements RFC 6455 directly: `Sec-WebSocket-Accept =
base64(sha1(key + GUID))`, and a frame codec handling FIN/opcode, 7/16/64-bit
payload lengths, and client mask XOR. No third-party dependency.

## Offline-first

- **DSP**: no FFT crate — Cooley-Tukey by hand.
- **SFU**: no WebSocket library — RFC 6455 by hand; a synthetic speaker
  (`BABELGRID_SYNTH=1`) drives the pipeline with no mic.
- **Grid**: no ASR weights — a deterministic token→word STT and a phrase-table
  MT. `BABELGRID_STT=whisper` / `BABELGRID_MT=real` swap in real models.
- **Transport**: PCM integer arrays over JSON — no Opus codec needed.

STT content is intentionally a mock offline; the point is that the DSP, routing,
worker pool, translation, and fan-out are all real and exercised end to end.

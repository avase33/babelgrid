# babelgrid wire protocol

Audio flows client → Go SFU → Rust DSP → Python grid, and transcripts flow back
to every listener. One JSON contract throughout; audio is carried as 16-bit PCM
integer arrays (keeps the whole path dependency-free — no Opus codec needed).

## 1. Audio frame (browser → Go SFU, WebSocket `/ws`)

```json
{
  "type": "audio",
  "room": "call1",
  "speaker": "alice",
  "lang": "en",
  "sr": 16000,
  "pcm": [12, -40, 130, ...]
}
```

`pcm` is signed 16-bit samples; `sr` the sample rate. Frames are ~20-40 ms.

## 2. DSP request/response (Go SFU → Rust `/process`)

Request: `{ "pcm": [ ... ], "sr": 16000 }`

Response — cleaned-signal features:

```json
{
  "rms": 0.121,
  "rms_gated": 0.118,
  "voiced": true,
  "dominant_hz": 218.7,
  "duration_ms": 32.0,
  "tokens": [11, 42, 7]
}
```

- `rms` / `rms_gated`: energy before/after the spectral noise gate.
- `voiced`: whether the frame carries speech (energy over the gate).
- `dominant_hz`: peak frequency from the from-scratch FFT.
- `tokens`: one stable id per detected voiced segment — the STT layer maps these
  to words.

## 3. Transcription (Go SFU → Python `/pipeline`)

Request: `{ "tokens": [11,42,7], "targets": ["es","fr","de"] }`

Response:

```json
{
  "text": "the quick meeting",
  "translations": { "es": "la rápida reunión", "fr": "la rapide réunion", "de": "die schnelle besprechung" }
}
```

Python also exposes `/transcribe` (`{tokens}` → `{text}`) and `/translate`
(`{text, targets}` → `{translations}`) separately.

## 4. Transcript broadcast (Go SFU → all room clients, WebSocket)

```json
{
  "type": "transcript",
  "room": "call1",
  "speaker": "alice",
  "text": "the quick meeting",
  "translations": { "es": "...", "fr": "...", "de": "..." },
  "dominant_hz": 218.7,
  "ts": 1752710400.5
}
```

## Ports

| service | port | protocol |
| --- | --- | --- |
| Go SFU | 8080 | WebSocket `/ws`, HTTP `/healthz` |
| Rust DSP | 8092 | HTTP `/process` `/healthz` |
| Python grid | 8000 | HTTP `/pipeline` `/transcribe` `/translate` `/healthz` |
| TS client | 3000 | HTTP |

> STT is a deterministic **mock** offline (no acoustic model without weights);
> `BABELGRID_STT=whisper` swaps in a real backend. Translation is a phrase-table
> MT with an optional real adapter.

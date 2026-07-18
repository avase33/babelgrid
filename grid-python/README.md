# babelgrid-grid

The transcription grid: a deterministic **mock STT** (DSP tokens → words) and a
**phrase-table MT** (en → es/fr/de), served over FastAPI. No model weights, no
network — the whole path is auditable. `BABELGRID_STT=whisper` /
`BABELGRID_MT=real` swap in real backends via `real.py`.

```bash
pip install -e ".[dev]"
python -m babelgrid_grid.cli demo     # tokens -> transcript -> translations
babelgrid-grid serve                  # FastAPI on :8000
pytest -q
```

Endpoints: `/transcribe {tokens}`, `/translate {text, targets}`,
`/pipeline {tokens, targets}`.

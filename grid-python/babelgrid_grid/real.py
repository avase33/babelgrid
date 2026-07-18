"""Optional real STT / MT adapters (used only when explicitly enabled).

Imports are lazy so the offline path never depends on these packages.
"""

from __future__ import annotations

import os
from typing import Dict, List


def transcribe_whisper(pcm: List[int], sr: int) -> str:
    """Transcribe real audio with faster-whisper (BABELGRID_STT=whisper)."""
    import numpy as np  # lazy
    from faster_whisper import WhisperModel  # lazy

    model = WhisperModel(os.getenv("BABELGRID_WHISPER_MODEL", "tiny"))
    audio = np.asarray(pcm, dtype=np.float32) / 32768.0
    segments, _ = model.transcribe(audio, language=None)
    return " ".join(seg.text.strip() for seg in segments).strip()


def translate_real(text: str, targets: List[str]) -> Dict[str, str]:
    """Translate with a neural MT backend (BABELGRID_MT=real)."""
    from transformers import pipeline as hf_pipeline  # lazy

    out: Dict[str, str] = {}
    for lang in targets:
        translator = hf_pipeline("translation", model=f"Helsinki-NLP/opus-mt-en-{lang}")
        out[lang] = translator(text)[0]["translation_text"]
    return out

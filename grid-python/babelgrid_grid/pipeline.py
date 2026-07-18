"""The transcription pipeline: tokens -> text -> translations."""

from __future__ import annotations

from typing import Dict, List

from .mt import translate
from .stt import transcribe


def run_pipeline(tokens: List[int], targets: List[str] | None = None) -> Dict:
    text = transcribe(tokens)
    translations = translate(text, targets) if text else {}
    return {"text": text, "translations": translations}

"""Mock speech-to-text.

Offline there is no acoustic model, so this is a deterministic map from the
DSP's per-segment tokens to words in the fixed vocabulary. It is honest about
being a stand-in — set ``BABELGRID_STT=whisper`` to route audio to a real model
via ``real.py``. The value here is that the *plumbing* (DSP -> tokens -> words
-> translations) is exercised end to end with stable, testable output.
"""

from __future__ import annotations

from typing import List

from .vocab import WORD_BANK


def transcribe(tokens: List[int]) -> str:
    """Map a sequence of DSP tokens to a deterministic phrase."""
    if not tokens:
        return ""
    words = [WORD_BANK[int(t) % len(WORD_BANK)] for t in tokens]
    return " ".join(words)

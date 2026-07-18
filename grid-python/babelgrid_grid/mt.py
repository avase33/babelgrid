"""Phrase-table machine translation (en -> es/fr/de).

Word-by-word substitution with graceful fallback for out-of-vocabulary words.
``BABELGRID_MT=real`` can route to a neural MT backend via ``real.py``.
"""

from __future__ import annotations

import os
from typing import Dict, List

from .vocab import DICT, SUPPORTED


def _translate_one(text: str, target: str) -> str:
    table = DICT.get(target, {})
    return " ".join(table.get(w, w) for w in text.split())


def translate(text: str, targets: List[str] | None = None) -> Dict[str, str]:
    """Translate `text` into each requested target language."""
    langs = targets or list(SUPPORTED)

    if os.getenv("BABELGRID_MT") == "real":
        try:
            from .real import translate_real

            return translate_real(text, langs)
        except Exception:
            pass  # fall back to the phrase table

    return {lang: _translate_one(text, lang) for lang in langs if lang}

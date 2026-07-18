#!/usr/bin/env python3
"""Offline end-to-end verifier for babelgrid's transcription grid.

Runs the STT -> MT pipeline with deterministic tokens and checks the phrase
table — no services, no models. Exits non-zero on any failure.
"""

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(ROOT / "grid-python"))

PASS, FAIL = 0, 0


def check(name, cond):
    global PASS, FAIL
    if cond:
        PASS += 1
        print(f"  ok   {name}")
    else:
        FAIL += 1
        print(f"  FAIL {name}")


def main() -> int:
    from babelgrid_grid.mt import translate
    from babelgrid_grid.pipeline import run_pipeline
    from babelgrid_grid.stt import transcribe
    from babelgrid_grid.vocab import WORD_BANK

    print("babelgrid offline verify")

    check("transcribe deterministic", transcribe([0, 1, 2]) == transcribe([0, 1, 2]))
    check("transcribe length matches tokens", len(transcribe([0, 1, 2]).split()) == 3)
    check("token modulo wraps", transcribe([0]) == transcribe([len(WORD_BANK)]))

    tr = translate("the team shipped", ["es", "fr", "de"])
    check("es translation", tr["es"] == "el equipo envió")
    check("fr translation", tr["fr"] == "le équipe livré")
    check("de translation", tr["de"] == "die team lieferte")
    check("oov fallback", translate("the zebra", ["es"])["es"] == "el zebra")

    result = run_pipeline([0, 5, 6], ["es"])
    check("pipeline text", result["text"] == "the team shipped")
    check("pipeline translation", result["translations"]["es"] == "el equipo envió")

    print(f"\nRESULT: {PASS} passed, {FAIL} failed")
    return 1 if FAIL else 0


if __name__ == "__main__":
    raise SystemExit(main())

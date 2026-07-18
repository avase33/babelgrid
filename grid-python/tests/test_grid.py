from babelgrid_grid.mt import translate
from babelgrid_grid.pipeline import run_pipeline
from babelgrid_grid.stt import transcribe
from babelgrid_grid.vocab import WORD_BANK


def test_transcribe_deterministic_and_lengths():
    a = transcribe([0, 1, 2])
    b = transcribe([0, 1, 2])
    assert a == b
    assert len(a.split()) == 3
    assert transcribe([]) == ""


def test_transcribe_token_modulo_wraps():
    n = len(WORD_BANK)
    assert transcribe([0]) == transcribe([n])  # wraps around the bank


def test_translate_known_words():
    tr = translate("the team shipped", ["es", "de"])
    assert tr["es"] == "el equipo envió"
    assert tr["de"] == "die team lieferte"


def test_translate_falls_back_on_unknown():
    tr = translate("the zebra", ["es"])
    assert tr["es"] == "el zebra"  # unknown word passes through


def test_pipeline_end_to_end():
    result = run_pipeline([0, 5, 6], ["fr"])
    assert result["text"] == "the team shipped"
    assert result["translations"]["fr"] == "le équipe livré"
    assert run_pipeline([], ["fr"])["translations"] == {}

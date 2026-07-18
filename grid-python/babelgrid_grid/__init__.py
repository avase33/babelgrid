"""babelgrid transcription grid: mock STT + phrase-table MT."""

from .stt import transcribe
from .mt import translate
from .pipeline import run_pipeline

__all__ = ["transcribe", "translate", "run_pipeline"]
__version__ = "0.1.0"

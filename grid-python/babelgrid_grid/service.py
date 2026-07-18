"""FastAPI transcription grid: /transcribe, /translate, /pipeline."""

from __future__ import annotations

from typing import Dict, List

from fastapi import FastAPI
from pydantic import BaseModel

from .mt import translate
from .pipeline import run_pipeline
from .stt import transcribe

app = FastAPI(title="babelgrid grid", version="0.1.0")


class TranscribeRequest(BaseModel):
    tokens: List[int] = []


class TranslateRequest(BaseModel):
    text: str = ""
    targets: List[str] = []


class PipelineRequest(BaseModel):
    tokens: List[int] = []
    targets: List[str] = []


@app.get("/healthz")
def healthz() -> dict:
    return {"status": "ok"}


@app.post("/transcribe")
def transcribe_ep(req: TranscribeRequest) -> Dict[str, str]:
    return {"text": transcribe(req.tokens)}


@app.post("/translate")
def translate_ep(req: TranslateRequest) -> Dict[str, Dict[str, str]]:
    return {"translations": translate(req.text, req.targets or None)}


@app.post("/pipeline")
def pipeline_ep(req: PipelineRequest) -> Dict:
    return run_pipeline(req.tokens, req.targets or None)

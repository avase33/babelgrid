.PHONY: up down test test-rust test-go test-python build-web demo verify

up:
	docker compose up --build

down:
	docker compose down

test: test-rust test-go test-python

test-rust:
	cd dsp-rust && cargo test

test-go:
	cd sfu-go && go test ./...

test-python:
	cd grid-python && pip install -e ".[dev]" && pytest -q

build-web:
	cd client-ts && npm install && npm run build

# Offline: run tokens through the transcription + translation pipeline.
demo:
	cd grid-python && python -m babelgrid_grid.cli demo

verify:
	python scripts/verify.py

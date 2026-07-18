"""babelgrid-grid CLI: `demo`, `serve`."""

from __future__ import annotations

import argparse
import sys


def main(argv: list[str] | None = None) -> int:
    argv = argv if argv is not None else sys.argv[1:]
    parser = argparse.ArgumentParser(prog="babelgrid-grid")
    sub = parser.add_subparsers(dest="cmd", required=True)
    sub.add_parser("demo", help="run tokens through the transcription pipeline")
    p_serve = sub.add_parser("serve", help="run the FastAPI grid server")
    p_serve.add_argument("--host", default="0.0.0.0")
    p_serve.add_argument("--port", type=int, default=8000)

    args = parser.parse_args(argv)

    if args.cmd == "demo":
        from .pipeline import run_pipeline

        result = run_pipeline([0, 1, 2, 11, 12], ["es", "fr", "de"])
        print(f"transcript: {result['text']!r}")
        for lang, text in result["translations"].items():
            print(f"  {lang}: {text}")
        return 0

    if args.cmd == "serve":
        import uvicorn

        uvicorn.run("babelgrid_grid.service:app", host=args.host, port=args.port)
        return 0

    return 1


if __name__ == "__main__":
    raise SystemExit(main())

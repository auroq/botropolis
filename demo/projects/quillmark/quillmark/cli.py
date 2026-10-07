"""Command line entry point: ``quillmark build`` and ``quillmark new``."""

import argparse
import datetime as dt
import sys
from pathlib import Path

from quillmark import __version__, frontmatter
from quillmark.site import build, slugify


def _cmd_build(args: argparse.Namespace) -> int:
    try:
        written = build(Path(args.root), Path(args.out) if args.out else None)
    except frontmatter.FrontMatterError as err:
        print(f"quillmark: {err}", file=sys.stderr)
        return 1
    print(f"wrote {len(written)} files")
    return 0


def _cmd_new(args: argparse.Namespace) -> int:
    content = Path(args.root) / "content"
    content.mkdir(parents=True, exist_ok=True)
    path = content / f"{slugify(args.title)}.md"
    if path.exists():
        print(f"quillmark: {path} already exists", file=sys.stderr)
        return 1
    today = args.date or dt.date.today().isoformat()
    path.write_text(f"---\ntitle: {args.title}\ndate: {today}\n---\n\n", encoding="utf-8")
    print(path)
    return 0


def parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(prog="quillmark", description=__doc__)
    p.add_argument("--version", action="version", version=f"quillmark {__version__}")
    p.add_argument("--root", default=".", help="site directory (default: .)")
    sub = p.add_subparsers(dest="command", required=True)

    b = sub.add_parser("build", help="render content/ into public/")
    b.add_argument("--out", help="output directory (default: <root>/public)")
    b.set_defaults(func=_cmd_build)

    n = sub.add_parser("new", help="start a new post")
    n.add_argument("title")
    n.add_argument("--date", help="ISO date for the post (default: today)")
    n.set_defaults(func=_cmd_new)
    return p


def main(argv: list[str] | None = None) -> int:
    args = parser().parse_args(argv)
    return args.func(args)

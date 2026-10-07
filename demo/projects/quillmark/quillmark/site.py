"""Turn a content directory into a built site."""

import html
import re
import tomllib
from dataclasses import dataclass, field
from pathlib import Path
from string import Template

from quillmark import frontmatter, markdown

DEFAULT_CONFIG = {"title": "Untitled site", "base_url": "", "author": ""}


@dataclass
class Post:
    slug: str
    title: str
    date: str
    html: str
    meta: dict[str, str] = field(default_factory=dict)


def slugify(text: str) -> str:
    slug = re.sub(r"[^a-z0-9]+", "-", text.lower()).strip("-")
    return slug or "untitled"


def load_config(root: Path) -> dict[str, str]:
    config = dict(DEFAULT_CONFIG)
    path = root / "site.toml"
    if path.exists():
        with path.open("rb") as fh:
            config.update(tomllib.load(fh))
    return config


def load_post(path: Path) -> Post:
    meta, body = frontmatter.split(path.read_text(encoding="utf-8"))
    title = meta.get("title") or _first_heading(body) or path.stem.replace("-", " ").title()
    return Post(
        slug=meta.get("slug") or slugify(path.stem),
        title=title,
        date=meta.get("date", ""),
        html=markdown.to_html(body),
        meta=meta,
    )


def _first_heading(body: str) -> str | None:
    for line in body.splitlines():
        if line.startswith("# "):
            return line[2:].strip()
    return None


def load_posts(content: Path) -> list[Post]:
    posts = [load_post(p) for p in sorted(content.glob("*.md"))]
    return sorted(posts, key=lambda p: p.date, reverse=True)


def render_index(template: Template, config: dict[str, str], posts: list[Post]) -> str:
    items = "\n".join(
        f'    <li><a href="{p.slug}.html">{html.escape(p.title)}</a>'
        f' <time datetime="{p.date}">{p.date}</time></li>'
        for p in posts
    )
    return template.substitute(site_title=html.escape(config["title"]), posts=items)


def render_post(template: Template, config: dict[str, str], post: Post) -> str:
    return template.substitute(
        site_title=html.escape(config["title"]),
        title=html.escape(post.title),
        date=post.date,
        author=html.escape(post.meta.get("author", config["author"])),
        content=post.html,
    )


def build(root: Path, out: Path | None = None) -> list[Path]:
    root = Path(root)
    out = Path(out) if out else root / "public"
    config = load_config(root)
    templates = root / "templates"
    post_tpl = Template((templates / "post.html").read_text(encoding="utf-8"))
    index_tpl = Template((templates / "index.html").read_text(encoding="utf-8"))

    posts = load_posts(root / "content")
    out.mkdir(parents=True, exist_ok=True)

    written = []
    for post in posts:
        target = out / f"{post.slug}.html"
        target.write_text(render_post(post_tpl, config, post), encoding="utf-8")
        written.append(target)

    index = out / "index.html"
    index.write_text(render_index(index_tpl, config, posts), encoding="utf-8")
    written.append(index)
    return written

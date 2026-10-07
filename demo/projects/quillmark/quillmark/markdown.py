"""A deliberately small Markdown subset.

Supported: ATX headings, paragraphs, **strong**, *emphasis* / _emphasis_,
`code spans`, [links](url), fenced code blocks, and flat ordered or
unordered lists. Anything else is treated as paragraph text.
"""

import html
import re

_HEADING = re.compile(r"^(#{1,6})\s+(.+?)\s*#*\s*$")
_ULIST = re.compile(r"^\s*[-*+]\s+(.*)$")
_OLIST = re.compile(r"^\s*\d+[.)]\s+(.*)$")
_FENCE = re.compile(r"^```\s*([\w+-]*)\s*$")
_LINK = re.compile(r"\[([^\]]+)\]\(([^)\s]+)\)")
_CODE = re.compile(r"`([^`]+)`")
_STRONG = re.compile(r"\*\*(.+?)\*\*")
_EM = re.compile(r"\*(.+?)\*|_(.+?)_")


def inline(text: str) -> str:
    out = html.escape(text, quote=False)
    out = _CODE.sub(lambda m: f"<code>{m.group(1)}</code>", out)
    out = _LINK.sub(lambda m: f'<a href="{m.group(2)}">{m.group(1)}</a>', out)
    out = _STRONG.sub(r"<strong>\1</strong>", out)
    out = _EM.sub(lambda m: f"<em>{m.group(1) or m.group(2)}</em>", out)
    return out


def to_html(source: str) -> str:
    blocks: list[str] = []
    paragraph: list[str] = []
    list_tag: str | None = None
    list_items: list[str] = []
    in_fence = False
    fence_lang = ""
    fence_lines: list[str] = []

    def flush_paragraph() -> None:
        if paragraph:
            blocks.append(f"<p>{inline(' '.join(paragraph))}</p>")
            paragraph.clear()

    def flush_list() -> None:
        nonlocal list_tag
        if list_tag:
            items = "".join(f"<li>{inline(item)}</li>" for item in list_items)
            blocks.append(f"<{list_tag}>{items}</{list_tag}>")
            list_items.clear()
            list_tag = None

    def flush_fence() -> None:
        attr = f' class="language-{fence_lang}"' if fence_lang else ""
        code = html.escape("\n".join(fence_lines), quote=False)
        blocks.append(f"<pre><code{attr}>{code}</code></pre>")
        fence_lines.clear()

    for line in source.splitlines():
        if in_fence:
            if line.strip() == "```":
                flush_fence()
                in_fence = False
            else:
                fence_lines.append(line)
            continue

        fence = _FENCE.match(line)
        if fence:
            flush_paragraph()
            flush_list()
            in_fence = True
            fence_lang = fence.group(1)
            continue

        if not line.strip():
            flush_paragraph()
            flush_list()
            continue

        heading = _HEADING.match(line)
        if heading:
            flush_paragraph()
            flush_list()
            level = len(heading.group(1))
            blocks.append(f"<h{level}>{inline(heading.group(2))}</h{level}>")
            continue

        for pattern, tag in ((_ULIST, "ul"), (_OLIST, "ol")):
            item = pattern.match(line)
            if item:
                flush_paragraph()
                if list_tag != tag:
                    flush_list()
                    list_tag = tag
                list_items.append(item.group(1))
                break
        else:
            if list_tag and line.startswith((" ", "\t")):
                list_items[-1] += " " + line.strip()
                continue
            flush_list()
            paragraph.append(line.strip())

    if in_fence:
        flush_fence()
    flush_paragraph()
    flush_list()
    return "\n".join(blocks) + ("\n" if blocks else "")

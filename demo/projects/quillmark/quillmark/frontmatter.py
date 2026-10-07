"""Split a document into its front-matter block and its body.

A front-matter block is a run of ``key: value`` lines fenced by ``---``
at the very top of the file.
"""

FENCE = "---"


class FrontMatterError(ValueError):
    pass


def split(text: str) -> tuple[dict[str, str], str]:
    if not text.startswith(FENCE + "\n"):
        return {}, text

    end = text.find("\n" + FENCE + "\n", len(FENCE))
    if end == -1:
        raise FrontMatterError("front matter is never closed with '---'")

    meta: dict[str, str] = {}
    for number, line in enumerate(text[len(FENCE) + 1 : end].splitlines(), start=2):
        if not line.strip() or line.lstrip().startswith("#"):
            continue
        key, sep, value = line.partition(":")
        if not sep or not key.strip():
            raise FrontMatterError(f"line {number}: expected 'key: value', got {line!r}")
        meta[key.strip().lower()] = value.strip()

    body = text[end + len(FENCE) + 2 :]
    return meta, body

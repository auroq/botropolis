from pathlib import Path

import pytest

from quillmark.site import build, slugify

ROOT = Path(__file__).resolve().parent.parent


@pytest.mark.parametrize(
    ("text", "expected"),
    [("Hello, World!", "hello-world"), ("  spaced  out ", "spaced-out"), ("???", "untitled")],
)
def test_slugify(text, expected):
    assert slugify(text) == expected


@pytest.fixture(scope="module")
def out(tmp_path_factory):
    out = tmp_path_factory.mktemp("public")
    build(ROOT, out)
    return out


class TestWhenBuildingTheSampleSite:
    def test_it_should_write_one_page_per_post_and_an_index(self, out):
        assert sorted(p.name for p in out.iterdir()) == [
            "hello-quillmark.html",
            "index.html",
            "tide-tables-by-hand.html",
        ]

    def test_it_should_list_the_newest_post_first(self, out):
        index = (out / "index.html").read_text()
        assert index.index("tide-tables-by-hand") < index.index("hello-quillmark")

    def test_it_should_use_the_site_title(self, out):
        assert "<title>Field Notes</title>" in (out / "index.html").read_text()

    def test_it_should_prefer_the_post_author_over_the_site_author(self, out):
        assert "M. Trevose" in (out / "tide-tables-by-hand.html").read_text()

    def test_it_should_escape_titles(self, out):
        assert "Hello, quillmark · Field Notes" in (out / "hello-quillmark.html").read_text()

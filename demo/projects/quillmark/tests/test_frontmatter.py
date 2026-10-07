import pytest

from quillmark.frontmatter import FrontMatterError, split


class TestWhenTheDocumentHasFrontMatter:
    text = "---\ntitle: Notes: part two\nDate: 2026-01-02\n# a comment\n\n---\nBody text\n"

    def test_it_should_keep_everything_after_the_first_colon_in_the_value(self):
        meta, _ = split(self.text)
        assert meta["title"] == "Notes: part two"

    def test_it_should_lowercase_keys(self):
        meta, _ = split(self.text)
        assert meta["date"] == "2026-01-02"

    def test_it_should_skip_comments_and_blank_lines(self):
        meta, _ = split(self.text)
        assert set(meta) == {"title", "date"}

    def test_it_should_return_the_body_without_the_block(self):
        _, body = split(self.text)
        assert body == "Body text\n"


class TestWhenTheDocumentHasNoFrontMatter:
    def test_it_should_return_the_whole_text_as_body(self):
        assert split("# Just a heading\n") == ({}, "# Just a heading\n")


class TestWhenTheFrontMatterIsMalformed:
    @pytest.mark.parametrize(
        "text",
        [
            pytest.param("---\ntitle: x\nno closing fence\n", id="unclosed"),
            pytest.param("---\ntitle x\n---\nbody\n", id="no-colon"),
            pytest.param("---\n: x\n---\nbody\n", id="empty-key"),
        ],
    )
    def test_it_should_raise(self, text):
        with pytest.raises(FrontMatterError):
            split(text)

import pytest

from quillmark.markdown import inline, to_html


class TestInline:
    @pytest.mark.parametrize(
        ("source", "expected"),
        [
            ("plain", "plain"),
            ("**bold**", "<strong>bold</strong>"),
            ("*soft*", "<em>soft</em>"),
            ("_soft_", "<em>soft</em>"),
            ("`code`", "<code>code</code>"),
            ("[home](https://example.org/)", '<a href="https://example.org/">home</a>'),
            ("a < b & c", "a &lt; b &amp; c"),
        ],
    )
    def test_when_given_a_span_it_should_render_it(self, source, expected):
        assert inline(source) == expected


class TestBlocks:
    @pytest.mark.parametrize(
        ("source", "expected"),
        [
            ("# Title", "<h1>Title</h1>\n"),
            ("### Third ###", "<h3>Third</h3>\n"),
            ("one\ntwo\n\nthree", "<p>one two</p>\n<p>three</p>\n"),
            ("- a\n- b", "<ul><li>a</li><li>b</li></ul>\n"),
            ("1. a\n2. b", "<ol><li>a</li><li>b</li></ol>\n"),
            ("- a\n  continued", "<ul><li>a continued</li></ul>\n"),
            ("- a\n1. b", "<ul><li>a</li></ul>\n<ol><li>b</li></ol>\n"),
            ("", ""),
        ],
    )
    def test_when_given_a_block_it_should_render_it(self, source, expected):
        assert to_html(source) == expected


class TestWhenTheSourceHasAFencedCodeBlock:
    source = "intro\n```python\nif a < b:\n    *x* = 1\n```\nafter"

    def test_it_should_tag_the_language(self):
        assert '<code class="language-python">' in to_html(self.source)

    def test_it_should_escape_and_not_format_the_contents(self):
        assert "if a &lt; b:\n    *x* = 1</code>" in to_html(self.source)

    def test_it_should_close_the_paragraph_before_it(self):
        assert to_html(self.source).startswith("<p>intro</p>\n<pre>")

    def test_and_the_fence_is_never_closed_it_should_still_emit_the_block(self):
        assert to_html("```\nleft open").endswith("<pre><code>left open</code></pre>\n")

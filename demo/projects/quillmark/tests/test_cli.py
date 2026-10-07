from quillmark.cli import main


class TestWhenStartingANewPost:
    def test_it_should_write_front_matter_with_the_title_and_date(self, tmp_path):
        main(["--root", str(tmp_path), "new", "Fog Signals", "--date", "2026-10-01"])
        text = (tmp_path / "content" / "fog-signals.md").read_text()
        assert text == "---\ntitle: Fog Signals\ndate: 2026-10-01\n---\n\n"

    def test_and_the_post_exists_it_should_refuse(self, tmp_path):
        main(["--root", str(tmp_path), "new", "Twice"])
        assert main(["--root", str(tmp_path), "new", "Twice"]) == 1


class TestWhenAPostHasBrokenFrontMatter:
    def test_it_should_exit_nonzero(self, tmp_path, capsys):
        (tmp_path / "content").mkdir()
        (tmp_path / "templates").mkdir()
        for name in ("post.html", "index.html"):
            (tmp_path / "templates" / name).write_text("$site_title")
        (tmp_path / "content" / "bad.md").write_text("---\ntitle: never closed\n")
        assert main(["--root", str(tmp_path), "build"]) == 1

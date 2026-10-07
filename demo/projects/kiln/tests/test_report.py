from pathlib import Path

import pytest

from kiln.__main__ import main
from kiln.report import format_text, summarize

SAMPLES = Path(__file__).parent.parent / "samples"


class TestFormatTextWhenTheFiringIsClean:
    @pytest.fixture
    def text(self, firing):
        return format_text(summarize(firing))

    def test_it_should_name_the_cone(self, text):
        assert "cone 04" in text

    def test_it_should_report_the_peak(self, text):
        assert "Peak     1000 °C at 06:52" in text

    def test_it_should_not_print_an_anomalies_section(self, text):
        assert "Anomalies" not in text


class TestMainWhenReportingTheBisqueSample:
    @pytest.fixture
    def run(self, capsys):
        code = main(["report", str(SAMPLES / "bisque-0914.csv")])
        return code, capsys.readouterr()

    def test_it_should_succeed(self, run):
        assert run[0] == 0

    def test_it_should_flag_the_thermocouple_glitch(self, run):
        assert "spike  09:23" in run[1].out


def test_when_the_file_does_not_exist_main_should_fail_politely(tmp_path, capsys):
    assert main(["report", str(tmp_path / "missing.csv")]) == 1

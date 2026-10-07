from datetime import datetime

import pytest

from kiln.reader import LogError, parse

HEADER = "timestamp,temperature_c,cone\n"


class TestParseWhenTheLogIsWellFormed:
    @pytest.fixture
    def readings(self):
        return parse(
            [
                HEADER,
                "2026-09-14 06:01:00,24.5,04\n",
                "2026-09-14 06:00:00,21.0,04\n",
                "2026-09-14 06:02:00,28.0,\n",
            ]
        )

    def test_it_should_read_every_row(self, readings):
        assert len(readings) == 3

    def test_it_should_sort_by_time(self, readings):
        assert readings[0].at == datetime(2026, 9, 14, 6, 0)

    def test_it_should_read_the_temperature(self, readings):
        assert readings[1].temp_c == 24.5

    def test_and_the_cone_is_blank_it_should_be_none(self, readings):
        assert readings[2].cone is None


@pytest.mark.parametrize(
    "lines, message",
    [
        (["timestamp,temp\n"], "missing columns: temperature_c, cone"),
        ([HEADER, "yesterday,20,04\n"], "bad timestamp 'yesterday'"),
        ([HEADER, "2026-09-14 06:00:00,hot,04\n"], "bad temperature 'hot'"),
    ],
)
def test_when_the_log_is_malformed_it_should_say_why(lines, message):
    with pytest.raises(LogError, match=message):
        parse(lines)

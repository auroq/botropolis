from datetime import timedelta

import pytest

from kiln.analysis import anomalies, hold_time, peak, segments
from tests.conftest import build


class TestSegmentsWhenTheFiringRisesHoldsAndFalls:
    @pytest.fixture
    def segs(self, firing):
        return segments(firing)

    def test_it_should_find_three_phases(self, segs):
        assert [s.phase for s in segs] == ["heating", "holding", "cooling"]

    def test_it_should_measure_the_heating_rate(self, segs):
        assert segs[0].rate_c_per_hour == pytest.approx(120.0, rel=0.05)

    def test_it_should_measure_the_cooling_rate(self, segs):
        assert segs[2].rate_c_per_hour == pytest.approx(-180.0, rel=0.1)


def test_when_a_log_has_one_reading_it_should_have_no_segments():
    assert segments(build([500.0])) == []


def test_peak_should_be_the_hottest_reading(firing):
    assert peak(firing).temp_c == 1000.4


def test_hold_time_should_count_readings_within_tolerance_of_peak(firing):
    assert hold_time(firing) == timedelta(minutes=8)


class TestAnomalies:
    def test_when_the_firing_is_smooth_it_should_find_none(self, firing):
        assert anomalies(firing) == []

    def test_when_a_reading_jumps_it_should_report_a_spike(self):
        found = anomalies(build([500.0, 502.0, 590.0, 504.0]))
        assert [a.kind for a in found] == ["spike", "spike"]

    def test_when_readings_stop_for_a_while_it_should_report_a_gap(self):
        readings = build([500.0, 502.0])
        readings[1] = readings[1].__class__(readings[0].at + timedelta(minutes=30), 503.0)
        assert [a.kind for a in anomalies(readings)] == ["gap"]

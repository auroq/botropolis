from dataclasses import dataclass
from datetime import timedelta
from typing import Literal

from kiln.reader import Reading

SAMPLE_INTERVAL = timedelta(minutes=1)

Phase = Literal["heating", "holding", "cooling"]


@dataclass(frozen=True)
class Segment:
    phase: Phase
    start: Reading
    end: Reading

    @property
    def duration(self) -> timedelta:
        return self.end.at - self.start.at

    @property
    def rate_c_per_hour(self) -> float:
        hours = self.duration.total_seconds() / 3600
        if hours == 0:
            return 0.0
        return (self.end.temp_c - self.start.temp_c) / hours


@dataclass(frozen=True)
class Anomaly:
    kind: Literal["spike", "gap"]
    at: Reading
    detail: str


def _phase(a: Reading, b: Reading, flat_c_per_hour: float) -> Phase:
    hours = (b.at - a.at).total_seconds() / 3600
    rate = (b.temp_c - a.temp_c) / hours if hours else 0.0
    if rate > flat_c_per_hour:
        return "heating"
    if rate < -flat_c_per_hour:
        return "cooling"
    return "holding"


def segments(readings: list[Reading], flat_c_per_hour: float = 15.0) -> list[Segment]:
    out: list[Segment] = []
    for a, b in zip(readings, readings[1:]):
        phase = _phase(a, b, flat_c_per_hour)
        if out and out[-1].phase == phase:
            out[-1] = Segment(phase, out[-1].start, b)
        else:
            out.append(Segment(phase, a, b))
    return out


def peak(readings: list[Reading]) -> Reading:
    return max(readings, key=lambda r: r.temp_c)


def hold_time(readings: list[Reading], tolerance_c: float = 5.0) -> timedelta:
    top = peak(readings).temp_c
    near = sum(1 for r in readings if r.temp_c >= top - tolerance_c)
    return SAMPLE_INTERVAL * near


def anomalies(
    readings: list[Reading],
    max_jump_c: float = 40.0,
    max_gap: timedelta = timedelta(minutes=10),
) -> list[Anomaly]:
    found = []
    for a, b in zip(readings, readings[1:]):
        jump = b.temp_c - a.temp_c
        if abs(jump) > max_jump_c:
            found.append(Anomaly("spike", b, f"{jump:+.0f} °C since {a.at:%H:%M}"))
        gap = b.at - a.at
        if gap > max_gap:
            found.append(Anomaly("gap", b, f"no readings for {gap}"))
    return found

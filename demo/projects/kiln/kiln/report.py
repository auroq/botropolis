from dataclasses import dataclass
from datetime import timedelta

from kiln.analysis import Anomaly, Segment, anomalies, hold_time, peak, segments
from kiln.reader import Reading


@dataclass(frozen=True)
class Summary:
    started: Reading
    finished: Reading
    peak: Reading
    hold: timedelta
    segments: list[Segment]
    anomalies: list[Anomaly]

    @property
    def cone(self) -> str | None:
        return self.peak.cone or self.started.cone


def summarize(readings: list[Reading]) -> Summary:
    return Summary(
        started=readings[0],
        finished=readings[-1],
        peak=peak(readings),
        hold=hold_time(readings),
        segments=segments(readings),
        anomalies=anomalies(readings),
    )


def _hm(d: timedelta) -> str:
    minutes = int(d.total_seconds() // 60)
    return f"{minutes // 60}h{minutes % 60:02d}m"


def format_text(s: Summary) -> str:
    lines = [
        f"Firing {s.started.at:%Y-%m-%d %H:%M} to {s.finished.at:%H:%M}"
        + (f", cone {s.cone}" if s.cone else ""),
        f"Peak     {s.peak.temp_c:.0f} °C at {s.peak.at:%H:%M}",
        f"Hold     {_hm(s.hold)} within 5 °C of peak",
        "",
        "Segments",
    ]
    for seg in s.segments:
        lines.append(
            f"  {seg.phase:<8} {seg.start.at:%H:%M}-{seg.end.at:%H:%M}"
            f"  {seg.start.temp_c:5.0f} -> {seg.end.temp_c:5.0f} °C"
            f"  {seg.rate_c_per_hour:+6.0f} °C/h"
        )
    if s.anomalies:
        lines += ["", "Anomalies"]
        lines += [f"  {a.kind:<6} {a.at.at:%H:%M}  {a.detail}" for a in s.anomalies]
    return "\n".join(lines) + "\n"

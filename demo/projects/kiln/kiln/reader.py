import csv
from collections.abc import Iterable
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

COLUMNS = ("timestamp", "temperature_c", "cone")


class LogError(ValueError):
    pass


@dataclass(frozen=True)
class Reading:
    at: datetime
    temp_c: float
    cone: str | None = None


def read_log(path: str | Path) -> list[Reading]:
    with open(path, newline="", encoding="utf-8") as f:
        return parse(f)


def parse(lines: Iterable[str]) -> list[Reading]:
    rows = csv.DictReader(lines)
    missing = [c for c in COLUMNS if c not in (rows.fieldnames or [])]
    if missing:
        raise LogError(f"missing columns: {', '.join(missing)}")

    readings = []
    for row in rows:
        try:
            at = datetime.fromisoformat(row["timestamp"].strip())
        except ValueError:
            raise LogError(f"bad timestamp {row['timestamp']!r}") from None
        try:
            temp = float(row["temperature_c"])
        except ValueError:
            raise LogError(f"bad temperature {row['temperature_c']!r}") from None
        cone = (row["cone"] or "").strip() or None
        readings.append(Reading(at=at, temp_c=temp, cone=cone))

    readings.sort(key=lambda r: r.at)
    return readings

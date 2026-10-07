from datetime import datetime, timedelta

import pytest

from kiln.reader import Reading

START = datetime(2026, 9, 14, 6, 0)


def build(temps: list[float], every: timedelta = timedelta(minutes=1), cone: str | None = "04"):
    return [Reading(at=START + i * every, temp_c=t, cone=cone) for i, t in enumerate(temps)]


@pytest.fixture
def firing():
    rise = [900.0 + 2 * i for i in range(50)]
    hold = [1000.0, 1000.2, 1000.4, 1000.2, 1000.0]
    fall = [998.0 - 3 * i for i in range(30)]
    return build(rise + hold + fall)

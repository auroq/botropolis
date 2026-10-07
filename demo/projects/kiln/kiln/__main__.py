import argparse
import sys

from kiln.reader import LogError, read_log
from kiln.report import format_text, summarize


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(prog="kiln", description="Summarise a kiln firing log.")
    sub = parser.add_subparsers(dest="command", required=True)
    report = sub.add_parser("report", help="print a summary of one firing")
    report.add_argument("log", help="CSV with timestamp, temperature_c and cone columns")
    args = parser.parse_args(argv)

    try:
        readings = read_log(args.log)
    except (OSError, LogError) as e:
        print(f"kiln: {args.log}: {e}", file=sys.stderr)
        return 1

    sys.stdout.write(format_text(summarize(readings)))
    return 0


if __name__ == "__main__":
    sys.exit(main())

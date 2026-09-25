#!/usr/bin/env python3
"""What the atlas costs, and what dropping pieces from it would save.

Answers "is this piece worth its page?" without a twenty-minute Blender
run: it replays the pipeline's own shelf packer over the sprite sizes in
the shipped manifests. The check that it is replaying rather than
approximating is that it reproduces the shipped page count exactly, which
it prints first — if that line says MISMATCH the packer has changed and
this tool is lying.

  python3 tools/atlas-cost.py              # today, and every piece's status
  python3 tools/atlas-cost.py --drop-unused  # what dropping the never-drawn ones saves

A piece counts as drawn if its full name or its bare name appears as a
literal anywhere in the tracked .go files, which is how pkg/render names
them — there is no runtime construction of piece names.
"""
import argparse, json, os, re, subprocess

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def pipeline():
    src = open(os.path.join(REPO, "tools/render-sprites/render.py")).read()
    page = int(re.search(r"^PAGE = (\d+)", src, re.M).group(1))
    pad = int(re.search(r"^PAD = (\d+)", src, re.M).group(1))
    ns = {}
    exec(re.search(r"^PIECES = \{.*?^\}", src, re.M | re.S).group(0), ns)
    order = [f"{kit}/{n}" for kit, names in ns["PIECES"].items() for n in names]
    return page, pad, order


def drawn_names():
    go = subprocess.run(["bash", "-c", "cat $(git ls-files '*.go')"], cwd=REPO,
                        capture_output=True, text=True).stdout
    return lambda n: f'"{n}"' in go or f'"{n.split("/")[-1]}"' in go


def pack(sizes, page, pad):
    """The pipeline's shelf packer, counting pages rather than filling them."""
    pages, shelf_y, shelf_h, x = 1, 0, 0, 0
    for w, h in sizes:
        if x + w + pad > page:
            shelf_y += shelf_h + pad
            shelf_h = x = 0
        if shelf_y + h + pad > page:
            pages += 1
            shelf_y = shelf_h = x = 0
        x += w + pad
        shelf_h = max(shelf_h, h)
    return pages


def main():
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--drop-unused", action="store_true",
                    help="also price the atlas without the pieces nothing draws")
    args = ap.parse_args()
    page, pad, order = pipeline()
    drawn = drawn_names()
    kits = os.path.join(REPO, "pkg/assets/kits")
    total_now = total_lean = 0
    for z in ("z1", "z2"):
        sp = json.load(open(os.path.join(kits, f"kits-{z}.json")))["sprites"]
        def sizes(keep):
            return [(sp[n][h]["w"], sp[n][h]["h"]) for n in order if n in sp and keep(n)
                    for h in ("0", "90", "180", "270")]
        now = pack(sizes(lambda n: True), page, pad)
        shipped = len([f for f in os.listdir(kits) if f.startswith(f"kits-{z}-")])
        total_now += now
        note = "replays the shipped atlas" if now == shipped else f"MISMATCH — shipped {shipped}"
        print(f"{z}: {now} pages ({note})")
        if args.drop_unused:
            lean = pack(sizes(drawn), page, pad)
            total_lean += lean
            print(f"    without the never-drawn pieces: {lean} pages ({now - lean} fewer)")
    mb = sum(os.path.getsize(os.path.join(kits, f)) for f in os.listdir(kits)
             if f.endswith(".png")) / 1e6
    print(f"\n{mb:.1f} MB over {total_now} pages")
    if args.drop_unused:
        print(f"{mb * total_lean / total_now:.1f} MB over {total_lean} pages if the never-drawn go")
    never = [n for n in order if not drawn(n)]
    print(f"\nnever drawn: {len(never)} of {len(order)} pieces")
    for n in never:
        print("   ", n)


if __name__ == "__main__":
    main()

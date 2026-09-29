import glob, os, re, subprocess, sys

fail = []

def note(check, msg):
    fail.append('%s: %s' % (check, msg))

docs = ['DESIGN.md', 'ROADMAP.md', 'CLAUDE.md', 'README.md']
docs += sorted(glob.glob('design/*.md')) + sorted(glob.glob('rules/*.md'))
docs += sorted(glob.glob('roadmap/*.md'))

# 1. links resolve from the linking file's own directory
links = 0
for p in docs:
    base = os.path.dirname(p) or '.'
    for m in re.finditer(r'\]\(([^)]+)\)', open(p, encoding='utf-8').read()):
        t = m.group(1).split('#')[0]
        if not t or t.startswith(('http', 'mailto')):
            continue
        links += 1
        if not os.path.exists(os.path.normpath(os.path.join(base, t))):
            note('links', '%s -> %s' % (p, t))

# 2. no stale section citations to the pre-split ROADMAP
NOTED = 'roadmap/inventory-2026-09-29.md'
for p in docs:
    if p in (NOTED, 'ROADMAP.md'):
        continue
    for i, line in enumerate(open(p, encoding='utf-8'), 1):
        for m in re.finditer(r'§\d+', line):
            note('citations', '%s:%d %s refers to a section that no longer exists'
                 % (p, i, m.group(0)))

# 3. every item file declares a status
items = sorted(glob.glob('roadmap/[0-9][0-9][0-9]-*.md'))
status = {}
for p in items:
    L = open(p, encoding='utf-8').read().split('\n')
    if len(L) < 3 or not L[2].startswith('**'):
        note('status', '%s has no status line under its title' % p)
        continue
    status[p] = L[2].strip('*')

# 4. ROADMAP's open table matches the item files
declared = {p for p, s in status.items() if s.lower().startswith('open')}
roadmap = open('ROADMAP.md', encoding='utf-8').read()
listed = {os.path.normpath(os.path.join('.', m))
          for m in re.findall(r'\]\((roadmap/[0-9][0-9][0-9]-[^)]+)\)', roadmap)}
listed = {os.path.relpath(x) for x in listed}
for p in sorted(declared - listed):
    note('open-table', '%s is Open but ROADMAP.md does not list it' % p)
for p in sorted(listed - declared):
    note('open-table', '%s is listed in ROADMAP.md but its status is not Open (%s)'
         % (p, status.get(p, '?')[:60]))

# 5. the corpus is not empty, and no item file has gone missing.
# Checks 1, 3 and 4 all pass on nothing: an empty tree reports "0 links
# resolve, 0 items carry a status, open table matches" and exits 0,
# verified. A green that cannot tell "correct" from "nothing to check" is
# the sprites-check date-chunk shape in rules/a-green-test-is-not-a-guard.md.
# The floor is the numbering itself rather than a constant, so it rises
# with the work instead of going stale: every number from 1 to the highest
# present must have a file, which fails loudly if any file disappears and
# needs no maintenance when the next item is filed.
for index in ('DESIGN.md', 'ROADMAP.md', 'CLAUDE.md', 'README.md'):
    if not open(index, encoding='utf-8').read().strip():
        note('corpus', '%s is empty' % index)

# The ceiling of that range comes from git, not from the working tree.
# Derived from the tree alone it falls with the damage: deleting the two
# highest-numbered items lowers max() and the check reports "1-66, none
# missing" and exits 0, verified. git is an independent source that a
# working-tree deletion cannot move, and it still needs no maintenance --
# committing item 69 raises it, and committing a deliberate removal lowers
# it, which is the correct semantics for each.
# When git cannot be reached the bound falls back to the working tree,
# which silently restores the fault above: with git stubbed out, deleting
# 067 and 068 prints "1-66, none missing" and exits 0 again, verified --
# and the summary read identically to a healthy run, which is what made it
# worth fixing rather than accepting. A fallback in a checker is a way for
# it to pass (r312), so the fallback is not removed -- running outside a
# checkout is legitimate -- but it is no longer silent: the summary says
# which source the bound came from, so a green from the degraded path
# cannot be mistaken for a green from the checked one.
numbers = {int(os.path.basename(p)[:3]) for p in items}
tracked = set()
bound_from_git = False
try:
    out = subprocess.run(['git', 'ls-files', 'roadmap/'], capture_output=True,
                         text=True, check=True).stdout.split()
    tracked = {int(os.path.basename(f)[:3])
               for f in out if re.match(r'^\d{3}-', os.path.basename(f))}
    bound_from_git = bool(tracked)
except Exception:
    pass  # not a checkout; the working tree is all we have, and we say so

if not numbers:
    note('corpus', 'no roadmap/NNN-*.md files found at all')
else:
    top = max(numbers | tracked)
    for n in range(1, top + 1):
        if n not in numbers:
            where = 'tracked in git' if n in tracked else 'expected'
            note('corpus', 'item %03d has no file (%s; highest known is %03d)' % (n, where, top))

if fail:
    for f in fail:
        print('  %s' % f)
    print('\n%d problem(s).' % len(fail))
    sys.exit(1)
print('  %d links resolve, %d items carry a status (1-%d, none missing), open table matches.'
      % (links, len(status), max(numbers | tracked)))
if not bound_from_git:
    print('  NOTE: git was not reachable, so the range came from the working tree '
          'alone. A deletion of the highest-numbered items cannot be seen from here.')

#!/usr/bin/env python3
"""Copy a scrubbed slice of ~/.claude into testing/helpers/fixtures/<name>/.

The slice is every session with a live record in ~/.claude/sessions, the N most
recently touched transcripts, the smallest bridge-session stub, and the most
recent session that spawned subagents. Structure, ids, timestamps, and `usage`
are kept verbatim; prompt text, tool inputs and outputs, thinking, titles,
slugs, file snapshots, and account identifiers are replaced with deterministic
placeholders so the fixture is stable across regenerations and safe to commit.

Layout written under the fixture directory:

  home/.claude/sessions/<pid>.json
  home/.claude/projects/<cwd>/<sid>.jsonl
  home/.claude/projects/<cwd>/<sid>/subagents/*.jsonl, *.meta.json
  home/.claude/stats-cache.json
  home/.claude.json                 (mcpServers only)
  manifest.json
  raw/...                           (unscrubbed copy, gitignored, --no-raw to skip)

Read-only against ~/.claude; stdlib only.
"""
import argparse
import hashlib
import json
import os
import re
import shutil
import sys
from pathlib import Path

HOME = Path.home()
REPO = Path(__file__).resolve().parent.parent
UUID_RE = re.compile(r"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", re.I)

KEEP_KEYS = {
    "type", "subtype", "role", "model", "id", "uuid", "parentUuid", "leafUuid",
    "logicalParentUuid", "sessionId", "session_id", "messageId", "snapshotMessageId",
    "requestId", "tool_use_id", "toolUseID", "toolUseId", "sourceToolUseID",
    "sourceToolAssistantUUID", "interruptedMessageId", "promptId", "timestamp",
    "cwd", "relocatedCwd", "gitBranch", "version", "entrypoint", "userType",
    "sessionKind", "effort", "perTurnEffort", "permissionMode", "mode", "agentType",
    "attributionMcpServer", "attributionMcpTool", "attributionSkill",
    "attributionPlugin", "stop_reason", "stopReason", "service_tier", "kind",
    "status", "nameSource", "messagingSocketPath", "jobId", "startTime",
    "lastComputedDate", "date", "firstSessionDate", "operation", "trigger", "level",
    "scope", "direction", "prRepository", "prUrl", "origin", "promptSource",
    "queueOrigin", "teamName", "cronKind", "atis", "agentSetting", "toolDenialKind",
    "procStart", "startedAt", "updatedAt", "statusUpdatedAt", "nameSince", "until",
    "peerFeatures", "ts", "fallbackModel", "originalModel", "apiRefusalCategory",
}
TOOL_PART_TYPES = {"tool_use", "server_tool_use", "mcp_tool_use"}
FORCE_SCRUB_KEYS = {"bridgeSessionId", "pidDomain", "ownerAccountUuid", "ownerOrganizationUuid"}


def placeholder(s):
    if s == "":
        return ""
    digest = hashlib.sha1(s.encode("utf-8", "replace")).hexdigest()
    if UUID_RE.match(s):
        h = digest[:32]
        return f"{h[:8]}-{h[8:12]}-{h[12:16]}-{h[16:20]}-{h[20:32]}"
    return f"<scrubbed:{digest[:12]}>"


def scrub(obj, key=None, parent_type=None):
    if isinstance(obj, dict):
        own_type = obj.get("type") if isinstance(obj.get("type"), str) else parent_type
        out = {}
        for k, v in obj.items():
            out[k] = scrub(v, k, None if k == "input" else own_type)
        return out
    if isinstance(obj, list):
        return [scrub(v, key, parent_type) for v in obj]
    if isinstance(obj, str):
        if key in FORCE_SCRUB_KEYS:
            return placeholder(obj)
        if key in KEEP_KEYS:
            return obj
        if key == "name" and parent_type in TOOL_PART_TYPES:
            return obj
        if key == "input_schema":
            return obj
        return placeholder(obj)
    return obj


def scrub_jsonl(src, dst):
    dst.parent.mkdir(parents=True, exist_ok=True)
    with open(src, encoding="utf-8", errors="replace") as fin, open(dst, "w", encoding="utf-8") as fout:
        for line in fin:
            line = line.rstrip("\n")
            if not line:
                continue
            try:
                rec = json.loads(line)
            except json.JSONDecodeError:
                fout.write(placeholder(line) + "\n")
                continue
            fout.write(json.dumps(scrub(rec), separators=(",", ":"), ensure_ascii=False) + "\n")


def scrub_json_file(src, dst, fn=scrub):
    dst.parent.mkdir(parents=True, exist_ok=True)
    with open(src, encoding="utf-8") as f:
        data = json.load(f)
    with open(dst, "w", encoding="utf-8") as f:
        json.dump(fn(data), f, separators=(",", ":"), ensure_ascii=False)
        f.write("\n")


def scrub_session_record(rec):
    out = scrub(rec)
    if isinstance(rec.get("name"), str):
        out["name"] = placeholder(rec["name"])
    for former in out.get("formerNames") or []:
        if isinstance(former, dict) and isinstance(former.get("name"), str):
            former["name"] = placeholder(former["name"])
    return out


def scrub_mcp_server(cfg):
    out = {}
    for k, v in cfg.items():
        if k in ("type", "command"):
            out[k] = v
        elif k == "url" and isinstance(v, str):
            out[k] = re.sub(r"^([a-z]+://[^/?#@]+).*$", r"\1/<scrubbed>", v)
        elif k == "args" and isinstance(v, list):
            out[k] = ["<scrubbed>" for _ in v]
        elif isinstance(v, dict):
            out[k] = {kk: "<scrubbed>" for kk in v}
        else:
            out[k] = "<scrubbed>"
    return out


def scrub_claude_json(data):
    out = {"mcpServers": {n: scrub_mcp_server(c) for n, c in (data.get("mcpServers") or {}).items()}}
    projects = {}
    for cwd, proj in (data.get("projects") or {}).items():
        servers = proj.get("mcpServers") if isinstance(proj, dict) else None
        if servers:
            projects[cwd] = {"mcpServers": {n: scrub_mcp_server(c) for n, c in servers.items()}}
    out["projects"] = projects
    return out


def is_bridge_stub(path):
    try:
        with open(path, encoding="utf-8") as f:
            lines = [ln for ln in f.read().splitlines() if ln.strip()]
    except OSError:
        return False
    if len(lines) != 1:
        return False
    try:
        return json.loads(lines[0]).get("type") == "bridge-session"
    except json.JSONDecodeError:
        return False


def select_transcripts(claude_dir, live_ids, recent, exclude_id=None):
    projects = claude_dir / "projects"
    all_main = [p for p in projects.glob("*/*.jsonl") if p.stem != exclude_id]
    by_id = {p.stem: p for p in all_main}
    chosen = {}
    for sid in live_ids:
        if sid in by_id:
            chosen[sid] = by_id[sid]
    by_mtime = sorted(all_main, key=lambda p: p.stat().st_mtime, reverse=True)
    for p in by_mtime:
        if len([s for s in chosen if s not in live_ids]) >= recent:
            break
        chosen.setdefault(p.stem, p)
    stubs = sorted((p for p in all_main if is_bridge_stub(p)), key=lambda p: p.stat().st_size)
    if stubs:
        chosen.setdefault(stubs[0].stem, stubs[0])
    with_subagents = [p for p in by_mtime if (p.parent / p.stem / "subagents").is_dir()]
    if with_subagents:
        chosen.setdefault(with_subagents[0].stem, with_subagents[0])
    return chosen


def copy_session(src, home_out, raw_out):
    rel = src.relative_to(HOME)
    scrub_jsonl(src, home_out / rel)
    if raw_out:
        (raw_out / rel).parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, raw_out / rel)
    sub = src.parent / src.stem / "subagents"
    count = 0
    if sub.is_dir():
        for f in sorted(sub.iterdir()):
            frel = f.relative_to(HOME)
            if f.suffix == ".jsonl":
                scrub_jsonl(f, home_out / frel)
            elif f.name.endswith(".meta.json"):
                scrub_json_file(f, home_out / frel)
            else:
                continue
            count += 1
            if raw_out:
                (raw_out / frel).parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(f, raw_out / frel)
    return count


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--name", default="sample", help="fixture name under testing/helpers/fixtures (default: sample)")
    ap.add_argument("--claude-dir", default=str(HOME / ".claude"), help="source directory (default: ~/.claude)")
    ap.add_argument("--recent", type=int, default=3, help="extra most-recently-touched transcripts to include (default: 3)")
    ap.add_argument("--out", default=str(REPO / "testing" / "helpers" / "fixtures"), help="fixtures root (default: testing/helpers/fixtures)")
    ap.add_argument("--no-raw", action="store_true", help="skip the unscrubbed raw/ copy")
    ap.add_argument("--include-self", action="store_true",
                    help="keep the Claude Code session running this script (excluded by default via $CLAUDE_CODE_SESSION_ID)")
    args = ap.parse_args(argv)
    self_id = None if args.include_self else os.environ.get("CLAUDE_CODE_SESSION_ID")

    claude_dir = Path(args.claude_dir).expanduser()
    if not claude_dir.is_dir():
        print(f"{claude_dir} does not exist; nothing to generate", file=sys.stderr)
        return 0
    fixture = Path(args.out) / args.name
    home_out = fixture / "home"
    raw_out = None if args.no_raw else fixture / "raw"
    if fixture.exists():
        shutil.rmtree(fixture)
    home_out.mkdir(parents=True)

    live = {}
    for f in sorted((claude_dir / "sessions").glob("*.json")):
        with open(f, encoding="utf-8") as fh:
            rec = json.load(fh)
        if rec.get("sessionId") == self_id:
            continue
        live[rec.get("sessionId")] = f
        scrub_json_file(f, home_out / f.relative_to(HOME), scrub_session_record)
        if raw_out:
            (raw_out / f.relative_to(HOME)).parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(f, raw_out / f.relative_to(HOME))

    chosen = select_transcripts(claude_dir, set(live), args.recent, self_id)
    sessions = []
    for sid, path in sorted(chosen.items()):
        n = copy_session(path, home_out, raw_out)
        sessions.append({
            "sessionId": sid,
            "project": path.parent.name,
            "live": sid in live,
            "bridgeStub": is_bridge_stub(path),
            "subagentFiles": n,
            "bytes": path.stat().st_size,
        })

    stats = claude_dir / "stats-cache.json"
    if stats.exists():
        scrub_json_file(stats, home_out / ".claude" / "stats-cache.json", lambda d: d)
        if raw_out:
            shutil.copy2(stats, raw_out / ".claude" / "stats-cache.json")

    claude_json = HOME / ".claude.json"
    if claude_json.exists():
        scrub_json_file(claude_json, home_out / ".claude.json", scrub_claude_json)

    manifest = {
        "args": {"recent": args.recent, "claudeDir": str(claude_dir)},
        "liveSessions": len(live),
        "sessions": sessions,
    }
    with open(fixture / "manifest.json", "w", encoding="utf-8") as f:
        json.dump(manifest, f, indent=2)
        f.write("\n")

    for leaked in home_out.rglob("*.key"):
        print(f"refusing to keep key file in fixture: {leaked}", file=sys.stderr)
        return 1
    print(f"wrote {fixture}: {len(live)} live records, {len(sessions)} transcripts")
    return 0


if __name__ == "__main__":
    sys.exit(main())

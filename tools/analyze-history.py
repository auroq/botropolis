#!/usr/bin/env python3
"""Summarise Claude Code usage from the transcripts under ~/.claude/projects.

Prints session, timing, token, tool, MCP, and per-project statistics, and writes a
sessions.json next to the output for further analysis. Read-only; stdlib only.
"""
import argparse, json, os, glob, sys, collections, datetime as dt, statistics
HOME=os.path.expanduser('~')
ap=argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
ap.add_argument('--root', default=f'{HOME}/.claude/projects', help='projects directory to scan (default: ~/.claude/projects)')
ap.add_argument('--out', default=None, help='where to write sessions.json (default: alongside this script)')
args=ap.parse_args()
ROOT=args.root
NOW=dt.datetime.now(dt.timezone.utc)
IDLE_GAP=30*60  # seconds; a gap longer than this splits "active time"

def ts(s):
    try: return dt.datetime.fromisoformat(s.replace('Z','+00:00'))
    except: return None

sessions=[]
for f in glob.glob(f'{ROOT}/*/*.jsonl'):
    sid=os.path.basename(f)[:-6]
    S=dict(id=sid, file=f, size=os.path.getsize(f), cwd='', branch='', title='', first=None, last=None,
           prompts=0, assistant=0, tools=collections.Counter(), models=collections.Counter(), effort=collections.Counter(),
           tok=collections.Counter(), sidechain=0, mcp=collections.Counter(), skills=collections.Counter(),
           compactions=0, prs=set(), versions=set(), entrypoint='', active=0.0, ends_with='', last_user_text='',
           context_now=0, errors=0, permission_modes=collections.Counter(), stamps=[], agents=set(), interrupts=0)
    prev=None
    last_kind=None
    for line in open(f, encoding='utf-8', errors='replace'):
        try: r=json.loads(line)
        except: continue
        t=r.get('type')
        if r.get('cwd') and not S['cwd']: S['cwd']=r['cwd']
        if r.get('gitBranch') and r['gitBranch']!='HEAD' and not S['branch']: S['branch']=r['gitBranch']
        if r.get('version'): S['versions'].add(r['version'])
        if r.get('entrypoint') and not S['entrypoint']: S['entrypoint']=r['entrypoint']
        if t=='custom-title': S['title']=r['customTitle']
        elif t=='ai-title' and not S['title']: S['title']=r['aiTitle']
        elif t=='summary' and not S['title']: S['title']=r.get('summary','')
        if t=='pr-link': S['prs'].add(r.get('prUrl'))
        if t=='agent-name': S['agents'].add(r.get('agentName'))
        if t=='system' and r.get('compactMetadata'): S['compactions']+=1
        if t=='system' and r.get('subtype')=='api_error': S['errors']+=1
        if r.get('permissionMode'): S['permission_modes'][r['permissionMode']]+=1
        T=ts(r.get('timestamp','')) if r.get('timestamp') else None
        if T:
            S['stamps'].append(T)
            if S['first'] is None or T<S['first']: S['first']=T
            if S['last'] is None or T>S['last']: S['last']=T
            if prev is not None:
                gap=(T-prev).total_seconds()
                if 0<gap<IDLE_GAP: S['active']+=gap
            prev=T
        if r.get('isSidechain'): S['sidechain']+=1
        if t=='user':
            m=r.get('message',{}); c=m.get('content')
            is_tool_result = isinstance(c,list) and any(isinstance(p,dict) and p.get('type')=='tool_result' for p in c)
            if not is_tool_result and not r.get('isMeta') and not r.get('isSidechain'):
                text = c if isinstance(c,str) else ' '.join(p.get('text','') for p in c if isinstance(p,dict) and p.get('type')=='text') if isinstance(c,list) else ''
                if text and not text.startswith('<') :
                    S['prompts']+=1; S['last_user_text']=text[:120]; last_kind='user'
            if r.get('interruptedMessageId'): S['interrupts']+=1
        if t=='assistant':
            m=r.get('message',{})
            if r.get('isApiErrorMessage'): S['errors']+=1
            if not r.get('isSidechain'): S['assistant']+=1
            if m.get('model'): S['models'][m['model']]+=1
            if r.get('effort'): S['effort'][r['effort']]+=1
            if r.get('attributionMcpServer'): S['mcp'][r['attributionMcpServer']]+=1
            if r.get('attributionSkill'): S['skills'][r['attributionSkill']]+=1
            u=m.get('usage') or {}
            S['tok']['in']+=u.get('input_tokens',0); S['tok']['out']+=u.get('output_tokens',0)
            S['tok']['cache_read']+=u.get('cache_read_input_tokens',0); S['tok']['cache_create']+=u.get('cache_creation_input_tokens',0)
            S['tok']['thinking']+=(u.get('output_tokens_details') or {}).get('thinking_tokens',0)
            if u and not r.get('isSidechain'):
                S['context_now']=u.get('input_tokens',0)+u.get('cache_read_input_tokens',0)+u.get('cache_creation_input_tokens',0)
            calling=False
            for p in (m.get('content') or []):
                if isinstance(p,dict) and p.get('type')=='tool_use':
                    calling=True; n=p.get('name',''); S['tools'][n]+=1
                    if n.startswith('mcp__'): S['mcp'][n.split('__')[1]]+=1
            if not r.get('isSidechain'): last_kind='assistant_tool' if calling else 'assistant_done'
    S['ends_with']=last_kind
    S['duration']=(S['last']-S['first']).total_seconds() if S['first'] and S['last'] else 0
    S['age_days']=(NOW-S['last']).total_seconds()/86400 if S['last'] else None
    S['project']=S['cwd'] or os.path.basename(os.path.dirname(f))
    sessions.append(S)

sessions=[s for s in sessions if s['first']]
sessions.sort(key=lambda s:s['first'])
def h(n): 
    n=float(n)
    for u in ['','K','M','B']:
        if abs(n)<1000: return f'{n:.1f}{u}'
        n/=1000
    return f'{n:.1f}T'
def hd(sec): 
    sec=int(sec); return f'{sec//3600}h{(sec%3600)//60:02d}m'

print(f'# {len(sessions)} sessions, {sessions[0]["first"].date()} -> {sessions[-1]["last"].date()}')
print()
print('## Age of last activity (is it "done"?)')
buckets=[(1,'<1d'),(3,'1-3d'),(7,'3-7d'),(30,'7-30d'),(10**9,'>30d')]
c=collections.Counter()
for s in sessions:
    for lim,name in buckets:
        if s['age_days']<lim: c[name]+=1; break
print(dict(c))
print()
print('## How sessions ended (last main-thread record)')
print(collections.Counter(s['ends_with'] for s in sessions))
print('last user text samples for sessions >3d old:')
lt=collections.Counter()
for s in sessions:
    if s['age_days']>3: lt[s['last_user_text'][:40].lower()]+=1
for k,v in lt.most_common(15): print(f'  {v:3d}  {k!r}')
print()
print('## Duration (wall) vs active time')
durs=[s['duration'] for s in sessions]; act=[s['active'] for s in sessions]
print('wall  median',hd(statistics.median(durs)),'p90',hd(sorted(durs)[int(len(durs)*.9)]),'max',hd(max(durs)))
print('active median',hd(statistics.median(act)),'p90',hd(sorted(act)[int(len(act)*.9)]),'total',hd(sum(act)))
multi_day=sum(1 for s in sessions if s['duration']>86400)
print(f'sessions spanning >1 day wall time: {multi_day} (resumed / left open)')
print()
print('## Hour of day (local) of activity, by assistant messages')
hod=collections.Counter(); dow=collections.Counter()
for s in sessions:
    for T in s['stamps']:
        L=T.astimezone(); hod[L.hour]+=1; dow[L.strftime('%a')]+=1
print(' '.join(f'{hh:02d}:{hod[hh]}' for hh in range(24)))
print(dict(dow))
print()
print('## Concurrency: max simultaneous active sessions (5-min buckets)')
buck=collections.defaultdict(set)
for s in sessions:
    for T in s['stamps']:
        buck[int(T.timestamp()//300)].add(s['id'])
conc=collections.Counter(len(v) for v in buck.values())
print(dict(sorted(conc.items())))
print()
print('## Per project')
P=collections.defaultdict(lambda: dict(n=0,active=0,out=0,cache=0,prs=set(),last=None))
for s in sessions:
    p=P[s['project']]; p['n']+=1; p['active']+=s['active']; p['out']+=s['tok']['out']; p['cache']+=s['tok']['cache_read']; p['prs']|=s['prs']
    p['last']=max(p['last'],s['last']) if p['last'] else s['last']
for k,p in sorted(P.items(), key=lambda kv:-kv[1]['active']):
    print(f"{p['n']:3d} sess  active {hd(p['active']):>8}  out {h(p['out']):>7}  cacheRead {h(p['cache']):>7}  PRs {len(p['prs']):2d}  last {(NOW-p['last']).days:3d}d ago  {k.replace(HOME,'~')}")
print()
print('## Tokens overall')
tot=collections.Counter()
for s in sessions: tot.update(s['tok'])
print({k:h(v) for k,v in tot.items()})
print('cache hit ratio', f"{tot['cache_read']/(tot['cache_read']+tot['cache_create']+tot['in']):.1%}")
print()
print('## Models / effort')
mc=collections.Counter(); ec=collections.Counter()
for s in sessions: mc.update(s['models']); ec.update(s['effort'])
print(mc.most_common()); print(ec.most_common())
print()
print('## Tools (top 25)')
tc=collections.Counter()
for s in sessions: tc.update(s['tools'])
for k,v in tc.most_common(25): print(f'  {v:6d} {k}')
print()
print('## MCP servers used (attribution or mcp__ tool calls)')
mcp=collections.Counter()
for s in sessions: mcp.update(s['mcp'])
print(mcp.most_common())
print('sessions using any MCP:', sum(1 for s in sessions if s['mcp']))
print()
print('## Skills')
sk=collections.Counter()
for s in sessions: sk.update(s['skills'])
print(sk.most_common(20))
print()
print('## Subagents / compaction / PRs / errors / interrupts')
print('sessions with sidechain msgs:', sum(1 for s in sessions if s['sidechain']), 'total sidechain msgs', sum(s['sidechain'] for s in sessions))
print('sessions with compactions:', sum(1 for s in sessions if s['compactions']), 'total', sum(s['compactions'] for s in sessions))
print('sessions with PR links:', sum(1 for s in sessions if s['prs']), 'distinct PRs', len(set().union(*[s['prs'] for s in sessions])))
print('sessions with api errors:', sum(1 for s in sessions if s['errors']))
print('user interrupts total:', sum(s['interrupts'] for s in sessions))
print('context_now distribution (last main-thread request):')
cn=sorted(s['context_now'] for s in sessions if s['context_now'])
print(' median',h(statistics.median(cn)),'p90',h(cn[int(len(cn)*.9)]),'max',h(max(cn)), '>150K:',sum(1 for c in cn if c>150000))
print('permission modes:', sum((s['permission_modes'] for s in sessions), collections.Counter()).most_common())
print('entrypoints:', collections.Counter(s['entrypoint'] for s in sessions))
print('versions seen:', len(set().union(*[s['versions'] for s in sessions])))
print()
print('## Top 15 sessions by active time')
for s in sorted(sessions,key=lambda s:-s['active'])[:15]:
    print(f"{hd(s['active']):>7} active / {hd(s['duration']):>8} wall  {s['prompts']:4d} prompts  out {h(s['tok']['out']):>6}  ctx {h(s['context_now']):>6}  {s['age_days']:5.1f}d  {os.path.basename(s['project'])[:22]:22s} {s['title'][:50]!r}")

json.dump([{k:(v if not isinstance(v,(set,dt.datetime,collections.Counter)) else (list(v) if isinstance(v,set) else v.isoformat() if isinstance(v,dt.datetime) else dict(v))) for k,v in s.items() if k!='stamps'} for s in sessions], open(args.out or os.path.join(os.path.dirname(os.path.abspath(__file__)),'sessions.json'),'w'), indent=1, default=str)

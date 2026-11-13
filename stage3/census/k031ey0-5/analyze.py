import json,collections,re,hashlib
from pathlib import Path
out=Path(__file__).resolve().parent
needle='which instantiates itself deeper at each step, so no walk over it ends'
def read(p): return json.loads(p.read_text())
def lines(p): return [json.loads(l) for l in p.read_text().splitlines()]
def stop(r): return (r['kind'],r.get('where',''),r.get('reason',''),r.get('diagnostics',[]))
def new(r): return needle in r.get('reason','')
def short(s): return s.replace('/workspace/scratch/k031ey0-5/adapted/src/compiler/','').replace('/workspace/scratch/k031ey0-5/adapted/','')
roots=read(out/'roots.json'); data={}; changes=[]
for side in ['parent','head']:
 data[side]=[]
 for i,root in enumerate(roots,1):
  p=out/f'{side}-root-{i:02}.jsonl'; rows=lines(p); assert len(rows)==2
  assert {k:v for k,v in rows[0].items() if k!='seconds'}=={k:v for k,v in rows[1].items() if k!='seconds'}
  data[side].append(dict(root=root,first=rows[0],second=rows[1],status=read(out/f'{side}-root-{i:02}-status.json')))
for p,h in zip(data['parent'],data['head']):
 if stop(p['first'])!=stop(h['first']): changes.append(dict(root=p['root'],parent=p['first'],head=h['first']))
summary=dict(changed=changes,ordinary={},latent={})
for side in ['parent','head']:
 d=data[side]; summary['ordinary'][side]=dict(first_seconds=sum(r['first']['seconds'] for r in d),second_seconds=sum(r['second']['seconds'] for r in d),child_wall_seconds=sum(r['status']['seconds'] for r in d),new_reason=[r['first'] for r in d if new(r['first'])])
 lat=lines(out/(side+'-latent.jsonl')); findings=[f for record in lat for f in record.get('findings',[])]; units=[u for r in lat for u in r.get('units',[])]; nf=[f for f in findings if new(f)]
 pre=[]
 for line in (out/(side+'-prepass-times.tsv')).read_text().splitlines():
  method,owner,ns=line.rsplit('\t',2)
  if method=='collectViewedElements': pre.append(dict(owner=owner,seconds=int(ns)/1e9))
 stderr=[l for l in (out/(side+'-latent.log')).read_text().splitlines() if l.startswith('prepass collectViewedElements error:')]
 summary['latent'][side]=dict(wall=read(out/(side+'-latent-status.json'))['seconds'],files=sum('file' in r for r in lat),units=len(units),findings=len(findings),new_findings=nf,new_units=sorted(set(f['unit'] for f in nf)),prepass_calls=len(pre),prepass_seconds=sum(r['seconds'] for r in pre),prepass_error_count=len(stderr),prepass_errors=dict(collections.Counter(stderr)))
(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
report=['# Paired census: k031ey0 -5','', 'Measurement only. Parent `f44b5c272a39bcbd7cd2fb278939ac1c1832ce2e`; head `e64bc123c51fc333133684f0abb3bee8e522acf3`. One adapted TypeScript v6.0.3 tree (upstream `050880ce59e30b356b686bd3144efe24f875ebc8`), prepared using parent stage3; stage3 is identical at head. Exact commands: [COMMANDS.md](COMMANDS.md); scratch-only diagnostic change: [seam.diff](seam.diff).','',f"## First stops: {len(changes)} / 79 changed",'']
if not changes: report+=['No ordinary root changed its first (kind, location, reason, diagnostics).']
for c in changes:
 report += [f"- `{c['root']}`: {c['parent']['kind']} `{short(c['parent'].get('where',''))}` — {c['parent'].get('reason','')} → {c['head']['kind']} `{short(c['head'].get('where',''))}` — {c['head'].get('reason','')}"]
report+=['','## New nesting reason','']
for side in ['parent','head']:
 o=summary['ordinary'][side]; l=summary['latent'][side]
 report += [f"{side}: {len(o['new_reason'])} ordinary roots; {len(l['new_units'])} latent units ({len(l['new_findings'])} findings) stop on the new reason. Latent stream: {l['files']} files, {l['units']} units, {l['findings']} findings."]
 for f in o['new_reason']+l['new_findings']: report += ['```',json.dumps(f,ensure_ascii=False),'```']
 report += [f"Prepass stderr: {l['prepass_error_count']} lines, {len(l['prepass_errors'])} distinct."]
 for line,n in l['prepass_errors'].items(): report += ['```',line,'```',f'Occurrences: {n}.']
report+=['','The seam only exposes the discarded prepass error on stderr and continues the existing meter pass; a prepass diagnostic is reported separately from recorded unit stops.','','## Wall time','', '| Side | Ordinary first observations (s) | Duplicate observations (s) | All ordinary children wall (s) | Entry latent whole wall (s) | collectViewedElements total (s) | Prepass calls |','|---|---:|---:|---:|---:|---:|---:|']
for side in ['parent','head']:
 o=summary['ordinary'][side]; l=summary['latent'][side]
 report += [f"| {side} | {o['first_seconds']:.6f} | {o['second_seconds']:.6f} | {o['child_wall_seconds']:.6f} | {l['wall']:.6f} | {l['prepass_seconds']:.6f} | {l['prepass_calls']} |"]
report += ['', 'Ordinary seconds in the next table are the tool’s first-observation load+Lower elapsed time. Child wall includes both observations and process overhead. Totals are sums; preparation and builds are excluded. A single paired run gives an observed difference, not a statistically established performance change. All ordinary duplicate observations agree in every non-timing field.','','## Worker and load','']
for side in ['parent','head']:
 statuses=[r['status'] for r in data[side]]; ls=read(out/(side+'-latent-status.json')); before=statuses[0]['before']; after=statuses[-1]['after']; loads=[s[k]['load'][0] for s in statuses for k in ['before','after']]
 report += [f"{side}: box `{before['hostname']}`, {before['cpus']} available CPUs, cpu.max `{before['cpu_max']}`. Ordinary start load {before['load']}, end {after['load']}; sampled 1-minute load range {min(loads):.2f}–{max(loads):.2f}. Entry latent start load {ls['before']['load']}, end {ls['after']['load']}. All three load values are 1/5/15-minute averages. Per-child timestamps and load snapshots are in status files."]
report+=['','Roots run sequentially, parent then head for each root, in four groups of at most 20. Entry latent runs parent then head after the ordinary census. No census overlaps another. Child timeout is 1780 seconds. Every wrapper reports the child exit code.','','## Per-root first stops and seconds','', '| Root | Parent stop | Head stop | Parent seconds | Head seconds | Parent child wall | Head child wall |','|---|---|---|---:|---:|---:|---:|']
for p,h in zip(data['parent'],data['head']):
 def cell(r): return (r['kind']+' '+short(r.get('where',''))+' '+r.get('reason','')).strip().replace('|','\\|')
 report += [f"| `{p['root'].removeprefix('src/compiler/')}` | {cell(p['first'])} | {cell(h['first'])} | {p['first']['seconds']:.6f} | {h['first']['seconds']:.6f} | {p['status']['seconds']:.6f} | {h['status']['seconds']:.6f} |"]
(out/'REPORT.md').write_text('\n'.join(report)+'\n')
print(json.dumps(summary,indent=2))

import json,collections,os
from pathlib import Path
out=Path(__file__).resolve().parent
needle='a view of a type that reaches '
def read(p):return json.loads(p.read_text())
def rows(p):return [json.loads(l) for l in p.read_text().splitlines()]
def clean(s):return s.replace('/workspace/scratch/k031ey0-5/adapted/src/compiler/','')
data={};summary={'ordinary':{},'latent':'not run','scope':'Incremental ordinary census evidence; latent census not yet started.'}
for side in ['parent','head']:
 d=[]
 for status in sorted(out.glob(side+'-root-*-status.json')):
  n=int(status.name.split('-')[2]); r=rows(out/f'{side}-root-{n:02}.jsonl'); assert len(r)==2
  assert {k:v for k,v in r[0].items() if k!='seconds'}=={k:v for k,v in r[1].items() if k!='seconds'}
  d.append(dict(index=n,first=r[0],second=r[1],status=read(status)))
 data[side]=d
 summary['ordinary'][side]=dict(count=len(d),first_seconds=sum(x['first']['seconds'] for x in d),duplicate_seconds=sum(x['second']['seconds'] for x in d),child_wall_seconds=sum(x['status']['seconds'] for x in d),new_reason=[x['first'] for x in d if x['first'].get('reason','').startswith(needle)],unrun_indices=sorted(set(range(1,80))-{x['index'] for x in d}))
pair={s:{x['index']:x for x in d} for s,d in data.items()};changes=[]
for i in sorted(pair['parent'].keys() & pair['head'].keys()):
 a,b=[pair[s][i]['first'] for s in ['parent','head']]
 if any(a.get(k)!=b.get(k) for k in ['kind','where','reason','diagnostics']):changes.append(dict(index=i,parent=a,head=b))
summary['changes']=changes
(out/'partial-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
roots=read(out/'roots.json')
report=['# k031ey0 -5 paired census: partial checkpoint','','Evidence only. The first 60 pairs were pushed at the requested checkpoint. Continuation groups are pushed after each group of at most 10 roots.','','Parent: `f44b5c272a39bcbd7cd2fb278939ac1c1832ce2e`. Head: `e64bc123c51fc333133684f0abb3bee8e522acf3`. TypeScript v6.0.3 upstream: `050880ce59e30b356b686bd3144efe24f875ebc8`. One adapted tree prepared from parent; stage3 is identical at head. Head production sources are unmodified. The diagnostic seam is scratch-only, saved in [seam.diff](seam.diff), and was compiled but has not been run.','','## Measured so far','', '| Side | Completed roots | First observations (s) | Duplicate observations (s) | All measured child wall (s) |','|---|---:|---:|---:|---:|']
for side,s in summary['ordinary'].items():report += [f"| {side} | {s['count']}/79 | {s['first_seconds']:.6f} | {s['duplicate_seconds']:.6f} | {s['child_wall_seconds']:.6f} |"]
report+=['',f"Changed first stops among completed pairs: **{len(changes)}**. First stop compares kind, location, exact reason and diagnostics. All duplicate observations agree in every non-timing field.",'']
for c in changes:report+=['```json',json.dumps(c,indent=2),'```']
for side,s in summary['ordinary'].items():
 report += [f"{side}: **{len(s['new_reason'])}** roots stop first on `a view of a type that reaches ...`." ]
 for r in s['new_reason']:report+=['```json',json.dumps(r,indent=2),'```']
report+=['','Entry latent census: **not run on either side**. No latent unit counts, whole-pass wall times, collectViewedElements prepass timings, or stderr prepass diagnostic have been measured. Absence of a diagnostic here is not a negative latent result.','','## Timing and load','', 'The first-observation seconds are the census tool’s load+Lower elapsed time. Child wall includes both observations and process overhead; totals exclude builds and adaptation. This is a single alternating measurement, with no statistical performance claim.']
for side,d in data.items():
 if not d:continue
 before=d[0]['status']['before'];after=d[-1]['status']['after'];loads=[x['status'][k]['load'][0] for x in d for k in ['before','after']]
 report += ['',f"{side}: box `{before['hostname']}`; {before['cpus']} available CPUs; cpu.max `{before['cpu_max']}` (4 CPU quota). Start load {before['load']}; end load {after['load']}; sampled 1-minute range {min(loads):.2f}–{max(loads):.2f}. Load triples are 1/5/15-minute averages. Every child’s before/after loads and UTC timestamps are in its status JSON."]
report += ['', 'Roots run sequentially, parent then head, in groups of at most 20; no measured children overlap. Each child has a 1780-second bound. All completed children exited 0. The initial checkpoint completed roots 1–60 in groups of 20. Continuation uses groups 61–70 and 71–79, pushed separately before proceeding. Both latent passes are not yet started.','','## Unrun','']
for side,s in summary['ordinary'].items():
 report += [f"{side}: {len(s['unrun_indices'])} ordinary roots:"]
 for i in s['unrun_indices']:report += [f"- {i}: `{roots[i-1]}`"]
report+=['','Both parent and head entry latent censuses remain unrun, including LATENT_TIMES, latentPrepassTimer output and stderr seam observation.','','## Per-root stops and seconds','', '| Root | Parent first stop | Head first stop | Parent seconds | Head seconds | Parent child wall | Head child wall |','|---|---|---|---:|---:|---:|---:|']
for i in sorted(pair['parent'].keys() | pair['head'].keys()):
 a=pair['parent'].get(i);b=pair['head'].get(i)
 def stop(x):
  if not x:return 'unrun'
  r=x['first'];return (r['kind']+' '+clean(r.get('where',''))+' '+r.get('reason','')).strip().replace('|','\\|')
 def t(x,key):return f"{x['first']['seconds'] if key=='first' else x['status']['seconds']:.6f}" if x else '-'
 report += [f"| {i}: `{roots[i-1].removeprefix('src/compiler/')}` | {stop(a)} | {stop(b)} | {t(a,'first')} | {t(b,'first')} | {t(a,'wall')} | {t(b,'wall')} |"]
report+=['','Raw JSONL observations, stderr logs and command/exit/load status JSON are adjacent to this report. Setup logs are under `preparation/`. See [COMMANDS.md](COMMANDS.md) and executable [measure.py](measure.py). `analyze.py` is saved for the later full run; this checkpoint report is produced by `partial-report.py`.']
(out/'REPORT.md').write_text('\n'.join(report)+'\n')
print(json.dumps(summary,indent=2))

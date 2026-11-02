#!/usr/bin/env python3
import collections, gzip, importlib.util, json, statistics
from pathlib import Path
out=Path(__file__).resolve().parent
root=Path('/workspace/scratch/tree')
spec=importlib.util.spec_from_file_location('meter_report','/workspace/scratch/compiler-A/stage3/meter/report.py')
m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
def read(p):
 text=p.read_text() if p.exists() else gzip.open(str(p)+'.gz','rt').read()
 return [json.loads(x) for x in text.splitlines()]
def short(x):return x.replace(str(root)+'/', '')
def dump(p,x):(out/p).write_text(json.dumps(x,indent=2)+'\n')
expected={str(p.resolve()) for p in (root/'src/compiler').rglob('*') if p.is_file() and p.suffix in ('.ts','.a')}
records={s:read(out/f'{s}.jsonl') for s in ('A','B')}
assert records['A'][0]==records['B'][0], 'identical checker diagnostics'
summaries={s:m.lowering_summary(r,expected,m.MEASUREMENT) for s,r in records.items()}
dump('validation.json',summaries)
units={};sites={}
for side,rs in records.items():
 us={};ss=set()
 for r in rs[1:]:
  findings=r.get('findings') or []
  for f in findings:ss.add(tuple(f[k] for k in ('kind','where','reason','text')))
  for u in r['units']:
   key=u['where']; fs=[f for f in findings if f['unit']==key and f['phase']=='lowering']
   stop=next((f for f in fs if f['kind'] not in ('Boundary',)),None)
   outcome='lowers' if u['status']=='attempted' and not stop else ('stopped' if stop else u['status'])
   us[key]=dict(u,outcome=outcome,first_stop=stop)
 units[side]=us;sites[side]=ss
assert units['A'].keys()==units['B'].keys()
def result(u):
 f=u['first_stop']
 return {'outcome':u['outcome'],'status':u['status'],'reason':f['reason'] if f else None,'where':short(f['where']) if f else None,'kind':f['kind'] if f else None,'checker_diagnostics':u.get('checker_diagnostics',[])}
source_names=json.loads((out/'source-unit-names.json').read_text())
def named(k):return units['A'][k].get('name') or source_names.get(k) or short(k)
def sig(u):return (u['outcome'],tuple((u['first_stop'] or {}).get(k) for k in ('kind','reason','where')))
moves=[{'unit':short(k),'name':named(k),'A':result(units['A'][k]),'B':result(units['B'][k])} for k in units['A'] if sig(units['A'][k])!=sig(units['B'][k])]
regressions=[r for r in moves if r['A']['outcome']=='lowers' and r['B']['outcome']!='lowers']
dump('unit-moves.json',moves);dump('lowering-regressions.json',regressions)
lowering_sites_A={tuple(f[k] for k in ('kind','where','reason','text')) for r in records['A'][1:] for f in (r.get('findings') or []) if f['phase']=='lowering'}
removed=[dict(zip(('kind','where','reason','text'),s)) for s in sorted((sites['A']-sites['B']) & lowering_sites_A)]
dump('removed-lowering-sites.json',removed)
# Every changing reason, including losses, gets up to five changed-site locations on each side.
reasonrows=[]
for kind in ('NotYet','Refused'):
 reasons=sorted({s[2] for ss in sites.values() for s in ss if s[0]==kind})
 for reason in reasons:
  a={s for s in sites['A'] if s[0]==kind and s[2]==reason};b={s for s in sites['B'] if s[0]==kind and s[2]==reason}
  if len(a)!=len(b):reasonrows.append({'kind':kind,'reason':reason,'A':len(a),'B':len(b),'delta':len(b)-len(a),'removed_examples':[short(s[1]) for s in sorted(a-b)][:5],'added_examples':[short(s[1]) for s in sorted(b-a)][:5]})
dump('reason-changes.json',reasonrows)
# Match AST top-level statements to the census's own unit start locations.
calls=json.loads((out/'direct-calls.json').read_text())
wanted={c['unit_where'] for c in calls}
wanted.update(k for k,u in units['A'].items() if 'nullParenthesizerRules'==named(k) or 'factory/parenthesizerRules.ts:675:' in k)
assert wanted <= units['A'].keys(), sorted(wanted-units['A'].keys())
selected=[{'unit':short(k),'name':named(k),'calls':sum(c['unit_where']==k for c in calls),'A':result(units['A'][k]),'B':result(units['B'][k])} for k in sorted(wanted)]
dump('selected-unit-outcomes.json',selected)
tr={s:{r['where']:r for r in read(out/f'{s}-times.jsonl')} for s in ('A','B')}
assert tr['A'].keys()==tr['B'].keys()
for side in ('A','B'):
 assert len(tr[side])==len(read(out/f'{side}-times.jsonl'))
 assert tr[side].keys()=={k for k,u in units[side].items() if u['status']=='attempted'}
times=[]
for k,a in tr['A'].items():
 b=tr['B'][k]; av=a['wall_seconds'];bv=b['wall_seconds']
 times.append({'unit':short(k),'name':named(k),'A_seconds':av,'B_seconds':bv,'delta_seconds':bv-av,'ratio_B_A':bv/av if av else None,'A_cpu_seconds':a['cpu_seconds'],'B_cpu_seconds':b['cpu_seconds']})
times.sort(key=lambda r:(-r['delta_seconds'],r['unit']))
dump('unit-times.json',times)
cols=['unit','name','A_seconds','B_seconds','delta_seconds','ratio_B_A','A_cpu_seconds','B_cpu_seconds']
(out/'unit-times.tsv').write_text('\t'.join(cols)+'\n'+'\n'.join('\t'.join(str(r[c]) for c in cols) for r in times)+'\n')
noise=[r for r in times if units['A'][str(root)+'/'+r['unit']]['kind'] in ('KindImportDeclaration','KindExportDeclaration','KindEmptyStatement')]
assert noise
absdiff=sorted(abs(r['delta_seconds']) for r in noise)
noise_stats={'definition':'import/export declarations and empty statements: no receiver read is lowered; includes registration, scheduling and GC noise','units':len(noise),'median_absolute_seconds':statistics.median(absdiff),'p90_absolute_seconds':absdiff[int(.9*(len(absdiff)-1))],'max_absolute_seconds':max(absdiff),'median_ratio':statistics.median(r['ratio_B_A'] for r in noise),'A_total_seconds':sum(r['A_seconds'] for r in noise),'B_total_seconds':sum(r['B_seconds'] for r in noise)}
dump('noise-floor.json',noise_stats)
lines=['| Unit | A seconds | B seconds | B − A seconds | B / A |','|---|---:|---:|---:|---:|']
for r in times[:20]:lines.append(f"| {r['name']} ({r['unit']}) | {r['A_seconds']:.6f} | {r['B_seconds']:.6f} | {r['delta_seconds']:+.6f} | {r['ratio_B_A']:.3f} |")
(out/'timing-top20.md').write_text('\n'.join(lines)+'\n')
headline={'files':len(expected),'totals':{s:summaries[s]['totals'] for s in ('A','B')},'unit_outcomes':{s:dict(collections.Counter(u['outcome'] for u in us.values())) for s,us in units.items()},'moves':len(moves),'regressions':len(regressions),'direct_calls':len(calls),'direct_call_units':len({c['unit_where'] for c in calls}),'timed_units':len(times),'unit_seconds':{s:sum(r[s+'_seconds'] for r in times) for s in ('A','B')},'noise':noise_stats}
dump('headline.json',headline)
print(json.dumps(headline,indent=2))

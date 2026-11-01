#!/usr/bin/env python3
import collections,csv,importlib.util,json,pathlib,statistics,sys
repo=pathlib.Path('/workspace/adamic'); raw=pathlib.Path('/workspace/scratch/changeb-final-meter'); out=pathlib.Path(__file__).resolve().parent
root=pathlib.Path('/workspace/scratch/tree'); source=root/'src/compiler'
def dump(name,x): (out/name).write_text(json.dumps(x,indent=2)+'\n')
def norm(x): return x.replace(str(root)+'/', '')
def records(side): return [json.loads(s) for s in (raw/(side+'.jsonl')).read_text().splitlines()]
a,b=records('A'),records('B')
spec=importlib.util.spec_from_file_location('meter',repo/'stage3/meter/report.py'); meter=importlib.util.module_from_spec(spec); spec.loader.exec_module(meter)
expected={str(p.resolve()) for p in source.rglob('*') if p.is_file() and p.suffix in ('.ts','.a')}
assert len(expected)==79
for side,rows in [('A',a),('B',b)]:
 summary=meter.lowering_summary(rows,expected,meter.MEASUREMENT)
 dump(side+'-validation.json',summary)
 print(side,summary['source_files'],summary['totals'])
def sites(rows): return {tuple(norm(f[k]) for k in ('kind','where','reason','text')) for r in rows[1:] for f in r['findings']}
asites,bsites=sites(a),sites(b)
reasons=[]
for kind,reason in sorted({(s[0],s[2]) for s in asites|bsites if s[0] in ('NotYet','Refused')}):
 left={s for s in asites if s[0]==kind and s[2]==reason}; right={s for s in bsites if s[0]==kind and s[2]==reason}
 if len(left)!=len(right): reasons.append({'kind':kind,'reason':reason,'A':len(left),'B':len(right),'delta':len(right)-len(left),'A_examples':sorted({s[1] for s in left})[:5],'B_examples':sorted({s[1] for s in right})[:5],'named_callout':any(s in reason.lower() for s in ('view','branded supertype','predicate'))})
dump('changed-reasons.json',reasons)
lines=['| Kind / reason | A | B | Delta | A examples (up to 5) | B examples (up to 5) |','|---|---:|---:|---:|---|---|']
for r in reasons:
 text=r['reason'].replace('|','&#124;').replace('\n',' ')
 lines.append(f"| {'**CALL OUT** ' if r['named_callout'] else ''}{r['kind']}: {text} | {r['A']} | {r['B']} | {r['delta']:+} | {'; '.join(r['A_examples'])} | {'; '.join(r['B_examples'])} |")
(out/'changed-reasons.md').write_text('\n'.join(lines)+'\n')
def units(rows):
 result={}
 for r in rows[1:]:
  for u in r['units']:
   key=norm(u['where'])
   assert key not in result,key
   fs=[f for f in r['findings'] if f['unit']==u['where'] and f['phase']=='lowering']
   result[key]={'unit':dict(u,where=key),'findings':[{k:norm(v) if isinstance(v,str) else v for k,v in f.items()} for f in fs]}
 return result
ua,ub=units(a),units(b)
def lowers(u): return u['unit']['status']=='attempted' and not u['findings']
regressions=[{'where':key,'A':x,'B':ub.get(key)} for key,x in sorted(ua.items()) if lowers(x) and (key not in ub or not lowers(ub[key]))]
dump('lowering-regressions.json',regressions)
dump('unit-ledger-delta.json',[{'where':k,'A':ua.get(k),'B':ub.get(k)} for k in sorted(ua.keys()|ub.keys()) if ua.get(k)!=ub.get(k)])
# Every lowering-phase site disappearing from B is retained, as well as successful unit regressions.
fa={tuple(norm(f[k]) for k in ('kind','where','reason','text')) for r in a[1:] for f in r['findings'] if f['phase']=='lowering'}
fb={tuple(norm(f[k]) for k in ('kind','where','reason','text')) for r in b[1:] for f in r['findings'] if f['phase']=='lowering'}
dump('A-lowering-findings-absent-in-B.json',[dict(zip(('kind','where','reason','text'),s)) for s in sorted(fa-fb)])
unknown=[]
for key,x in sorted(ub.items()):
 for f in x['findings']:
  if 'which this body is lowered without knowing' in f['reason']:
   path,line,col=f['where'].rsplit(':',2)
   text=(root/path).read_text().splitlines()[int(line)-1]
   unknown.append({'unit_where':key,'name':x['unit'].get('name') or x['unit']['kind'],'stop_where':f['where'],'reason':f['reason'],'predicate_call_line':text.strip()})
dump('unknown-instance-units.json',unknown)
def times(side):
 result={}
 for line in (raw/(side+'-times.jsonl')).read_text().splitlines():
  x=json.loads(line); k=norm(x['where']); assert k not in result,k
  result[k]=x
 return result
ta,tb=times('A'),times('B')
assert ta.keys()==tb.keys(),(ta.keys()-tb.keys(),tb.keys()-ta.keys())
rows=[]
for k in ta:
 x,y=ta[k],tb[k]
 rows.append({'where':k,'name':ub[k]['unit'].get('name') or ub[k]['unit']['kind'],'kind':ub[k]['unit']['kind'],'A_seconds':x['wall_seconds'],'B_seconds':y['wall_seconds'],'B_minus_A_seconds':y['wall_seconds']-x['wall_seconds'],'ratio':y['wall_seconds']/x['wall_seconds'] if x['wall_seconds'] else None,'A_cpu_seconds':x['cpu_seconds'],'B_cpu_seconds':y['cpu_seconds']})
rows.sort(key=lambda x:(-x['B_minus_A_seconds'],x['where']))
with (out/'unit-times.csv').open('w') as file:
 writer=csv.DictWriter(file,fieldnames=list(rows[0]),lineterminator='\n'); writer.writeheader(); writer.writerows(rows)
controls=[x for x in rows if x['kind'] in ('KindImportDeclaration','KindExportDeclaration','KindInterfaceDeclaration','KindTypeAliasDeclaration')]
def quantile(xs,p):
 xs=sorted(xs); return xs[int((len(xs)-1)*p)] if xs else None
noise={'control_kinds':sorted({x['kind'] for x in controls}),'units':len(controls),'median_absolute_delta_seconds':statistics.median(abs(x['B_minus_A_seconds']) for x in controls),'p95_absolute_delta_seconds':quantile([abs(x['B_minus_A_seconds']) for x in controls],.95),'median_ratio':statistics.median(x['ratio'] for x in controls),'p05_ratio':quantile([x['ratio'] for x in controls],.05),'p95_ratio':quantile([x['ratio'] for x in controls],.95)}
summary={'units':len(rows),'A_seconds':sum(x['A_seconds'] for x in rows),'B_seconds':sum(x['B_seconds'] for x in rows),'A_cpu_seconds':sum(x['A_cpu_seconds'] for x in rows),'B_cpu_seconds':sum(x['B_cpu_seconds'] for x in rows),'noise':noise,'top20':rows[:20],'lowering_A':sum(lowers(x) for x in ua.values()),'lowering_B':sum(lowers(x) for x in ub.values()),'regressions':len(regressions),'unknown_instance_units':len(unknown)}
dump('timing-summary.json',summary)
lines=[f"Attempted units: {len(rows)}. A total {summary['A_seconds']:.6f}s; B total {summary['B_seconds']:.6f}s; difference {summary['B_seconds']-summary['A_seconds']:+.6f}s; ratio {summary['B_seconds']/summary['A_seconds']:.4f}.",f"CPU totals: A {summary['A_cpu_seconds']:.6f}s; B {summary['B_cpu_seconds']:.6f}s.",'','| Unit | A seconds | B seconds | B−A seconds | B/A |','|---|---:|---:|---:|---:|']
for r in rows[:20]: lines.append(f"| {r['name']} ({r['where']}) | {r['A_seconds']:.6f} | {r['B_seconds']:.6f} | {r['B_minus_A_seconds']:+.6f} | {r['ratio']:.4f} |")
lines+=['',f"Noise controls: {noise['units']} import/export/interface/type-alias units (no function body/predicate call lowered). Median absolute paired delta {noise['median_absolute_delta_seconds']:.6f}s, p95 {noise['p95_absolute_delta_seconds']:.6f}s; median B/A {noise['median_ratio']:.4f}, p05–p95 {noise['p05_ratio']:.4f}–{noise['p95_ratio']:.4f}. Single concurrent sample; scheduling, cache and GC effects limit causal timing attribution."]
(out/'unit-times.md').write_text('\n'.join(lines)+'\n')
print(json.dumps(summary,indent=2))

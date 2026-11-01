"""Reproduce coverage, paired first stops, site deltas and unit timing tables."""
import collections, csv, importlib.util, json, statistics, sys
from pathlib import Path
scratch=Path('/workspace/scratch')
out=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('meter', scratch/'compiler-A/stage3/meter/report.py')
meter=importlib.util.module_from_spec(spec); spec.loader.exec_module(meter)
def dump(name,value): (out/name).write_text(json.dumps(value,indent=2)+'\n')
def read(name): return [json.loads(x) for x in (scratch/name).read_text().splitlines()]
def norm(x): return x.replace('/workspace/scratch/tree/','')
def site(f): return tuple(norm(f[k]) for k in ('kind','where','reason','text'))
records={s:read(s+'.jsonl') for s in 'AB'}
expected={str(p.resolve()) for p in (scratch/'tree/src/compiler').rglob('*') if p.is_file() and p.suffix in ('.ts','.a')}
assert len(expected)==79
summaries={s:meter.lowering_summary(records[s],expected,meter.MEASUREMENT) for s in 'AB'}
dump('validation.json',summaries)
units={}; findings={}; sites={}
for s in 'AB':
 units[s]={norm(u['where']):dict(u,file=norm(r['file'])) for r in records[s][1:] for u in r['units']}
 by_line=collections.defaultdict(list)
 for w,u in units[s].items(): by_line[w.rsplit(':',1)[0]].append(u)
 for r in records[s][1:]:
  for d in r.get('declarations',[]):
   for u in by_line[norm(d['where']).rsplit(':',1)[0]]:
    if not u.get('name') and d['kind']=='KindVariableDeclaration': u['name']=d['name'].strip()
 findings[s]=collections.defaultdict(list)
 sites[s]=set()
 for r in records[s][1:]:
  for f in r['findings']:
   sites[s].add(site(f))
   if f['phase']=='lowering' and f['kind']!='Boundary': findings[s][norm(f['unit'])].append(f)
def state(s,w):
 u=units[s].get(w)
 if u is None:return {'status':'absent'}
 fs=findings[s][w]
 if fs:return {'status':'stopped','unit_status':u['status'],'kind':fs[0]['kind'],'reason':norm(fs[0]['reason']),'where':norm(fs[0]['where']),'all_stops':[{k:norm(f[k]) for k in ('kind','reason','where')} for f in fs]}
 return {'status':'lowers' if u['status']=='attempted' else u['status']}
regress=[]
for w,u in units['A'].items():
 if state('A',w)['status']=='lowers' and state('B',w)['status']!='lowers':regress.append({'unit':w,'name':u.get('name',''), 'A':state('A',w),'B':state('B',w)})
dump('lowering-regressions.json',regress)
# A lowering-phase diagnostic sites no longer witnessed in B, and sites whose owning unit stopped differently.
lowering_a={site(f) for r in records['A'][1:] for f in r['findings'] if f['phase']=='lowering' and f['kind']!='Boundary'}
removed=sorted(lowering_a-sites['B'])
dump('removed-lowering-sites.json',[dict(zip(('kind','where','reason','text'),x)) for x in removed])
names=['isOptionalJSDocPropertyLikeTag','isStatic','isClassOrTypeElement','isLiteralTypeLiteral','notImplementedResolver']
named=[]
accessor='a narrowed accessor reread, which can return a different value'
for name in names:
 ws=sorted({w for s in 'AB' for w,u in units[s].items() if u.get('name')==name})
 assert ws,name
 for w in ws:named.append({'name':name,'unit':w,'A':state('A',w),'B':state('B',w),'B_narrowed_accessor_reread':state('B',w).get('reason')==accessor})
dump('named-units.json',named)
rows=[]
for reason in sorted({x[2] for s in 'AB' for x in sites[s] if x[0] in ('NotYet','Refused')}):
 a=sorted(x for x in sites['A'] if x[0] in ('NotYet','Refused') and x[2]==reason)
 b=sorted(x for x in sites['B'] if x[0] in ('NotYet','Refused') and x[2]==reason)
 if len(a)!=len(b):rows.append({'reason':reason,'A':len(a),'B':len(b),'delta':len(b)-len(a),'callout':any(x in reason.lower() for x in ('view','branded supertype','predicate')),'A_examples':[x[1] for x in a[:5]],'B_examples':[x[1] for x in b[:5]],'removed_examples':[x[1] for x in a if x not in b][:5],'added_examples':[x[1] for x in b if x not in a][:5]})
dump('changed-reasons.json',rows)
lines=['| Reason | A | B | B−A | A examples (up to five) | B examples (up to five) |','|---|---:|---:|---:|---|---|']
esc=lambda x:str(x).replace('|','\\|').replace('\n',' ')
for r in rows:lines.append('| '+ ' | '.join([('**CALLOUT** ' if r['callout'] else '')+esc(r['reason']),str(r['A']),str(r['B']),str(r['delta']),'<br>'.join(r['A_examples']),'<br>'.join(r['B_examples'])])+' |')
(out/'changed-reasons.md').write_text('\n'.join(lines)+'\n')
times={s:{norm(t['where']):t for t in read(s+'-times.jsonl')} for s in 'AB'}
assert times['A'].keys()==times['B'].keys()
timing=[]
for w,a in times['A'].items():
 b=times['B'][w]; av=a['wall_seconds']; bv=b['wall_seconds']
 timing.append({'unit':w,'name':units['A'][w].get('name',''),'A_seconds':av,'B_seconds':bv,'delta_seconds':bv-av,'ratio':bv/av if av else None,'A_cpu_seconds':a['cpu_seconds'],'B_cpu_seconds':b['cpu_seconds'],'A_state':state('A',w)['status'],'B_state':state('B',w)['status']})
timing.sort(key=lambda t:(-t['delta_seconds'],t['unit']))
with (out/'unit-times.csv').open('w') as f:
 writer=csv.DictWriter(f,fieldnames=list(timing[0]),lineterminator="\n");writer.writeheader();writer.writerows(timing)
totals={s:{k:sum(t[k] for t in times[s].values()) for k in ('wall_seconds','cpu_seconds')} for s in 'AB'}
dump('timing-summary.json',{'units':len(timing),'totals':totals,'top20':timing[:20]})
lines=['| Unit | Name | A seconds | B seconds | B−A seconds | B/A |','|---|---|---:|---:|---:|---:|']
for t in timing[:20]:lines.append(f"| {t['unit']} | {t['name']} | {t['A_seconds']:.6f} | {t['B_seconds']:.6f} | {t['delta_seconds']:+.6f} | {t['ratio']:.3f} |")
(out/'timing-top20.md').write_text('\n'.join(lines)+'\n')
print(json.dumps({'totals':{s:summaries[s]['totals'] for s in 'AB'},'regressions':len(regress),'changed_reasons':len(rows),'removed_lowering_sites':len(removed),'timing_totals':totals,'named':named},indent=2))
# A source-only negative control: namespace import/export units contain no field
# reads, writes, contracts or predicates for this change to alter.
controls=[t for t in timing if '/_namespaces/ts.ts:' in t['unit']]
assert len(controls)==77
assert all(units['A'][t['unit']]['kind'] in ('KindImportDeclaration','KindExportDeclaration') for t in controls)
def quant(values,p):
 values=sorted(values);return values[round((len(values)-1)*p)]
deltas=[t['delta_seconds'] for t in controls]
absd=list(map(abs,deltas))
noise={'definition':'77 namespace import/export units: no field/contract/predicate lowering; includes shared snapshot, GC and scheduler noise','n':len(controls),'median_delta_seconds':statistics.median(deltas),'median_abs_delta_seconds':statistics.median(absd),'p90_abs_delta_seconds':quant(absd,.9),'p95_abs_delta_seconds':quant(absd,.95),'max_abs_delta_seconds':max(absd),'median_ratio':statistics.median(t['ratio'] for t in controls),'min_ratio':min(t['ratio'] for t in controls),'max_ratio':max(t['ratio'] for t in controls),'A_seconds':sum(t['A_seconds'] for t in controls),'B_seconds':sum(t['B_seconds'] for t in controls),'units':controls}
dump('noise-floor.json',noise)
changed=[];success={s:sum(state(s,w)['status']=='lowers' for w in units[s]) for s in 'AB'}
for w in sorted(units['A'].keys()|units['B'].keys()):
 if state('A',w)!=state('B',w):changed.append({'unit':w,'name':units['A'].get(w,{}).get('name',''),'A':state('A',w),'B':state('B',w)})
dump('unit-comparison.json',{'unit_counts':{s:len(units[s]) for s in 'AB'},'lowers':success,'changed':changed,'regressions':regress})
special=[]
for reason in sorted({r['reason'] for s in 'AB' for r in summaries[s]['reason_rows']}):
 if any(word in reason.lower() for word in ('view','predicate','branded supertype')):
  special.append({'reason':reason,**{s:next((r['count'] for r in summaries[s]['reason_rows'] if r['reason']==reason),0) for s in 'AB'}})
dump('view-predicate-reasons.json',special)
lines=['| Reason mentioning view, branded supertype, or predicate | A | B |','|---|---:|---:|']
for r in special:lines.append(f"| {esc(r['reason'])} | {r['A']} | {r['B']} |")
(out/'view-predicate-reasons.md').write_text('\n'.join(lines)+'\n')

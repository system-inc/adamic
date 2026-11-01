import collections, importlib.util, json, pathlib, re
B=pathlib.Path('/workspace/scratch/axolotl')
T=B/'adapted'
spec=importlib.util.spec_from_file_location('meter','/workspace/adamic/stage3/meter/report.py')
meter=importlib.util.module_from_spec(spec);spec.loader.exec_module(meter)
MODES=['main','es2020','es2020-capped-node']
def norm(s):return s.replace(str(T)+'/', '').replace('/workspace/adamic/cohere/TypeScript/tsc/internal/bundled/libs/','checker-libs/')
def records(p):return [json.loads(x) for x in p.read_text().splitlines()]
def esc(s):return s.replace('|','&#124;').replace('\n',' ')
expected={str(p.resolve()) for p in (T/'src/compiler').rglob('*') if p.is_file()}
allsets={}; reasons={}; sites={}; summary={}
for mode in MODES:
 d=B/mode; rows=records(d/'census.jsonl'); entry=meter.entry_summary(T,d/'tsc')
 own=set(); external=set(); unlocated=set()
 for r in rows:
  for s in r.get('diagnostics',[]):
   m=re.match(r'^(.+?):\d+:\d+: error TS\d+:',s)
   if m and m[1] in expected:own.add(norm(s))
   elif m:external.add(norm(s))
   else:unlocated.add(norm(s))
 whole=set(map(norm,rows[-1].get('diagnostics',[])))
 entryset=set(map(norm,entry['diagnostics']))
 latent=records(d/'latent.jsonl')
 assert latent[0]['latent_mode']=='first-error'
 measure=latent[0]['measurement']
 lower=meter.lowering_summary(latent,{str(p.resolve()) for p in (T/'src/compiler').rglob('*.ts')},measure)
 lower['unit_status_counts']=dict(collections.Counter(u['status'] for r in latent[1:] for u in r.get('units',[])))
 ss={tuple(norm(str(f[k])) for k in ['kind','where','reason','text']) for r in latent[1:] for f in r.get('findings',[])}
 sites[mode]=ss
 reasons[mode]=collections.Counter((k,r) for k,w,r,t in ss if k in ('NotYet','Refused'))
 allsets[mode]={'entry':entryset,'compiler_whole':whole,'own_file':own,'external':external,'unlocated':unlocated}
 summary[mode]={'entry_diagnostics':len(entryset),'compiler_whole_diagnostics':len(whole),'own_file_sum':len(own),'external_diagnostics_union':len(external),'unlocated_diagnostics_union':len(unlocated),'latent':lower}
 (d/'diagnostics.json').write_text(json.dumps({k:sorted(v) for k,v in allsets[mode].items()},indent=2)+'\n')
 (d/'latent-summary.json').write_text(json.dumps(lower,indent=2)+'\n')
(B/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
lines=['Measured on checker-rejected programs. Latent mode: first-error (`LATENT_FULL=0`).','', '| Measurement | Main | ES2020 | ES2020, capped Node references |','| --- | ---: | ---: | ---: |']
for key,label in [('entry_diagnostics','tsc entry diagnostics'),('compiler_whole_diagnostics','Compiler aggregate diagnostics'),('own_file_sum','Own-file diagnostic sum'),('external_diagnostics_union','External diagnostic union'),('unlocated_diagnostics_union','Unlocated diagnostic union')]:
 lines.append('| '+label+' | '+' | '.join(str(summary[m][key]) for m in MODES)+' |')
for k in ['NotYet','Refused','SkippedDependency','error','panic']:
 lines.append('| Latent distinct '+k+' | '+' | '.join(str(summary[m]['latent']['totals'][k]) for m in MODES)+' |')
for status in sorted(set().union(*(summary[m]['latent']['unit_status_counts'].keys() for m in MODES))):
 lines.append('| Units '+status+' | '+' | '.join(str(summary[m]['latent']['unit_status_counts'].get(status,0)) for m in MODES)+' |')
lines+=['','Own-file sum follows the meter: union all census diagnostics by primary compiler-file location, then sum each file\'s unique diagnostics. It is not a sum of imported repeats. Latent sites use the meter\'s distinct `(kind, where, reason, text)` key.','']
for mode in MODES[1:]:
 lines+=['Comparison: main to '+mode+'.','']
 for scope in ['entry','compiler_whole','own_file','external','unlocated']:
  before=allsets['main'][scope];after=allsets[mode][scope]
  lines+=['Diagnostics, '+scope+': '+str(len(before-after))+' disappear; '+str(len(after-before))+' appear.','']
  for verb,values in [('Disappears',before-after),('Appears',after-before)]:
   for s in sorted(values):
    lines+=['```text',verb+': '+s,'```','']
 lines+=['Every changed latent reason row:','', '| Kind | Reason | Main | '+mode+' | Example removed | Example added |','| --- | --- | ---: | ---: | --- | --- |']
 changed=[]
 for k,r in sorted(reasons['main'].keys()|reasons[mode].keys()):
  a=reasons['main'][k,r];b=reasons[mode][k,r]
  if a==b:continue
  removed=sorted(w for kk,w,rr,t in sites['main']-sites[mode] if (kk,rr)==(k,r))
  added=sorted(w for kk,w,rr,t in sites[mode]-sites['main'] if (kk,rr)==(k,r))
  lines.append(f'| {k} | {esc(r)} | {a} | {b} | {esc(removed[0]) if removed else ""} | {esc(added[0]) if added else ""} |')
  changed.append({'kind':k,'reason':r,'before':a,'after':b,'removed_examples':removed[:3],'added_examples':added[:3]})
 if not changed:lines+=['| No changed rows | | | | | |']
 (B/(mode+'-latent-changes.json')).write_text(json.dumps({'reason_rows':changed,'removed_sites':sorted(sites['main']-sites[mode]),'added_sites':sorted(sites[mode]-sites['main'])},indent=2)+'\n')
 lines+=['']
(B/'tables.md').write_text('\n'.join(lines).rstrip()+'\n')
print(json.dumps({m:{k:v for k,v in s.items() if k!='latent'}|{'latent_totals':s['latent']['totals']} for m,s in summary.items()},indent=2))

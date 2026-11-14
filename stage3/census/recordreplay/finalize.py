"""Audit complete observations and publish the final report, never submit to Loom."""
from pathlib import Path
import json,gzip,subprocess as s,hashlib,time,os
D=Path('/workspace/scratch/recordreplay');R=D/'base';O=R/'stage3/census/recordreplay'
def load(p):return json.loads(p.read_text())
def stop(row):return (row.get('where','-'),row['kind'],row.get('reason','-'))
def escape(x):return str(x).replace('|','\\|').replace('\n',' ')
def display(f):
 if not f:return 'No finding'
 return f"{f['where'].split('/src/')[-1]}: {f['kind']}: {f['reason']}"
while True:
 a=(D/'ordinary.log').read_text();p=O/'first-error/summary.json'
 if 'ordinary tables complete' in a and p.exists() and load(p).get('complete'):break
 time.sleep(5)
s.run(['python3',str(O/'native-source.py')],cwd=R,check=True)
first=load(O/'first-error/summary.json');assert len(first['pieces'])==12 and all(p['complete'] and p['seconds']<3600 for p in first['pieces'])
root_moves=[];ordinary_resources=[];legends={};tables=[]
for pair in ['A','B']:
 for mode in ['CommonJSBundle','EcmaScript']:
  data=load(O/'ordinary'/(pair+'-'+mode+'-paired.json'));assert data['complete'] and len(data['rows'])==79
  lines=[f'# {pair}, {mode}: every root paired','', '| Root | Main first stop | Head first stop |','|---|---|---|']
  for row in data['rows']:
   before=stop(row['before']);after=stop(row['after'])
   if before!=after:root_moves.append(dict(pair=pair,module_format=mode,**row))
   for item in [before,after]:
    if item not in legends:legends[item]='S'+str(len(legends)+1)
   lines.append('| '+escape(row['root'])+' | '+legends[before]+' | '+legends[after]+' |')
  lines+=['','| Stop | Location | Kind | Exact reason |','|---|---|---|---|']
  lines += ['| '+v+' | '+' | '.join(map(escape,k))+' |' for k,v in legends.items()]
  dest=O/'ordinary'/(pair+'-'+mode+'-every-root.md');dest.write_text('\n'.join(lines)+'\n');tables.append(str(dest.relative_to(O)))
  for side in ['main','head']:
   folder=O/'ordinary'/(pair+'-'+side)/mode
   resources=[]
   for i in range(1,80):
    observations=[json.loads(l) for l in (folder/f'root-{i:02}.jsonl').read_text().splitlines()];assert len(observations)==2
    assert {k:v for k,v in observations[0].items() if k!='seconds'}=={k:v for k,v in observations[1].items() if k!='seconds'}
    r=load(folder/f'root-{i:02}-resource.json');assert r['exit']==0;resources.append(r)
   ordinary_resources.append({'pair':pair,'side':side,'module_format':mode,'root_children':79,'observations':158,'peak_rss_kib':max(r['peak_rss_kib'] for r in resources),'seconds':sum(r['seconds'] for r in resources)})
unit_moves={p:[] for p in ['A','B']};process_new={p:{'raw':[],'distinct':[]} for p in ['A','B']};unit_counts={};latent_resources=[]
for pair in ['A','B']:
 for side in ['base','head']:
  units={};file_names=set()
  for g in range(6):
   folder=O/'first-error'/(pair+f'-group-{g}');rows=[json.loads(l) for l in gzip.decompress((folder/(side+'.jsonl.gz')).read_bytes()).decode().splitlines()]
   assert rows[0]['status']=='measurement' and not rows[0]['checker_rejected'] and rows[-1]['status']=='relation_report'
   r=load(folder/(side+'-resource.json'));assert r['exit']==0;latent_resources.append({'pair':pair,'group':g,'side':side,**r})
   for record in rows[1:-1]:
    file=record['file'].split('/src/')[-1];assert file not in file_names;file_names.add(file)
    for u in record['units']:
     key=u['where'].split('/src/')[-1];assert key not in units;units[key]=u
  assert len(file_names)==81;unit_counts[pair+'-'+side]={'files':81,'units':len(units),'attempted':sum(u['status']=='attempted' for u in units.values()),'statuses':{status:sum(u['status']==status for u in units.values()) for status in sorted({u['status'] for u in units.values()})}}
 for g in range(6):
  delta=load(O/'first-error'/(pair+f'-group-{g}')/'delta.json')
  unit_moves[pair]+=[{'group':g,**v} for v in delta['distinct_reads']['unit_first_stop_delta']]
  for name,key in [('raw_findings','raw'),('distinct_reads','distinct')]:
   process_new[pair][key]+=[{'group':g,**v} for v in delta[name]['new_findings'] if 'process' in (v['finding'].get('reason','')+' '+v['finding'].get('text','')).lower()]
latent={'unit_counts':unit_counts,'moved_units':unit_moves,'new_process_findings':process_new,'resources':latent_resources,'all_sources_covered':True,'groups_disjoint':True}
(O/'first-error/complete.json').write_text(json.dumps(latent,indent=2)+'\n')
lines=['# Every moved unit and every new process finding','']
for pair in ['A','B']:
 lines += ['## '+pair,'','| Unit | Main first finding | Head first finding |','|---|---|---|']
 for row in unit_moves[pair]:lines.append('| '+escape(row['unit']['where'].split('/src/')[-1])+' | '+escape(display(row.get('base')))+' | '+escape(display(row.get('head')))+' |')
 if not unit_moves[pair]:lines+=['','No unit first stop moved.']
 for key in ['raw','distinct']:
  lines+=['',key+' new findings naming process:','']
  for row in process_new[pair][key]:lines+=['```json',json.dumps(row,indent=2),'```']
  if not process_new[pair][key]:lines+=['None.']
(O/'first-error/every-change.md').write_text('\n'.join(lines)+'\n')
violations=[r for r in root_moves if not any(r['before'].get('where','').endswith('sys.ts:1588:'+c) for c in ['191','192']) or 'process' in r['after'].get('reason','').lower()]
verdict={'complete':True,'ordinary_root_children':632,'ordinary_observations':1264,'ordinary_tables':8,'paired_rows':316,'root_first_stop_changes':root_moves,'lower_condition_holds':not violations and not process_new['B']['distinct'],'violations':violations,'ordinary_resources':ordinary_resources,'latent_pieces':12,'latent_units':unit_counts,'new_process_findings':{p:{k:len(v) for k,v in data.items()} for p,data in process_new.items()},'unit_first_stop_changes':{p:len(v) for p,v in unit_moves.items()},'native_B_all_unit_source_comparison':load(O/'first-error/B-native-source.json'),'never_submitted_to_loom':True}
(O/'complete.json').write_text(json.dumps(verdict,indent=2)+'\n')
report=(O/'REPORT.md').read_text();start=report.index('| Pair |');end=report.index('The read gate holds')
intro=['| Pair / format | Moved root first stops, row by row |','|---|---|']
for pair in ['A','B']:
 for mode in ['CommonJSBundle','EcmaScript']:
  moved=[v for v in root_moves if v['pair']==pair and v['module_format']==mode]
  if not moved:intro.append(f'| {pair} / {mode} | None; 79/79 paired |')
  else:
   for v in moved:intro.append('| '+pair+' / '+mode+' / '+v['root']+' | '+escape(str(stop(v['before']))+' -> '+str(stop(v['after'])))+' |')
intro+=['','All requested censuses are complete. '+('Lower’s condition holds.' if verdict['lower_condition_holds'] else 'Lower’s condition fails; see complete.json for every violation.')+' Never submitted to Loom.','', 'The 79-root tables contain 316 paired rows across A/B and both formats, from 632 serial children and 1,264 observations. Every root was measured twice.','']
report=report[:start]+'\n'.join(intro)+'\n'+report[end:]
report+='\nThe ordinary B table uses original main source against amended head source.\nThe controlled same-input B latent pairs remain separate from the native\nsource-changing comparison. See [full-sys/B-native-every-change.md](full-sys/B-native-every-change.md)\nand [first-error/B-native-every-change.md](first-error/B-native-every-change.md)\nfor every native B change and the separately reported new erased import unit.\n\n## Complete first-error census\n\nThe 81 resolved tsc implementation sources are covered exactly once per side,\nin six disjoint file groups for each pair. All 12 paired pieces completed below\n60 minutes, and each was committed and pushed as it finished. Each child retains\nthe full tsc entry, source graph, normal prepasses and no-output guards.\nGeneric call witnesses and relation trailers remain per child; no synthetic\naggregate relation report is claimed.\n\n| Pair | Units on each side | Moved unit first stops | New process findings, raw / distinct |\n|---|---:|---:|---:|\n'
for pair in ['A','B']:
 assert unit_counts[pair+'-base']['units']==unit_counts[pair+'-head']['units'];report+=f"| {pair} | {unit_counts[pair+'-base']['units']} | {len(unit_moves[pair])} | {len(process_new[pair]['raw'])} / {len(process_new[pair]['distinct'])} |\n"
report+='\n[first-error/every-change.md](first-error/every-change.md) lists every moved unit\nand every new finding naming process in full. Each piece retains both raw streams,\nthe read-paired delta, complete unit eligibility, actual exit, deadline and peak RSS.\n\n## Every root paired\n\n'
report+='\n'.join('- ['+p+']('+p+')' for p in tables)+'\n\n## Peak memory\n\nPeak RSS is measured for each serial child, in KiB.\n\n| Census | Peak RSS, KiB |\n|---|---:|\n'
for r in ordinary_resources:report+=f"| {r['pair']} {r['side']} {r['module_format']}, 79-root table | {r['peak_rss_kib']} |\n"
for pair in ['A','B']:
 for side in ['base','head']:report+=f"| {pair} {side}, all-unit pieces | {max(v['peak_rss_kib'] for v in latent_resources if v['pair']==pair and v['side']==side)} |\n"
(O/'REPORT.md').write_text(report)
setup=load(O/'setup.json');setup['status']='complete';setup['amendment']='e00ce47c6';(O/'setup.json').write_text(json.dumps(setup,indent=2)+'\n')
s.run(['git','diff','--check'],cwd=R,check=True);s.run(['git','add','stage3/census/recordreplay'],cwd=R,check=True);s.run(['git','commit','-m','stage3: finish recordreplay paired census report and coverage audit'],cwd=R,check=True);s.run(['git','push','origin','HEAD:refs/heads/typescript/recordreplay-census'],cwd=R,check=True);print(json.dumps(verdict),flush=True)

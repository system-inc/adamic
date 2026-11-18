#!/usr/bin/env python3
"""Parse whole stderr without deleting lines; bind probe records to next canonical unit completion."""
import json,re,collections,sys
from pathlib import Path
root=Path(sys.argv[1]);out=Path(sys.argv[2]);out.mkdir(parents=True,exist_ok=True)
allrecords=[];summaries={}
targets=json.loads(Path(sys.argv[3]).read_text()) if len(sys.argv)>3 else []
target_ids={t.get('where'):t['id'] for t in targets}
unit_names={};unit_targets=collections.defaultdict(set)
for stream in root.glob('*/latent.jsonl'):
 for line in stream.read_text().splitlines():
  try:record=json.loads(line)
  except Exception:continue
  for u in record.get('units',[]):unit_names[u['where']]=u.get('name','')
  for f in record.get('findings',[]):
   if f.get('where') in target_ids:unit_targets[f.get('unit')].add((target_ids[f['where']],f['where']))
for run in sorted(root.iterdir()):
 if not run.is_dir() or not (run/'stderr.log').exists():continue
 pending=[];records=[];times=[];last=None
 for no,raw in enumerate((run/'stderr.log').read_bytes().splitlines(),1):
  line=raw.decode('utf8',errors='replace')
  if line.startswith(('DEPTHGATE ','DEPTHWALK ')):
   rec={'run':run.name,'stderr_line':no,'line':line,'type':line.split(' ',1)[0],'unit':None,'attribution':'pending'}
   m=re.search(r'pastDepth=(\d+)',line)
   if m:rec['pastDepth']=int(m.group(1))
   if rec['type']=='DEPTHGATE':
    stop=line.split(' stop=',1)[1]
    try:rec['stop']=json.loads(stop)
    except Exception:rec['stop_quoted']=stop
   records.append(rec);pending.append(rec)
  elif line.startswith('{'):
   try:t=json.loads(line)
   except Exception:continue
   if {'file','where','wall_seconds','cpu_seconds'}<=t.keys():
    times.append(t)
    for r in pending:r.update(unit=t['where'],attribution='next canonical unit completion in the same pipe')
    pending=[];last=t['where']
 for r in pending:r.update(attribution='unfinished interval after '+str(last))
 for r in records:
  r['unit_name']=unit_names.get(r['unit'],'')
  r['target_sites_in_unit']=[{'id':i,'where':w} for i,w in sorted(unit_targets.get(r['unit'],set()))]
 gates=[r for r in records if r['type']=='DEPTHGATE'];walks=[r for r in records if r['type']=='DEPTHWALK'];freq=collections.Counter(r['pastDepth'] for r in gates)
 summaries[run.name]={'DEPTHGATE_lines':len(gates),'DEPTHWALK_lines':len(walks),'largest_pastDepth':max(freq,default=None),'pastDepth_distribution':dict(sorted(freq.items())),'above_1000':sum(r['pastDepth']>1000 for r in gates),'unattributed_probe_lines':sum(r['unit'] is None for r in records),'completed_timed_units':len(times),'stops':dict(collections.Counter(r.get('stop','<unparsed>') for r in gates))}
 (out/(run.name+'.unit-times.jsonl')).write_text(''.join(json.dumps(t)+'\n' for t in times))
 allrecords.extend(records)
(out/'probe-lines.jsonl').write_text(''.join(json.dumps(r)+'\n' for r in allrecords))
(out/'probe-summary.json').write_text(json.dumps(summaries,indent=2)+'\n')
freq=collections.Counter(r['pastDepth'] for r in allrecords if r['type']=='DEPTHGATE')
(out/'pastDepth-distribution.tsv').write_text('pastDepth\tcount\n'+''.join(f'{n}\t{c}\n' for n,c in sorted(freq.items())))
for label,filtered in [('above-1000',[r for r in allrecords if r['type']=='DEPTHGATE' and r['pastDepth']>1000]),('DEPTHWALK',[r for r in allrecords if r['type']=='DEPTHWALK'])]:
 lines=['# '+label+' probe lines\n\nEvery literal stderr line is retained below. The containing unit is identified by the next canonical timing record in the same pipe. The probe does not print an exact comparison-node position; listed target sites are context within that unit, not guessed origins of a line.\n']
 if not filtered:lines.append('None.\n')
 for i,r in enumerate(filtered,1):
  context=', '.join(t['id']+' '+t['where'] for t in r['target_sites_in_unit']) or 'No historical target site in this unit.'
  lines += ['## '+str(i)+'. '+r['run']+' stderr line '+str(r['stderr_line'])+'\n','Containing unit: `'+str(r['unit'] or r['attribution'])+'` ('+r['unit_name']+').\n','Target context: '+context+'\n','````text\n'+r['line']+'\n````\n']
 (out/(label+'.md')).write_text('\n'.join(lines)+'\n')
print(json.dumps(summaries,indent=2))

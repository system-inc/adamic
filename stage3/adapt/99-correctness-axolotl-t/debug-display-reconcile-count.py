from pathlib import Path
import gzip,hashlib,json,subprocess
root=Path('/workspace/adamic');scratch=Path('/workspace/scratch');work=scratch/'debug-display-lane'
plan=json.loads((root/'stage3/lane/shard-plan.json').read_text())
assert (work/'summary.json').exists(),'controller still running'
controller=json.loads((work/'summary.json').read_text())
assert len(controller['results'])==398
accepted=[];raw=[];totals={name:0 for name in ['pass','fail','skip']}
for unit in plan['units']:
 path=(scratch/'debug-display-byte-retry' if unit['name']=='stage3-lane-byte-comparisons' else work/'pieces')/unit['name']/'report.json'
 report=json.loads(path.read_text());assert report['status']=='pass' and not report['errors'],(unit['name'],report['errors'])
 assert report['counts']==unit['expect']
 for k,v in report['counts'].items():totals[k]+=v
 accepted.append({'unit':unit['name'],'status':report['status'],'counts':report['counts'],'child_exit':report['upstream_exit'],'report_sha256':hashlib.sha256(path.read_bytes()).hexdigest(),'inputs_hash':report['inputs_hash']})
 raw.append(report)
assert totals==plan['counts'],(totals,plan['counts'])
names=['attachFlowNodeDebugInfo','attachNodeArrayDebugInfo','enableDebugInfo'];counts=dict.fromkeys(names,0);processes=[]
for path in [work/'counts.jsonl',scratch/'debug-display-byte-retry-counts.jsonl']:
 for line in path.read_text().splitlines():
  row=json.loads(line);processes.append(row)
  for k,v in row['counts'].items():assert v>=0;counts[k]+=v
assert counts==dict.fromkeys(names,0)
control=json.loads((scratch/'debug-display-control-valid-counts.jsonl').read_text());plain=json.loads((scratch/'debug-display-control-plain-counts.jsonl').read_text())
assert all(v>0 for v in control['counts'].values());assert not any(plain['counts'].values())
for stream in ['stdout','stderr']:assert (scratch/f'debug-display-control-valid.{stream}').read_bytes()==(scratch/f'debug-display-control-plain.{stream}').read_bytes()==b''
before=scratch/'debug-display-before';counted=scratch/'debug-display-counted'
manifest=json.loads((counted/'debug-display-instrumentation.json').read_text());js=[]
for p in sorted((before/'built/local').glob('*.js')):
 original=p.read_bytes();observed=(counted/'built/local'/p.name).read_bytes();restored=observed
 for name in names:restored=restored.replace(('\n globalThis.__debugDisplayCount('+json.dumps(name)+');\n').encode(),b'')
 assert restored==original,p
 js.append({'file':p.name,'baseline_sha256':hashlib.sha256(original).hexdigest(),'counted_sha256':hashlib.sha256(observed).hexdigest(),'only_counter_insertions':True})
report={'base':subprocess.check_output(['git','rev-parse','origin/main'],cwd=root,text=True).strip(),'typescript_source':subprocess.check_output(['git','rev-parse','HEAD'],cwd='/workspace/typescript-6.0.3',text=True).strip(),'units':398,'ordinary_totals':totals,'counters':counts,'process_records':len(processes),'includes_failed_attempt_zero_records':True,'forced_debug_control':{'child_exit':0,'counts':control['counts'],'stdout_bytes':0,'stderr_bytes':0},'ordinary_control':{'child_exit':0,'counts':plain['counts'],'stdout_bytes':0,'stderr_bytes':0},'preliminary_attempts':[{'kind':'API shard metadata omission','child_exit':1,'retry_child_exit':0},{'kind':'small control lacked ignoreConfig','child_exit':1,'retry_child_exit':0}],'controller_child_exit':1,'reconciliation_child_exit':0,'instrumentation':manifest,'emitted_js':js,'accepted':accepted}
(work/'reconciled.json').write_text(json.dumps(report,indent=2)+'\n')
with gzip.open(work/'accepted-reports.jsonl.gz','wt') as f:
 for row in raw:f.write(json.dumps(row)+'\n')
with gzip.open(work/'process-counts.jsonl.gz','wt') as f:
 for row in processes:f.write(json.dumps(row)+'\n')
print(json.dumps({'units':398,'ordinary_totals':totals,'counters':counts,'positive_control':control['counts'],'reconciliation_child_exit':0}))

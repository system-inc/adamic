import json,pathlib,subprocess
r=pathlib.Path(__file__).resolve().parent;xs=json.loads((r/'matrix.json').read_text());cs=json.loads((r/'skipped.json').read_text());keep=[];deletable=[]
for c in cs:
 guards=[x for x in xs if c in x['candidates_failed']];lost=[x for x in guards if not x['stale'] and not x.get('broken') and x.get('complete') and not x['still_caught_by']]
 if lost:keep.append({'test':c,'because':', '.join(x['mutant']+' ('+x['branch']+')' for x in lost)+' loses its last catcher when both candidates are skipped. Preserve at least one candidate; prior evidence says both catch these mutants.'})
 elif all(x.get('still_caught_by') for x in guards):deletable.append(c)
report={'package':'stage1/typescript/scanner','main':subprocess.check_output(['git','rev-parse','origin/main']).decode().strip(),'skipped':cs,'mutants':[{k:x[k] for k in ['mutant','file_line','branch','candidates_failed','still_caught_by','stale']} for x in xs],'keep':keep,'deletable':deletable,'witness_failures':{x['branch']+':'+x['mutant']:x['witness_failures'] for x in xs if x['witness_failures']},'panicking_tests':{x['branch']+':'+x['mutant']:x['panicking_tests'] for x in xs if x['panicking_tests']},'replay_wall_seconds':round(sum(z['wall_seconds'] for x in xs for z in x['runs']),3)}
(r/'report.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report,indent=2))

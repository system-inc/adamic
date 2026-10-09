import json,pathlib,gzip
P=pathlib.Path(__file__).resolve().parent
items=json.loads((P/'mutant-list.json').read_text());matrix=json.loads((P/'matrix.json').read_text())
assert len(items)==len(matrix)==77,'Replay is not complete'
assert json.loads((P/'baseline-status.json').read_text())['exit']==0
byindex={x['replay_index']:x for x in matrix};assert set(byindex)==set(range(len(items)))
witnesses=set(json.loads((P/'witnesses.json').read_text()));actual_skip=set(json.loads((P/'expanded-skipped.json').read_text()))
lines=json.loads((P/'changed-lines.json').read_text())
# Candidate rows and family members are the initial audit scope, not inferred from failures.
scope=json.loads((P/'candidates.json').read_text())
print('scope shape',type(scope))
mutants=[];lost={};unknown={}
for i,item in enumerate(items):
 r=byindex[i]
 catches=r.get('still_caught_by',[])
 assert not set(catches)&witnesses
 assert not set(catches)&actual_skip
 m={'mutant':item['mutant'],'diff':item['saved_diff'],'file_line':lines[item['saved_diff']],'branch':item['branch'],'candidates_failed':item['candidates_failed'],'still_caught_by':catches,'stale':r['stale']}
 if r.get('witness_failures'):m['witness_failures']=r['witness_failures']
 if r.get('panicking_tests'):m['panicking_tests']=r['panicking_tests']
 if r.get('broken'):m['broken']=r['broken']
 m['wall_seconds']=r.get('wall_seconds',0)
 mutants.append(m)
 for candidate in item['candidates_failed']:
  if r.get('completed_without_catcher'):lost.setdefault(candidate,[]).append(item['saved_diff'])
  elif r['stale'] or r.get('broken') or not catches:unknown.setdefault(candidate,[]).append(item['saved_diff'])
# Skip correctness is checked on every run, including a panic retry.
for r in matrix:
 for run in r.get('runs',[]):
  with (gzip.open(P/run['log'],'rt') if run['log'].endswith('.gz') else (P/run['log']).open()) as f:
   for line in f:
    try:event=json.loads(line)
    except ValueError:continue
    if event.get('Action')=='run':assert event.get('Test','').split('/')[0] not in actual_skip
report={'package':'internal/oracle','main':(P/'main.txt').read_text().strip(),'skipped':scope['rows'],'mutants':mutants,'keep':[],'deletable':[]}
for candidate in scope['rows']:
 if candidate in lost:report['keep'].append({'test':candidate,'because':', '.join(lost[candidate])+' loses its last non-witness catcher without the set'})
 elif candidate in unknown:report['keep'].append({'test':candidate,'because':'Replay was inconclusive for '+', '.join(unknown[candidate])})
 else:report['deletable'].append(candidate)
none=[c for c in scope['rows'] if not any(c in x['candidates_failed'] for x in items)]
report['notes']={'baseline_default_skips':json.loads((P/'baseline-default-skips.json').read_text()),'no_gathered_mutant':none,'no_gathered_mutant_reason':'No M*.diff or D*.diff was recorded as failing these candidates; they guard nothing demonstrated by this inventory.','scope':'Deletable is limited to these gathered mutants on this main commit.','evidence_branch':'test-defend/deletion-set/internal-oracle','evidence_path':'review/test-defend/deletion-set/internal-oracle/'}
(P/'report.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps({'mutants':len(mutants),'keep':report['keep'],'deletable':report['deletable']},indent=2))

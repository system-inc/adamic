import json,pathlib,collections
r=pathlib.Path('review/test-defend/deletion-set/internal-native');m=json.load(open(r/'matrix.json'));skipped=json.load(open(r/'skip.json'))
mutants=[];keep=[];deletable=[];notes=[]
for e in m['mutants']:
 runs=e.get('runs',[]);last=runs[-1] if runs else {}
 # A completed package run whose only failures are witnesses has no valid catcher.
 witness_only=bool(last.get('failed')) and set(last['failed'])<=set(last.get('witness_failures',[])) and not last.get('panic') and not last.get('stopped_after_catch')
 if witness_only:
  e.pop('broken',None);e['completed_without_valid_catch']=True
 if last.get('exit')==0:e['completed_without_valid_catch']=True
 mutants.append({k:e[k] for k in ['replay','mutant','file_line','branch','source_diff','candidates_failed','still_caught_by','stale','witness_failures','panicking_tests','broken','completed_without_valid_catch','build_failure_rows'] if k in e})
for t in skipped:
 relevant=[e for e in m['mutants'] if t in e['candidates_failed']]
 lost=[e for e in relevant if e.get('completed_without_valid_catch')]
 uncertain=[e for e in relevant if e.get('stale') or e.get('broken')]
 if lost:keep.append(dict(test=t,because=', '.join(e['replay']+'/'+e['mutant'] for e in lost)+' lose their last non-witness catcher without it'))
 elif uncertain:keep.append(dict(test=t,because='Replay unresolved for '+', '.join(e['replay'] for e in uncertain)+'; deletion is not established'))
 elif all(e.get('still_caught_by') for e in relevant):
  deletable.append(t)
  if not relevant:notes.append(t+': no qualifying recorded mutant, so no shown guard.')
 else:raise RuntimeError('unclassified '+t)
report=dict(package='internal/native',main=(r/'main.txt').read_text().strip(),skipped=skipped,mutants=mutants,keep=keep,deletable=deletable,notes=notes)
(r/'report.json').write_text(json.dumps(report,indent=2)+'\n');(r/'matrix.json').write_text(json.dumps(m,indent=2)+'\n')
passes=[];skips=[]
for line in pathlib.Path(m['baseline']['log']).read_text().splitlines():
 try:e=json.loads(line)
 except:continue
 if e.get('Test') and '/' not in e['Test']:
  if e['Action']=='pass':passes.append(e['Test'])
  if e['Action']=='skip':skips.append(e['Test'])
(r/'baseline-rows.json').write_text(json.dumps(dict(passed=passes,skipped=skips),indent=2)+'\n')
time=sum(v['seconds'] for e in m['mutants'] for v in e.get('runs',[]));(r/'timings.json').write_text(json.dumps(dict(baseline_wall_seconds=m['baseline']['seconds'],replay_wall_seconds=round(time,3),total_test_wall_seconds=round(time+m['baseline']['seconds'],3)),indent=2)+'\n')
print(json.dumps(report,indent=2));print('TIMES',m['baseline']['seconds'],time)

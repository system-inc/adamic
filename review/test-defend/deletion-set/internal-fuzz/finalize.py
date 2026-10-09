import pathlib,json,subprocess
R=pathlib.Path('review/test-defend/deletion-set/internal-fuzz');m=json.load(open(R/'matrix.json'));C=json.load(open(R/'skip.json'));items=json.load(open(R/'mutant-list.json'));assert len(m['mutants'])==len(items)
mutants=[];evidence=[]
for e in m['mutants']:
 mutants.append({k:e[k] for k in ['mutant','file_line','branch','source_diff','candidates_failed','still_caught_by','stale','witness_failures','pin_failures','panicking_tests','broken'] if k in e})
 for run in e['runs']:
  lines=[]
  for raw in pathlib.Path(run['log']).read_text().splitlines():
   try:v=json.loads(raw)
   except:continue
   if v.get('Test','').split('/')[0] in run['failed'] and v.get('Action')=='output':
    s=v.get('Output','')
    if '.go:' in s or 'FAIL:' in s:lines.append(s.rstrip())
  evidence.append(dict(mutant=e['mutant'],command=run['command'],rows_failed=run['failed'],failure_lines=lines))
keep=[];deletable=[];notes=[]
for t in C:
 reached=[e for e in m['mutants'] if t in e['candidates_failed']]
 lost=[e for e in reached if e.get('completed_without_valid_catch')]
 unknown=[e for e in reached if e['stale'] or e.get('broken')]
 if lost:keep.append(dict(test=t,because=', '.join(e['mutant'] for e in lost)+' loses its last valid catcher without it'))
 elif unknown:keep.append(dict(test=t,because='Unresolved replay: '+', '.join(e['mutant'] for e in unknown)))
 else:
  assert all(e['still_caught_by'] for e in reached);deletable.append(t)
  if not reached:notes.append(t+': no qualifying historical mutant')
notes.append('Candidate-specific defender diffs use G/R/S names instead of D; included as supplemental evidence.')
report=dict(package='internal/fuzz',main=(R/'main.txt').read_text().strip(),skipped=C,mutants=mutants,keep=keep,deletable=deletable,notes=notes)
(R/'report.json').write_text(json.dumps(report,indent=2)+'\n');(R/'failure-evidence.json').write_text(json.dumps(evidence,indent=2)+'\n')
rows=dict(passed=[],skipped=[])
for line in (R/'logs/baseline.log').read_text().splitlines():
 try:e=json.loads(line)
 except:continue
 if e.get('Test') and '/' not in e['Test']:
  if e['Action']=='pass':rows['passed'].append(e['Test'])
  if e['Action']=='skip':rows['skipped'].append(e['Test'])
(R/'baseline-rows.json').write_text(json.dumps(rows,indent=2)+'\n')
seconds=round(sum(run['wall_seconds'] for e in m['mutants'] for run in e['runs']),3)
(R/'timings.json').write_text(json.dumps(dict(baseline_binary_seconds=m['baseline']['binary_seconds'],mutant_wall_seconds=seconds),indent=2)+'\n')
assert subprocess.check_output(['git','diff']).decode()==''
print(json.dumps(report,indent=2));print('mutant wall',seconds,'baseline',m['baseline']['binary_seconds'])

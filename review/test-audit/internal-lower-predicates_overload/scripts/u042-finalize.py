import pathlib,json,collections,datetime,re,shutil
R=pathlib.Path('/workspace/adamic');E=R/'review/test-audit/internal-lower-predicates_overload';rows=json.load(open(E/'rows.json'));plans=json.load(open(E/'plan.json'));matrix=json.load(open(E/'matrix.json'));base=json.load(open(E/'base.json'))
for row in rows:
 if row['verdict']=='untrue':
  path=E/('P04-'+row['test']+'.log');fl=None
  for line in path.read_text().splitlines():
   try:e=json.loads(line)
   except:continue
   if re.search(r'_test.go:\d+:',e.get('Output','')):fl=e['Output'].strip();break
  row['evidence']='Probe only: ADAMIC_MUTANT=P04 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/P04 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^'+row['test']+'$ > '+path.name+' 2>&1; '+str(fl)
(E/'rows.json').write_text(json.dumps(rows,indent=2))
setup=json.load(open(E/'setup.json'));records=json.load(open(E/'matrix-time.json'));probes=json.load(open(E/'probe-time.json'));vet=[json.loads(l) for l in (E/'vet-times.jsonl').read_text().splitlines()];assert all(x['exit']==0 for x in vet);assert len(vet)==20
for f in ['lower.go','refusals.go','predicates.go','predicates_proof.go','expression.go']:assert (R/'internal/lower'/f).read_text()==base[f],f
assert not (R/'internal/lower/audit_selector.go').exists()
binaries=0;skips=[]
for line in (E/'baseline.log').read_text().splitlines():
 try:e=json.loads(line)
 except:continue
 if not e.get('Test') and e.get('Action')=='pass':baseline=e['Elapsed']
 if e.get('Action')=='skip' and e.get('Test'):skips.append(e['Test'])
for path in E.glob('timing-*.log'):
 for line in reversed(path.read_text().splitlines()):
  try:e=json.loads(line)
  except:continue
  if not e.get('Test') and e.get('Action')=='pass':binaries+=e['Elapsed'];break
matrixbinary=0
for path in E.glob('M??*.log'):
 if '-vet' in path.name:continue
 for line in reversed(path.read_text().splitlines()):
  try:e=json.loads(line)
  except:continue
  if not e.get('Test') and e.get('Action') in ['pass','fail'] and 'Elapsed' in e:matrixbinary+=e['Elapsed'];break
report={'start_commit':setup['start_commit'],'nproc':5,'setup':'skipped, warm env works','npm_seconds':setup['npm_seconds'],'baseline_binary_seconds':baseline,'baseline_command_seconds':json.load(open(E/'baseline-time.json'))['seconds'],'baseline_skips':skips,'timing_binary_seconds':binaries,'matrix_and_replay_command_seconds':sum(x['seconds'] for x in records),'matrix_and_replay_binary_seconds':matrixbinary,'probe_command_seconds':sum(x['seconds'] for x in probes),'standalone_vet_seconds':sum(x['seconds'] for x in vet),'clean_build_seconds':json.load(open(E/'clean-build-time.json'))['seconds'],'switch_build_seconds':'not separately timed; included in matrix command overhead','finished_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'verdict_counts':dict(collections.Counter(r['verdict'] for r in rows))}
(E/'totals.json').write_text(json.dumps(report,indent=2));(E/'validation.json').write_text(json.dumps({'standalone_diffs':20,'apply_check':'all passed against starting index','go_vet':'all twenty passed','production_restored':True},indent=2))
summary=['u042 audited at origin/main '+setup['start_commit']+'; all twelve requested names exist.','Verdicts: 4 sacred, 5 subsumed, 1 overlapping, 2 untrue under the fixed production plan.','Twenty mutations: nineteen killed; M18 survived with a changed helper output.','M01 is bounded after a Go panic; nineteen columns completed 239 tests; two rows accept empty answers.','Evidence on test-audit/internal-lower-predicates_overload under review/test-audit/internal-lower-predicates_overload/; production restored.']
text='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\nAll file locations are against the starting origin/main. M01 failed rows come from individual bounded reruns; kills outside them are unknown.\n\n| ID | File:line | Change | Failed rows |\n|---|---|---|---|\n'
for p in plans:
 change=(p['old']+' -> '+p['new']).replace('\n',' ').replace('|','\\|');text+='| '+p['id']+' | '+p['file']+':'+str(p['line'])+' | '+change+' | '+', '.join(matrix[p['id']])+' |\n'
text+='\nSurvivor M18: assignment helper directions 0 -> 3; entry directions 0 -> 0. witness-fixed-clean.log and witness-fixed-M18.log record actual outputs. This shows an unguarded helper result, not a demonstrated whole-checker or native miscompile.\n\n'+(E/'limitations.md').read_text()+'\nMeasured timings:\n\n```json\n'+json.dumps(report,indent=2)+'\n```\n'
(E/'report.md').write_text(text);shutil.copyfile('/tmp/u042-finalize.py',E/'scripts/u042-finalize.py')

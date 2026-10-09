import pathlib,json,collections,datetime,subprocess,shutil,re
R=pathlib.Path('/workspace/adamic');E=R/'review/test-audit/cmd-adamic-test262-corpus';rows=json.load(open(E/'rows.json'));plans=json.load(open(E/'plan.json'));matrix=json.load(open(E/'matrix.json'))
for r in rows:
 if r['verdict']!='helper':r['mutants_in_matrix']=17;r['admissible_mutants_in_matrix']=15
 if r['subsumer_seconds'] is not None:r['subsumer_seconds']=round(r['subsumer_seconds'],3)
 if r['verdict']=='subsumed':r['subsumption_mutants']=len(r['kills'])
 if r['test']=='TestNodeHarnessIdentity':
  for line in (E/'M13.log').read_text().splitlines():
   try:e=json.loads(line)
   except:continue
   if e.get('Test')=='TestNodeHarnessIdentity' and '_test.go:' in e.get('Output',''):
    fl=e['Output'].strip();r['last_proven_fail']='M13 (supplemental): '+fl;r['evidence']='ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M13 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M13.log 2>&1; '+fl;break
(E/'rows.json').write_text(json.dumps(rows,indent=2))
checks={};binary_total=0
for p in plans:
 events=[]
 for line in (E/(p['id']+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 terminal=[e for e in events if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ['pass','fail','skip']]
 elapsed=next(e['Elapsed'] for e in reversed(events) if not e.get('Test') and e.get('Action') in ['pass','fail'])
 checks[p['id']]={'top_level_terminal':len(terminal),'skipped':sum(e['Action']=='skip' for e in terminal),'binary_seconds':elapsed};assert len(terminal)==37;assert elapsed<90;binary_total+=elapsed
(E/'completion-checks.json').write_text(json.dumps(checks,indent=2))
timing_binary=0
for f in E.glob('timing-*.log'):
 for line in reversed(f.read_text().splitlines()):
  try:e=json.loads(line)
  except:continue
  if not e.get('Test') and e.get('Action') in ['pass','fail']:timing_binary+=e['Elapsed'];break
wall=sum(x['seconds'] for x in json.load(open(E/'matrix-time.json')));probewall=sum(x['seconds'] for x in json.load(open(E/'probe-time.json')))
totals={'setup':'warm environment worked, cloud/setup.sh skipped','npm_seconds':json.load(open(E/'setup.json'))['npm_seconds'],'nproc':5,'baseline_binary_seconds':56.101,'baseline_command_seconds':57.92519243300194,'timing_binary_seconds':timing_binary,'matrix_binary_seconds':binary_total,'matrix_command_seconds':wall,'matrix_frontend_and_build_overhead_seconds':wall-binary_total,'probe_command_seconds':probewall,'clean_build_seconds':json.load(open(E/'clean-build-time.json'))['seconds'],'standalone_vet_seconds':'not separately measured','finished_utc':datetime.datetime.now(datetime.timezone.utc).isoformat()}
(E/'totals.json').write_text(json.dumps(totals,indent=2))
summary=['u013 audited at origin/main 7b18d0576930caca4e22ce2eef92fcf563af52d0; all fifteen names exist.','Verdicts: 8 sacred, 5 subsumed, 1 cannot-judge, 1 helper.','Full matrix: 37 top-level tests per column; 15 admissible mutations plus 2 supplemental; no survivors or skips.','Three-run medians: 0.010 to 5.156 seconds; nproc=5; no scoped row passed its own empty-answer probe.','Evidence: test-audit/cmd-adamic-test262-corpus, review/test-audit/cmd-adamic-test262-corpus/; production restored.']
text='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\nM06 and M13 are supplemental and excluded from kills, unique kills and verdicts. Subsumption rests only on each row\'s listed admissible kills.\n\n| ID | File:line at starting origin/main | Change | Failed rows |\n|---|---|---|---|\n'
for p in plans:
 change=(p['old']+' -> '+p['new']).replace('\n',' ').replace('|','\\|')
 if p['id']=='M16':change='Drop the entire e.fallback.once.Do(func(){...}) statement'
 if p['id'] in ['M06','M13']:change='Supplemental: '+change
 text+='| '+p['id']+' | '+p['file']+':'+str(p['line'])+' | '+change+' | '+', '.join(matrix[p['id']])+' |\n'
text+='\nSurvivors: none. All fifteen admissible mutations and both supplemental mutations were caught. survivor-clean.log and survivor-M02.log are an extra behavior witness for the killed M02 mutation, not survivor claims.\n\n'+(E/'limitations.md').read_text()+'\nMeasured totals:\n\n```json\n'+json.dumps(totals,indent=2)+'\n```\n'
(E/'report.md').write_text(text)
for n in ['u013-report.py','u013-finalize.py']:
 if pathlib.Path('/tmp/'+n).exists():shutil.copyfile('/tmp/'+n,E/'scripts'/n)
shutil.copyfile('/tmp/u013-graph.go',E/'scripts/graph.go.fixture')
# immutable base source is redundant with origin/main; retain it as exact replay input

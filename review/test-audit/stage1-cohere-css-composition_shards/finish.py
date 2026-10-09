import pathlib,json,subprocess,shutil,re,time
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/stage1-cohere-css-composition_shards';rows=json.loads((p/'rows.json').read_text());runs=json.loads((p/'audit-runs.json').read_text())+json.loads((p/'followup-runs.json').read_text());groups=json.loads((p/'row-members.json').read_text());menu=json.loads((p/'menu.json').read_text());wanted=set(json.loads((p/'requested-tests.json').read_text()))
def events(f):
 out=[]
 for l in f.read_text().splitlines():
  try:out.append(json.loads(l))
  except:pass
 return out
checks=[]
for run in runs:
 if run['selector']=='clean' or '-json' not in run['command']:continue
 regex=run['command'][-1];requested=regex.removeprefix('^(').removesuffix(')$').removeprefix('^').removesuffix('$').split('|');ev=events(p/run['log']);term={e['Test']:e['Action'] for e in ev if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ['pass','fail','skip']}
 checks.append(dict(id=run['selector'],log=run['log'],command=run['command'],wall=run['wall'],requested=requested,results=term,unknown=[t for t in requested if t not in term]))
(p/'checks.json').write_text(json.dumps(checks,indent=2))
clean=set();skips=set()
for f in list(p.glob('baseline*.log'))+list(p.glob('timing*.log'))+[p/'bounded-printer-063-clean.log']:
 for e in events(f):
  if e.get('Test') in wanted and e.get('Action')=='pass':clean.add(e['Test'])
  if e.get('Test') in wanted and e.get('Action')=='skip':skips.add(e['Test'])
(p/'baseline-coverage.json').write_text(json.dumps(dict(passed=sorted(clean),unknown=sorted(wanted-clean-skips),skipped=sorted(skips)),indent=2))
apply=[]
for f in sorted(p.glob('*.diff')):
 q=subprocess.run(['git','apply','--check','--cached',str(f)],cwd=r,capture_output=True,text=True);apply.append(dict(diff=f.name,exit=q.returncode,output=q.stdout+q.stderr));assert q.returncode==0
(p/'diff-validation.json').write_text(json.dumps(apply,indent=2))
for row in rows:
 if row['test']=='TestCSSThroughput':row['vacuous_subcases']=['PMain native round 1 accepts empty output','PMain Node source round 1 accepts empty output'];row['limitations']+=' The second round rejects the empty answer after PostCSS establishes a nonempty checksum.'
 row['probe_evidence']=[]
 for c in checks:
  if not c['id'].startswith('P') or not set(c['requested'])&set(row['members']):continue
  errs=[e['Output'].strip() for e in events(p/c['log']) if e.get('OutputType')=='error' and e.get('Test','').split('/')[0] in row['members']]
  row['probe_evidence'].append(dict(id=c['id'],log=c['log'],command=c['command'],failing_output=errs[0] if errs else None,results=c['results']))
 row['matrix_rows']=['TestCompositionMatchesGo family','TestCSSThroughput','TestCSSPrinterOptimizedMatchesGo'] if row['mutants_in_matrix'] else [row['test']]
 if row['test']=='TestCSSPrinterAgreesWithGo family':row['limitations']+=' Full 64-member runtime exceeded 90 seconds twice. Witness proved on member 063 after a passing clean single-leaf run; agreement members and unfinished checks are not given production verdicts.'
(p/'rows.json').write_text(json.dumps(rows,indent=2))
for script in ['clean','timings','menu','audit','followup','report','finish']:
 f=pathlib.Path('/tmp/u078-'+script+'.py')
 if f.exists():shutil.copyfile(f,p/(script+'.py'))
cleanruns=json.loads((p/'clean-runs.json').read_text());timings=json.loads((p/'timing-runs.json').read_text());timingtotal=sum(x['wall'] for x in timings);cleantotal=sum(x['wall'] for x in cleanruns)
metadata=dict(base=(p/'base.txt').read_text().strip(),nproc=5,setup='warm env.sh, cloud/setup.sh skipped',dependency_install={'npm_ci_api_reported_seconds':0.421,'fixture_clone_wall':1.951,'postcss_install_wall':0.715,'prettier_install_wall':0.693},clean_controller_wall_sum=cleantotal,timing_controller_wall_sum=timingtotal,audit_and_followup_wall_sum=sum(x['wall'] for x in runs),native_mutant_build_wall={x['log'].split('-')[0]:x['wall'] for x in runs if re.match(r'M[1-4]-standalone-build',x['log'])},compiler_build_wall=next(x['wall'] for x in runs if x['log']=='compiler-build.log'),conditions='Go test -count=1 per invocation. Setup/composition timing families sequential; remaining singleton timing rows used two concurrent processes on 5 CPUs. Durations are each test binary package line, not command wall time. The third composition timing log completed while its controller was paused; its binary elapsed is valid, controller wall unavailable. Successful build validations were resumed rather than repeated after two invalid special edits. Total wall sums count the failed validation attempts.',overall_elapsed='Approximately 45 minutes before publishing; exceeded the suggested 30-minute port budget.')
(p/'timing-summary.json').write_text(json.dumps(metadata,indent=2))
# The complete required format is also preserved as a repository artifact.
summary=['u078: 95 listed tests present at base '+metadata['base'][:12]+', grouped into 16 rows.','Bounded verdicts: 2 sacred, 1 subsumed, 5 witnesses, 3 setup checks, 5 cannot-judge.','Four valid production mutants were caught; all standalone diffs apply to the starting commit.','Both composition setup wrappers pass the empty-setup probe; complete printer-family timing is over budget.','Evidence branch: test-audit/stage1-cohere-css-composition_shards; production sources restored.']
text='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n'
text+='| ID | Origin file:line | Change | Failed rows in bounded matrix |\n|---|---|---|---|\n'
mat=json.loads((p/'matrix.json').read_text())
for m in menu:text+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+m['change']+' | '+', '.join(x['row'] for x in mat if x['mutant']==m['id'] and x['failed'])+' |\n'
text+='\nSurvivors: none in the declared production matrix. M2 survives the optimized printer row but is caught by composition. Kills outside the declared rows remain unknown.\n\n'
text+='Special checks, separate from production kills: W1 disables firstDifference at css_test.go:277; WMemory disables native.go:103 sanitizer flags; WRange drops mismatch reports at nodes.ts:102 and tree.ts:179. SSetup erases the constructed oracle path at composition_shards_test.go:94; SMode changes count/8 to count/4 at printer_shards_test.go:293; SParser starts at variant 0 instead of -1 at parser_shards_test.go:77. Each fails its declared witness or construction row; checks.json contains commands, failed members and unknown members. None counts toward sacred/subsumed.\n\n'
text+='Brief problems, costs and limits:\n\n'
notes=[
'The brief says 14 rows but lists 95 Test functions. The shared-checker and distinct-assertion rules produce 16 rows here. row-members.json lists every member; no supplied name moved or vanished.',
'The supplied historical file commit 8de93800f4 is not current origin/main. All locations and diffs use the fetched base 60397548dd8a9494a7d2aaa874b7648b8e625607.',
'Whole-package clean testing cooked at the 90-second binary limit. No observed test assertion failed before that timeout. Production mutation testing therefore uses source-call bounds and runs each reached row separately.',
'The 64-member printer family cooked on both cold baseline and warm timing retry. Its full median is null, not an invented lower-bound median. A single clean member 063 followed by its weakened-check failure proves the bounded witness.',
'There are 26 requested printer leaves without a completed clean result. They are explicitly unknown in baseline-coverage.json. The family witness verdict does not claim production quality for those leaves or its agreement members.',
'The 48 printer mutant leaves share a checker with the agreement leaves. Their witness guard is conditional on ownership of a pinned witness; some other leaves check termination only. Whole-family witness status does not prove each leaf can detect disagreement.',
'The broad W1 printer run cooked after real assertion failures; unfinished members stay unknown in checks.json. The single-leaf replay removes reliance on the timeout as evidence.',
'TestCSSParserOptimizedMatchesNode executes the same port source on Node as its expected oracle. A port-source change would alter both implementation and oracle, so no such production kill is counted. PC probes only native C emission and leaves Node source intact.',
'The open-gap and three closed-gap rows check compiler or dependency-port snippets rather than the selected CSS port behaviors. No admissible production compiler/dependency mutant was built for them within the four port-mutant budget; their verdict is cannot-judge. PLower is only an empty-answer probe.',
'The maximum-four rebuild guidance and the aim of three mutants per row conflict for this mixed slice. Four production mutants cover parser, composition, canonical rendering and document printing. No verdict treats empty probes or guard edits as production mutants.',
'The function inventory is a conservative source declaration inventory, not dynamic TypeScript coverage. Exact transitive runtime reach was not proved. Entry/import reasoning defines the declared matrix bounds.',
'Throughput compares aggregate counts/checksum, not full trees. Its native-first expected value allows empty native and Node answers to pass round 1 under PMain; round 2 fails after PostCSS provides the nonempty checksum.',
'Both setup wrappers return successfully when compositionSetup returns its zero value at entry. Breaking construction instead makes their normal guard fail. They are setup-check rows with vacuous=true, not production-unique tests.',
'The first SSetup edit left oracle unused and failed go vet. It was replaced by oracle[:0], preserving variable use. The first WRange edit used constant false and lost TypeScript field narrowing; it was replaced by dropping the reporting statements. Invalid attempts are preserved and support no verdict.',
'Automatic approval review initially rejected a resume over a possible leftover source mutation. Read-only verification showed equal HEAD/origin/main hashes and no production diff; the subsequent safer resume was approved. No authorization remains blocked.',
'Timing medians are the package binary elapsed line from three separate count=1 runs. Some singleton rows ran concurrently with one other process, so their costs include contention on 5 CPUs. The printer family has no three successful full timings.',
'Warm tooling did not include CSS fixture or npm dependencies. API npm ci, pinned Prettier fixtures, PostCSS and Prettier installation were performed before baseline; throughput opt-in and both library comparison variables were enabled. No completed requested row skipped.',
'The WRange weakened run fails first on the raw native check; it does not independently prove that the later composed branch catches its disabled report. Its standalone diff compiles both modules, but only the observed raw failure supports the witness verdict.',
'No repo-wide replay, package uniqueness outside the declared bounds, full untimed CSS run, or extra compiler mutant campaign was performed. Central replay can use every standalone diff. No main push or PR was made.',
'The suggested 30-minute port budget was exceeded. Cold/warm 90-second family attempts, repeated native products, three-run timing requirements and special-edit validation consumed the extra time. All timed-out runs were stopped by their binary deadline rather than extended.'
]
text+='\n'.join('- '+n for n in notes)+'\n\n'
text+='Setup, building and running: '+json.dumps(metadata,indent=2)+'\n'
(p/'REPORT.md').write_text(text)
print(json.dumps(metadata,indent=2))

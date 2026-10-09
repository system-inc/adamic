import pathlib,json,statistics,re,subprocess,collections,gzip
p=pathlib.Path('review/test-audit/internal-oracle-private_generic_mutant');raw=json.loads((p/'scope.json').read_text());fm=['TestReadinessMutants','TestUninitializedIsNotNullishMutant','TestLazyInitializerIsNotEagerMutant'];family='TestReadinessMutants family';scope=[]
for r in raw:
 if r['test'] in fm:
  if r['test']!=fm[0]:continue
  r=dict(r,test=family,members=fm)
 else:r=dict(r,members=[r['test']])
 scope.append(r)
mapping={n:r['test'] for r in scope for n in r['members']};normal=['TestRuntimeLastIndexOfMatchesNode','TestScannerNestedReferences'];setup='TestRegexCycleFixtureHasItsNativeDependency';wrows=[r['test'] for r in scope if r['test'] not in normal+[setup]]
def events(id):
 file=p/(id+'.log');text=file.read_text() if file.exists() else gzip.open(str(file)+'.gz','rt').read()
 return [json.loads(l) for l in text.splitlines() if l.startswith('{')]
def failed(id):return sorted({mapping[e['Test'].split('/')[0]] for e in events(id) if e.get('Action')=='fail' and e.get('Test','').split('/')[0] in mapping})
def line(id,row):
 ev=events(id);outs=[e.get('Output','').strip() for e in ev if mapping.get(e.get('Test','').split('/')[0])==row and e.get('OutputType')=='error']
 if not outs:outs=[e.get('Output','').strip() for e in ev if mapping.get(e.get('Test','').split('/')[0])==row and 'panic:' in e.get('Output','')]
 return outs[0].splitlines()[0][:250] if outs else next((e.get('Output','').strip() for e in ev if e.get('Action')=='output' and '--- FAIL:' in e.get('Output','')),'no failure')
mat={m:{'failed_rows':failed(m),'matrix_rows':normal} for m in ['M01','M02','M03','M04']}
mat['W']={'failed_rows':failed('W'),'matrix_rows':wrows};mat['S01']={'failed_rows':failed('S01'),'matrix_rows':[setup]}
for q in ['P02','P03','P04']:mat[q]={'failed_rows':failed(q),'matrix_rows':normal if q!='P04' else normal[:1]}
mat['P01']={'failed_rows':sorted(set(failed('P01-1')+failed('P01-2'))),'matrix_rows':normal,'rerun':'Each semantic row alone, because nil Lower panics'}
assert len(mat['W']['failed_rows'])==10
oracle={family:('Self eager non-null panic contract and harmless-output detector; source Node only sanity-checked, own JavaScript backend agreement.','self'),setup:('Self reflection on ir.Program.Regexps. Body has no failure assertion; false dependency result only skips.','self'),'TestNonNullWeakFreedNamesExpression':('Live source Node for pinned tracing output; self pinned native freed-Weak diagnostic.',['external-run','self']),'TestRegExpReplacementTypeGuardMutants':('Self declared-type guard policy, using our JavaScript backend output as native expectation.','self')}
report=[]
for r in scope:
 row=r['test'];members=r['members'];times=json.loads((p/('Readiness-family-times.json' if row==family else row+'-times.json')).read_text());assert len(times)==3 and all(t['exit']==0 for t in times);seconds=statistics.median(t['seconds'] for t in times)
 if row in normal:
  kills=[m for m in ['M01','M02','M03','M04'] if row in mat[m]['failed_rows']];unique=[m for m in kills if len(mat[m]['failed_rows'])==1];verdict='sacred' if unique else 'untrue';lastid=kills[-1];last=line(lastid,row);command=json.loads((p/(lastid+'-run.json')).read_text())['command'];mr=normal;qs=['P01','P02','P03']+(['P04'] if row==normal[0] else []);pk=[q for q in qs if row in mat[q]['failed_rows']];vacuous=False if pk else None;mi=4
 elif row==setup:
  kills=[];unique=[];verdict='untrue';lastid=None;last=None;command=json.loads((p/'S01-run.json').read_text())['command'];mr=[setup];qs=[];pk=[];vacuous=None;mi=0
 else:
  kills=[];unique=[];verdict='witness';lastid='W';last=line('W',row);command=json.loads((p/'W-run.json').read_text())['command'];mr=wrows;qs=[];pk=[];vacuous=None;mi=0
 default=('Live source Node compared with deliberately changed native or JavaScript output.','external-run')
 if row in normal:default=('Live source Node stdout, stderr and exit compared with sanitized, release and our JavaScript backend; leak checking.','external-run')
 ora,kind=oracle.get(row,default)
 result={'test':row,'package':'internal/oracle','file':r['file'],'seconds':seconds,'oracle':ora,'oracle_kind':kind,'kills':kills,'unique_kills':unique,'last_proven_fail':lastid+': '+last if last else None,'verdict':verdict,'subsumed_by':[],'mutants_in_matrix':mi,'probe_kills':pk,'subsumer_seconds':None,'vacuous':vacuous,'bounded':True,'matrix_rows':mr,'evidence':command+' => '+(last if last else 'SKIP: native regex lowering/emission is on codex/stage1-css; run this fixture on the scratch merge'),'members':members}
 report.append(result)
(p/'rows.json').write_text(json.dumps(report,indent=2));(p/'matrix.json').write_text(json.dumps(mat,indent=2));(p/'grouped-scope.json').write_text(json.dumps(scope,indent=2))
plan=json.loads((p/'mutant-plan.json').read_text());table=['| ID | Origin/main file:line | Change | Failed rows |','|---|---|---|---|']
for m in plan:table.append('| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+m['before']+' → '+m['after']+' | '+', '.join(mat[m['id']]['failed_rows'])+' |')
(p/'mutant-table.md').write_text('\n'.join(table)+'\n')
for src,dst in [('/tmp/u068-baseline.log','package-baseline.log'),('/tmp/u068-list.log','test-list.log'),('/tmp/u068-npm.log','npm.log')]: (p/dst).write_bytes(pathlib.Path(src).read_bytes())
skips=[e['Test'] for e in events('package-baseline') if e.get('Action')=='skip'];(p/'baseline-skips.json').write_text(json.dumps(skips,indent=2))
rebuilds=[]
for id in ['CONTROL','M01','M02','M03','M04','P02','P03','P01-1','P01-2','P04']:
 ev=events(id);end=next(e for e in reversed(ev) if e.get('Action') in ['pass','fail'] and 'Test' not in e);run=json.loads((p/(id+'-run.json')).read_text());rebuilds.append({'id':id,'wall':run['wall'],'binary_seconds':end['Elapsed'],'native_rebuild_seconds':None,'note':'Native Build and execution occur inside binary time; not separately instrumented.'})
(p/'build-and-run-times.json').write_text(json.dumps(rebuilds,indent=2))
summary={'unit':'u068','commit':'6c60da091afddc9c2fe88b3a1067845b6dc79cb3','nproc':5,'raw_functions':15,'grouped_rows':13,'moved_or_missing':[],'setup_seconds':0,'npm_ci_reported_seconds':0.389,'full_baseline_binary_seconds':90.108,'slice_baseline_binary_seconds':2.945,'coverage_binary_seconds':4.208,'standalone_validation_wall':json.loads((p/'standalone-validation.json').read_text())['wall'],'verdicts':dict(collections.Counter(r['verdict'] for r in report)),'survivors':[],'native_rebuild_timing':'Included in binary timings, not separately measured.'}
(p/'summary.json').write_text(json.dumps(summary,indent=2))
intro='''u068 started from origin/main 6c60da091afddc9c2fe88b3a1067845b6dc79cb3, nproc 5.
All 15 names exist in their listed files; the shared readiness checker groups three names into one family, yielding 13 rows.
Full clean baseline timed out at 90.108 seconds; the selected clean slice passed in 2.945 seconds, with no selected skips.
Four production mutants support 2 bounded sacred rows; weakened checks prove 10 witness rows; the dependency construction row is untrue.
Four empty-entry probes were caught; no production mutant survived; standalone evidence is on test-audit/internal-oracle-private_generic_mutant.
'''
friction='''The reference commit 8de93800f4 is older than the fetched starting commit. None of the fifteen names moved or vanished. The brief calls them fifteen rows, but three readiness wrappers differ only by inputs to assertMigratedNonNullCheck and must be grouped. This family has three additional count=1 timing runs; individual member timing logs are retained too. There is no common name prefix, so the family takes its first member's name.

Twelve original functions are built-in-mutant witnesses. Counting their production failures as kills would be wrong. The Readiness helper now tests one removed non-null check per input, despite historical subcase names such as initialize-to-zero or miss-exception-path. The probe struct's stdout and stderr fields are not used by its body. The expected eager panic text is self-written policy. The replacement type-guard witness takes its expected result from our JavaScript backend, so its expected answer is self, not an independent Node source oracle.

The regex dependency row has no assertion after its skip guard. S01 makes nativeRegexSlicePresent return false; the row skips and the binary exits zero. Therefore the construction break is not caught. The skip text still says native regex is on codex/stage1-css even though the clean starting commit has the Regexps representation and the row passes. This is an observed untrue setup row, not a failed semantic test.

The first npm/list/baseline shell was launched from stage3/api. npm ci succeeded there, but relative Go package paths did not resolve. Those commands never executed tests. They were rerun from the repository root before any audit. The 90 second full-package timeout was not an individual failure. The matrix was narrowed to the two ordinary semantic rows in this slice; all witness and construction checks were run separately. Other package rows, including the general fixture family, were not replayed. Every unique kill is bounded to this two-row production matrix. Central replay must settle package and repository uniqueness.

Full baseline skips are listed in baseline-skips.json. They are WASI opt-ins outside the requested slice; they were not enabled or installed for this audit. All selected rows ran without skips in the clean baseline, including the dependency check. Node dependencies were refreshed with npm ci. The warm toolchain worked, so setup.sh was skipped.

reached-functions.txt lists 386 measured Go functions across oracle, lower, native and javascript. These measurements include oracle cache/harness production files; harness helpers defined in _test.go are not instrumented by Go coverage. All nine requested test files and the migrated-check helper were read whole. Runtime C definitions are conservatively listed in runtime-functions-static-superset.json. Dynamic C reachability was not instrumented because llvm-cov and llvm-profdata were unavailable. Thus the exact exhaustive dynamic function inventory remains a limitation; mutants themselves come from read, reached compiler/runtime code, not oracle or fixture expectations.

The fixed menu was written in mutant-plan.json before production outcomes. M01/M02 cover last-index constants and bounds; M03 reverses the positive frame-identity condition; M04 flips canonical-closure cache construction. No supplemental inserted-statement mutant contributes to a verdict. Selector instrumentation is isolated in switch.diff. Every production/probe standalone diff applies to the starting sources. Go changes passed go vet against original-source overlays; runtime changes compiled with native C11 warnings, count and sanitizer flags. The Lower empty probe drops the whole body and removes its now-unused fmt/filepath imports. Witness edits only weaken comparison checks; they pass go vet separately. Production files were restored before commit.

Compiler and runtime selectors use separate ADAMIC_BUILD_CACHE_DIR paths and ADAMIC_GATE_UNCACHED=1. This second setting matters: cached oracle observations can otherwise conceal a runtime environment selector even when the compiled bytes are identical. The switched clean control passed in 19.812 binary seconds. Production mutants completed in 0.666, 0.593, 0.677 and 0.604 binary seconds. The Lower nil probe panics, so each ordinary row was run alone; only observed row failures are recorded, and untouched subcases remain unknown. Witness vacuity is null because production empty-answer probes do not judge their check.

Setup.sh time was zero; npm reported 389ms. The full baseline cost 90.108 binary seconds, slice baseline 2.945, measured Go coverage 4.208. All grouped rows have three separate count=1 timing runs. build-and-run-times.json records each matrix/probe wall and binary time. Individual native rebuild time was not separately instrumented and remains included in binary elapsed time. The fixed switched runtime was reused; the per-selector build-cache roots and uncached oracle results avoided stale answers. Total session time was about 19 minutes including reading, reporting and push. No selected row exceeded 60 seconds alone.

Not covered: production kills outside the two ordinary rows, repository-wide uniqueness, exact dynamic C/test-helper coverage, separately measured native rebuild times, and unrelated WASI opt-ins. No survivor needs an equivalence witness. No PR, main push, production source change, or deletion recommendation is included.
'''
(p/'friction-and-limits.md').write_text(friction);(p/'REPORT.md').write_text(intro+'\n```json\n'+json.dumps(report,indent=2)+'\n```\n\n'+(p/'mutant-table.md').read_text()+'\nSurvivors: none. Every production mutant was caught by a semantic row.\n\n'+friction)
for r in report:print(r['test'],r['seconds'],r['verdict'],r['last_proven_fail'])

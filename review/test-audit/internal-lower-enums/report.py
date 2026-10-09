from pathlib import Path
import json,re,statistics,subprocess
out=Path('review/test-audit/internal-lower-enums'); rows=out.joinpath('rows.txt').read_text().splitlines(); menu=json.loads(out.joinpath('menu.json').read_text()); ids=[m['id'] for m in menu]+['PLower','POrder','PProve']
def events(path):
 e=[]
 for s in path.read_text().splitlines():
  try:e.append(json.loads(s))
  except ValueError:pass
 return e
matrix={}; failures={}; subpasses={}; elapsed={}
for mid in ids:
 e=events(out/(mid+'.log')); matrix[mid]={}
 for r in rows:
  use=events(out/(mid+'-'+r+'.log')) if out.joinpath(mid+'-'+r+'.log').exists() else e
  status=[x['Action'] for x in use if x.get('Test')==r and x['Action'] in ['pass','fail','skip']]
  matrix[mid][r]=status[-1] if status else 'unknown'
  fail=[x.get('Output','').strip() for x in use if x.get('Test','').split('/')[0]==r and x['Action']=='output' and re.search(r'_test.go:\d+:',x.get('Output','')) and 'statements=map' not in x.get('Output','')]
  panic=[x['Output'].strip() for x in use if x.get('Output','').startswith('panic: ')]
  if status and status[-1]=='fail':failures[mid,r]=fail[0] if fail else panic[0] if panic else '--- FAIL: '+r
  subpasses[mid,r]=[x['Test'].split('/',1)[1] for x in use if x['Action']=='pass' and x.get('Test','').startswith(r+'/')]
 elapsed[mid]=[x.get('Elapsed') for x in e if x['Action'] in ['pass','fail'] and 'Test' not in x][-1]
seconds={};rounds={}
for r in rows:
 rounds[r]=[float(re.search(r'\t([0-9.]+)s',out.joinpath(f'{r}-{i}.log').read_text())[1]) for i in range(1,4)];seconds[r]=statistics.median(rounds[r])
locations={}
for p in Path('internal/lower').glob('*_test.go'):
 for match in re.finditer(r'^func (Test\w+)\(',p.read_text(),re.M):
  if match[1] in rows:locations[match[1]]=str(p)+':'+str(p.read_text()[:match.start()].count('\n')+1)
oracles=[
 'Self-written accepted/refused slot views and broad reason substrings; positive cases only require err=nil.',
 'Self-written Refused type and non-exhaustive enum switch substring; accepted neighbors only require err=nil.',
 'Self-written NotYet or Refused type, without checking the specific reason for most inputs.',
 'Self-written absence of enum binding and ObjectLiteral nodes. POrder yields an empty IR program and this row passes, showing its absence-only oracle weakness; primary PLower instead panics.',
 'Self-written Refused type and enum-members diagnostic substring after loader admission.',
 'Self-written exact cycle-refusal fragments, location and Weak fix; accepted neighbors require err=nil.',
 'Self-written NotYet type only; no diagnostic identity or behavior comparison.',
 'Self-written Refused type and polymorphic recursion substring.',
 'Self-written count of two functions with describe_ prefix; no signature/type-identity comparison.',
 'Self-written exact named JSON union array NotYet.What.',
 'git verifies upstream source pin; Go cohere findings must be nonempty. Module order compares to recorded independently verified Node order (stage3/fixtures/cycles/order.json), not a live Node run here; no order value independently rechecked this session. Self-written 914/58 totals and binding identities. Proof decision counts are exported without assertion: PProve passes with 489 checked/425 proven reads instead of 9/905.',
 'Original source Node, generated-JavaScript Node and sanitized native each compared to expected output/exit; self-written emitted-C absence assertions also check proof elision.',
 'Self-written NotYet type and diagnostic naming each prelude door.',
 'Self-written fixture count 47, Lower NotYet name/location or local NotYet; copied TypeScript array-spread diagnostic text is also matched against the running checker. The seven _array cases never call Lower and pass PLower.'
]
kinds=['self']*14;kinds[10]=['external-run','external-authority','self'];kinds[11]=['external-run','self'];kinds[13]=['external-authority','self']
result=[]
for n,r in enumerate(rows):
 kills=[mid for mid in ids if mid.startswith('M') and matrix[mid][r]=='fail'];unique=[mid for mid in kills if sum(v=='fail' for v in matrix[mid].values())==1]
 covers=[other for other in rows if other!=r and kills and all(matrix[mid][other]=='fail' for mid in kills)]
 if unique:verdict='sacred'; subs=[]
 elif covers:verdict='subsumed';subs=[min(covers,key=lambda other:seconds[other])]
 elif kills:verdict='overlapping';subs=sorted({other for mid in kills for other in rows if other!=r and matrix[mid][other]=='fail'})
 else:verdict='untrue';subs=[]
 last=unique[-1] if unique else kills[-1] if kills else None
 obj=dict(test=r,package='internal/lower',file=locations[r],seconds=seconds[r],oracle=oracles[n],oracle_kind=kinds[n],kills=kills,unique_kills=unique,last_proven_fail=last+' '+failures[last,r] if last else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=20,probe_kills=[mid for mid in ['PLower','POrder','PProve'] if matrix[mid][r]=='fail'],subsumer_seconds=seconds[subs[0]] if verdict=='subsumed' else None,vacuous=False,bounded=True,matrix_rows=rows,evidence='ADAMIC_MUTANT='+last+' ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/'+last+' timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run "$pattern" > '+last+'.log 2>&1; '+failures[last,r] if last else 'No kill observed')
 if r in ['TestEnumSlotViews','TestInputSpreadCoverage']:obj['vacuous_subcases']=subpasses['PLower',r]
 if r=='TestOriginalCycleLedger':obj.update(entry_probe_results={'esmModuleOrder':'POrder: fail','proveModuleReads':'PProve: pass'},vacuous_subcases=['proveModuleReads proof decisions: empty proof passes PProve; module order still fails POrder'])
 result.append(obj)
assert all(v in ['pass','fail'] for mid in ids for v in matrix[mid].values())
assert all(matrix[m['id']][r]=='fail' for m in menu for r in rows if m['id'] in next(x for x in result if x['test']==r)['kills'])
out.joinpath('rows.json').write_text(json.dumps(result,indent=2)+'\n');out.joinpath('matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');out.joinpath('matrix.csv').write_text('id,'+','.join(rows)+'\n'+''.join(mid+','+','.join(matrix[mid][r] for r in rows)+'\n' for mid in ids));out.joinpath('timings.json').write_text(json.dumps(dict(nproc=5,rounds=rounds,medians=seconds,matrix_binary_elapsed=elapsed),indent=2)+'\n')
for name in ['baseline','PProve']:
 p=Path('/tmp/u032/ledger-'+name+'.json');out.joinpath('ledger-'+name+'.json').write_bytes(p.read_bytes())
survivors=[m['id'] for m in menu if not any(v=='fail' for v in matrix[m['id']].values())];assert not survivors
sumwall=lambda name:sum((int(x.split()[-1])-int(x.split()[-2]))/1e9 for x in out.joinpath(name).read_text().splitlines())
wallmatrix=sumwall('matrix-times.txt');wallclean=sumwall('times.txt');wallvet=sum(float(x.split()[-1]) for x in out.joinpath('vet-times.txt').read_text().splitlines());wallrecovery=sum(float(x.split()[-1]) for x in out.joinpath('recovery-times.txt').read_text().splitlines())
summary='''u032: 14 discovered rows; 9 sacred and 5 subsumed within the bounded matrix.
Base: 8171b3173bdbfce1f7982d3c4f731279307ece37; clean package passed in 55.418s.
20 fixed-menu mutants, three entry probes; no production survivor or skipped unit row.
Ledger proof decisions pass an empty proof; module-order obligations remain checked.
Evidence pushed on test-audit/internal-lower-enums under review/test-audit/internal-lower-enums/.
'''
text=summary+'\n```json\n'+json.dumps(result,indent=2)+'\n```\n\n'+out.joinpath('code-and-oracles.md').read_text()+'\n'
text+='All file:line references use the base commit. The complete pre-mutation reached-function list is reached-functions.txt (273 declarations), backed by coverage.out and functions-coverage.txt. Every requested name exists in list.log at the starting commit and remains in the file named by the brief; none moved or vanished. No families, witnesses, setup-only rows or subprocess helpers were found after reading all seven complete test files.\n\n'
text+='Matrix pattern is `pattern="^($(paste -sd \'|\' review/test-audit/internal-lower-enums/rows.txt))$"`. Source env.sh in every shell. Set ADAMIC_CYCLE_LEDGER_ROOT=/tmp/u032/typescript and ADAMIC_CYCLE_LEDGER_OUTPUT=/tmp/u032/ledger-<id>.json. Each matrix command sets ADAMIC_MUTANT=<id> and ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/<id>, then runs `timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run "$pattern" > <id>.log 2>&1`. Paths in JSON evidence are shortened relative to the evidence directory. Timings run `timeout 95 go test -count=1 -timeout 90s ./internal/lower/ -run ^<row>$ > <row>-<round>.log 2>&1` three separate times, with no concurrent test workload. No native product shares a cache directory across mutants.\n\n'
text+='| ID | Origin file:line | Change | Failed rows |\n|---|---|---|---|\n'
for m in menu:
 text+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+m['menu']+' | '+', '.join(r for r in rows if matrix[m['id']][r]=='fail')+' |\n'
text+='\nNo production survivors. Every production mutant uses the fixed menu; no supplemental mutant is needed. Diagnostic changes are production output changes, not oracle edits. M08 drops the entire representative append statement. M02/M03 standalone diffs replace their full check bodies with early returns. Probes replace full bodies and remove unused imports for standalone compilation. Switch scaffolding is stored under switch/*.txt and is absent from every standalone diff.\n\n'
text+='PLower returns nil, nil at Lower entry; POrder returns nil, false from esmModuleOrder; PProve returns immediately from proveModuleReads. PLower panicked during the whole slice, so every row was rerun alone for it, in PLower-<row>.log. Only those isolated observations establish its results; ledger passes because it does not call Lower. Thirteen Lower rows fail, with nil-pointer panics in const-enum, generic-instance and readiness rows. For the ledger, POrder fails but PProve passes. This multi-entry row is overall non-vacuous because it rejects empty scheduling; its proof-decision portion is explicitly vacuous. PProve changes read counts from checked=9/proven=905 to checked=489/proven=425, and statement counts from checked=3/proven=55 to checked=29/proven=29, without a failure. Both complete ledger artifacts are saved.\n\n'
text+='Under PLower, all 17 open enum-slot positive subcases pass while six closed/refusal subcases fail; exact names are in rows.json. All seven InputSpreadCoverage _array.a subcases pass because they stop at the loader checker before Lower, while lowering subcases fail. ConstEnumErasesRuntimeObject also passes POrder because its checks assert only absence of a runtime binding/object in an empty IR program; this supplemental entry is not used to set primary-Lower vacuity. No test or oracle was edited.\n\n'
text+='Subsumption rests only on this fixed bounded menu: ConstEnumErasesRuntimeObject by EnumSwitchExhaustiveness (2 kills); FunctionValueUnionViewsStayNotYet by EnumSlotViews (1 kill); GenericUnionFixtureHasSeparateInstances by GenericJSONUnionArrayIsNotYet (2 kills); the two input-spread rows mutually subsume each other (2 kills each). These are hints, not deletion recommendations. All unique_kills mean unique within these fourteen rows, not the package or repository.\n\n'
text+='Brief issues and time costs:\n\n- origin/main is newer than the brief\'s 8de93800f4 file-location pin. The fetched starting commit and actual discovery define scope; all fourteen names and file locations still match.\n- Warm env.sh does not imply warm Go dependencies. Cold test discovery ran about 102 seconds before I stopped it; a bounded 90-second retry also cooked, and the next bounded retry succeeded. This was compilation, not a red test baseline. The first discovery was insufficiently bounded, a process-control error documented here. Completed dependency compilation was reused; no reduced row selection could avoid compiling this common package dependency.\n- The ledger opt-in was installable, so pristine upstream v6.0.3 was cloned at 050880ce59e30b356b686bd3144efe24f875ebc8 and processDiagnosticMessages.mjs generated diagnostics. All fourteen rows then ran with no skip. No npm install was required in this source-only upstream directory; stage3/api npm ci reported 548ms.\n- Whole-package baseline fits 90 seconds but costs 55.418s. Repeating it for twenty mutants would exceed the unit budget, so the matrix is the explicitly requested fourteen-row slice and bounded. Outside-slice kills are unknown. This uses the brief\'s slice exception rather than claiming package uniqueness.\n- The ledger has multiple lowering entries. The singular entry-probe rule is ambiguous here. Both were probed; overall vacuous=false records failure on empty order, and vacuous_subcases records acceptance of empty proof.\n- Nil Lower output aborts the test binary; all fourteen isolated recovery runs were necessary. Panics are observed failures, not inferred failures for unseen rows.\n- InputSpreadCoverage has seven preparation-only checker subcases. Mutating Load or TypeScript would violate the declared lowering scope; those cases are reported as passing the lowering probe.\n- The absence-only const-enum oracle passes an empty IR through POrder. This does not make it sacred or establish primary-entry vacuity; it is an additional demonstrated weakness.\n- The first npm redirect used a relative evidence directory from stage3/api and failed before installation. It was corrected to the absolute root evidence path.\n- Logs are ignored by repository policy and were force-added only under the authorized evidence directory.\n\n'
text+=f'Timings: nproc=5; Go 1.27.1, Node 24.19.0. Full toolchain setup skipped because env.sh worked. Exact total setup and cold compile-only durations were not instrumented, so cannot precisely report them; cooked discovery observations are above. Clean coverage slice wall: 18.921s. Clean coverage plus 42 isolated timing commands: {wallclean:.3f}s wall total. Matrix/probes: {wallmatrix:.3f}s wall total, {sum(elapsed.values()):.3f}s reported binary elapsed total. First switched matrix command: {(int(out.joinpath("matrix-times.txt").read_text().splitlines()[0].split()[3])-int(out.joinpath("matrix-times.txt").read_text().splitlines()[0].split()[2]))/1e9:.3f}s including Go rebuild, native builds and tests; subsequent command walls are in matrix-times.txt. Panic recovery: {wallrecovery:.3f}s wall. All 23 standalone diff vet checks: {wallvet:.3f}s wall total, each exit zero. Native build-only times were not separately instrumented; per-subcase elapsed times include lowering, builds and executions and are in each JSON log. Three-round binary medians are in rows.json/timings.json. No unit row exceeds 60s.\n\n'
text+='Validation: every standalone diff applies to the exact base and compiles with go vet ./internal/lower/; production source is restored exactly, and restored.log records a clean passing slice. No main push or PR. Not covered: kills outside the fourteen rows, repo-wide uniqueness, exhaustive mutation coverage, external validation of the recorded Node order, live native tsc for the ledger, or isolated native compile-only timings. The full clean package skipped TestOptionalWideningCensus (outside this unit, opt-in source/corpus root) and one explicitly pending MixedUnionContractGraph subcase; no requested row skipped.\n'
out.joinpath('REPORT.md').write_text(text)
print('verdicts',[(x['test'],x['verdict'],x['kills'],x['unique_kills'],x['subsumed_by']) for x in result]);print('walls',wallclean,wallmatrix,wallrecovery,wallvet)

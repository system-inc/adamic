import pathlib,json,datetime
p=pathlib.Path('review/test-defend/internal-native-library')
mat=json.loads((p/'matrix.json').read_text()); by={r['mutant']:r for r in mat}
targets=[('TestRuntimeCacheConcurrentProcesses','subsumed',['TestRuntimeCacheKeepsCountFlags'],['D1']),('TestNbodyBorrowedLoopC','subsumed',['TestLoopBorrowPlan'],['D2']),('TestMapHashProbeBound','untrue',[],['D3','D4'])]
fail={'D1':'library_test.go:250: process 0: exit status 1','D2':'loop_borrow_test.go:79: function 1 offsetMomentum still counts binding adamic_local_18_body','D4':'map_hash_test.go:91: probe bound 64 exceeded: integers entries=1024 hit=743 miss=764'}
results=[]
for name,prior,subs,ids in targets:
 unique=ids[-1];r=by[unique]
 results.append(dict(test=name,package='internal/native',prior_verdict=prior,subsumed_by=subs,defense='defended',unique_mutant=unique+' '+r['file_line'],attempts=[{k:by[mid][k] for k in ['mutant','file_line','change','rows_failed']} for mid in ids],evidence='source /workspace/adamic-tools/env.sh; python3 '+str(p/'run-defense.py')+(' D4' if unique=='D4' else '')+'; '+unique+'.log: '+fail[unique]+'; exact timeout/go argv in matrix.json',bounded=True,matrix_rows=r['matrix_rows'],passed_tests=r['rows_passed'],skipped_tests=r['rows_skipped'],unique_scope='Completed bounded matrix only. Omitted package tests remain unknown.'))
(p/'results.json').write_text(json.dumps(results,indent=2)+'\n')
allrows=[l for l in (p/'test-list.log').read_text().splitlines() if l.startswith('Test')]
selected=json.loads((p/'matrix-rows.json').read_text()); omitted=sorted(set(allrows)-set(selected))
(p/'omitted-tests.json').write_text(json.dumps(omitted,indent=2)+'\n')
def info(name):
 es=[]
 for l in (p/name).read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 return dict(package_seconds=[e.get('Elapsed') for e in es if e['Action'] in ['pass','fail'] and 'Test' not in e],fails=[e.get('Test') for e in es if e['Action']=='fail'],skips=[e.get('Test') for e in es if e['Action']=='skip'])
logs=['baseline-package.log','baseline-slice.log','rest-matrix.coverage.log','restored-targets.log']+[n+'.coverage.log' for n in ['TestRuntimeCacheConcurrentProcesses','TestRuntimeCacheKeepsCountFlags','TestNbodyBorrowedLoopC','TestLoopBorrowPlan','TestMapHashProbeBound','TestMapHashProbeCatchesMutants','TestLibraryMapSetIteratorResources']]
timing={n:info(n) for n in logs if (p/n).exists()};(p/'checks.json').write_text(json.dumps(timing,indent=2)+'\n')
summary='''All three requested rows have unique catches within completed bounded matrices; retain them.
Four production mutants compiled; D1, D2 and D4 each failed only its intended row in its tested set.
The full clean package exceeded 90 seconds; omitted package tests and the opt-in measurement row remain unknown.

Starting origin/main: {base}. nproc=5. Go 1.27.1, Node 24.19.0, clang 20.1.8. Warm env.sh worked; setup skipped. npm ci completed, reported 372 ms. Sources and tests are restored, and only evidence is committed.

CODE UNDER TEST AND ORACLES were named before mutations in plan.md. All expectations are self-authored: successful competing cache publications with the correct header, executable answer 42 and one final directory; borrowed nbody locals plus absence of generated retain/release for loop bindings; actual C SameValueZero, map lookups, and maximum probe length 64. No oracle, harness, test or input was changed. TestLibraryMapSetIteratorResources is also a self-authored resource gate, including an exact computed total and 0.35 CPU-second lookup bound.

Coverage commands and raw profiles are retained. ConcurrentProcesses has zero exclusive Go blocks against KeepsCountFlags. Its children use independent processes and reach the competing publication recovery branch; parent-process coverage does not record their execution. The actual failing child messages show the real rename conflict, so the defense is semantic concurrency evidence. Its normal subsumer builds in one process and reuses cached runtimes.

NbodyBorrowedLoopC has 671 exclusive covered Go blocks against LoopBorrowPlan, including forOf binding emission at emit_statements.go:529-531. Those blocks mostly follow from generating C, which the plan-only subsumer does not do. D2 drops that fast path while leaving borrow planning and iterator ownership unchanged. It adds balanced retain/release traffic: this is an answer-preserving cost regression. Nbody's binding assertion fails while TestLoopBorrowPlan, TestLoopArrayHoldC, TestLoopCallCoverage and TestNbodyIndexedElementsBorrow pass.

Go coverage does not instrument C. MapHashProbeBound has zero exclusive Go blocks against the combined other 37 bounded rows, one against its witness and 21 against the resource row. Those Go blocks do not establish exclusive hash behavior. The semantic defense is the hash distribution budget: D3's 256-home mask failed both Bound and the newer resource speed gate (1.507457 CPU seconds). D4 allows 1024 homes. At 1024 integer entries Bound measures hit=743 and miss=764. The resource row's 1000-key lookup result remains correct and takes 0.067062 CPU seconds, below its 0.35 bound. Both malloc/slabs Bound subcases fail; the independent built-in-mutant witness passes. All non-resource assertion rows in that completed matrix pass.

matrix.json contains every exact timeout/go argv, cache directory convention, validation command, timing, failed and passed test list. D1 runs eight rows; D2 runs 36; D3 and D4 run 38. Every matrix completed below 90 seconds. TestMeasureClangUnits skipped: it needs ADAMIC_CLANG_MEASURE with pre-existing emitted original/changed C products and a checker archive; it is a separate optional compiler measurement workflow, and no measurement products were generated here. The raw discovery list has {count} current rows; omitted-tests.json lists {omitted} omitted rows. They are unknown, not passes. TestRuntimeKeyKeepsBoundaries and TestLibraryMapSetIteratorResources, absent from the old audit matrix, were included.

Each standalone diff applies to the starting origin/main snapshot. Go mutants passed go vet ./internal/native/. Both C masks passed standalone clang compilation with the runtime warning, optimization and sanitizer flags. D2's matrix also compiled and ran generated native C in its ordinary integration rows. ADAMIC_BUILD_CACHE_DIR=/tmp/defend-native-library/cache/<id> was set separately per mutant. Native runtime archives actually use a source/header/flags/compiler-content key under the user cache directory, so C source changes rebuild the archive; ADAMIC_BUILD_CACHE_DIR alone is not that cache key. No cold XDG_CACHE_HOME/GOCACHE relocation was used.

Brief friction and limits:
1. /tmp is an 8.8 GB filesystem, so 15 GB free is impossible. It had 6.3 GB free initially and about 6.2 GB during the runs. Earlier named defender and CSS scratch directories were removed; repository and tools were untouched. No full-disk failure occurred.
2. Recursive audit fetch stalled while fetching optional nested TypeScript submodule history. The parent audit ref was already fetched; the optional child was stopped and the full-refspec fetch completed with --no-recurse-submodules. No submodule content was changed.
3. The number unit was interrupted by this new assignment before any mutants. Its untracked logs were left out of this commit. An environment restart briefly made execution tools unavailable; they returned on the next environment update.
4. The full package baseline cooked at 90.058 seconds with no individual failed-test event. Clean target coverage runs and the clean cache/loop baseline passed before mutations. The restored other-row coverage matrix and restored targets passed. A complete package-wide uniqueness proof was not obtained; all defended labels here are explicitly bounded leads, not claims about omitted rows.
5. Go coverage cannot observe C bodies or automatically merge the subprocess cache branch into the parent profile. Exclusive lines alone could not decide two defenses. Commands, byte/probe evidence and semantic input/work histories decide them instead.
6. The original map audit tried one weak changed hash and called Bound untrue within that set. D3 shows the new speed gate catches broad hash degradation; D4 shows the probe gate checks a stricter, distinct distribution property.
7. A diagnostic logged by the built-in-mutant witness is a successful witness assertion, not a production-mutant kill. Matrix construction uses top-level fail events, preventing those diagnostics from becoming false catches.
8. Native archive compilation occurs inside test-binary time and was not separately measured. Validation seconds and complete command wall seconds are recorded separately in matrix.json. Clean and mutant package binary durations are in checks.json and raw logs.
9. No requested row remains undefended in the bounded evidence. Nbody's name describes its C ownership assertion; it does not promise or assert elapsed runtime speed, and it has no required-function/loop inventory guard. The map bound is a generous regression gate for its fixed patterns, not a worst-case guarantee over arbitrary adversarial keys. The processes row validates successful publication and artifacts, not one total compile across processes. These are owner-facing limits even though all three rows have concrete unique catches.

Replay: apply one D1.diff, D2.diff, D3.diff or D4.diff on the recorded base, source env.sh and use that record's exact command from matrix.json. For wider replay, run all current reached package rows with the appropriate larger budget centrally. No tests were deleted, rewritten or weakened. No main push and no pull request.
'''.format(base=(p/'base.txt').read_text().strip(),count=len(allrows),omitted=len(omitted))
(p/'REPORT.md').write_text(summary)
print('current rows',len(allrows),'omitted',len(omitted))
print([(r['mutant'],round(r['wall_seconds'],3),round(r['validation_seconds'],3)) for r in mat])

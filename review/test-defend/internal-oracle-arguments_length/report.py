import pathlib,json,subprocess,gzip,shutil
src=pathlib.Path('/tmp/def-args');dst=pathlib.Path('/workspace/adamic/review/test-defend/internal-oracle-arguments_length');dst.mkdir(parents=True,exist_ok=True)
m=json.loads((src/'matrix.json').read_text());by={x['mutant']:x for x in m}
rows=[]
for name,ids,verdict in [('TestCheckedCastRunsNoCatchOrFinally',['D2','D3','D4'],'not defended'),('TestCheckedCastFailureContract',['D1'],'defended')]:
 attempts=[{k:by[i][k] for k in ['mutant','file_line','change','rows_failed']} for i in ids]
 if name.endswith('Finally'):
  evidence=by['D3']['runs'][0]['command']+' > D3-casts.log 2>&1; cast_test.go:87: missing finally instrumentation site. D4 passed this row; both neighbor catchers failed.'
  subs=['TestCallTargetThrowAgreesWithNode','TestNativeAgreesWithNode'];unique=None
 else:
  evidence=by['D1']['runs'][0]['command']+' > D1-casts.log 2>&1; '+next(s for s in by['D1']['errors'] if 'Base<number>' in s)
  subs=[];unique='D1 internal/lower/cast_proof.go:311'
 rows.append(dict(test=name,package='internal/oracle',prior_verdict='subsumed',prior_subsumed_by=['TestCheckedCastFlushesOutput'],subsumed_by=subs,defense=verdict,unique_mutant=unique,attempts=attempts,evidence=evidence,bounded=True))
(dst/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
for p in src.iterdir():
 if p.name in ['cache','prior-preserve.log'] or not p.is_file():continue
 if p.suffix=='.log':
  with gzip.open(dst/(p.name+'.gz'),'wb') as f:f.write(p.read_bytes())
 else:shutil.copy2(p,dst/p.name)
base=subprocess.check_output(['git','rev-parse','HEAD'],cwd='/workspace/adamic',text=True).strip()
passed=by['D1']['rows_passed'];scope=sorted(set(passed+by['D1']['rows_failed']))
(dst/'scope.json').write_text(json.dumps({'starting_commit':base,'top_level_tests':scope,'aggregate_subcases':'TestNativeAgreesWithNode/internal/oracle/testdata/(cast|exceptions|try|finally)','bounded':True,'unknown':'Every other oracle row and other aggregate fixture subcase; repository-wide uniqueness.'},indent=2))
(dst/'REPORT.md').write_text('''One row defended within the bounded cast/exception matrix; one row not defended after three aimed attempts.
Production and tests restored; four standalone diffs pass go vet and apply to the starting commit.
Evidence is on test-defend/internal-oracle-arguments_length under this directory.

Starting commit: '''+base+'''. Audit starting commit: 0942c5169d0ea736d9dfaa19881af1ad8adad162. Both requested names exist in go test -list output at this start, in internal/oracle/cast_test.go. No requested row moved or vanished. The intervening current-main changes since the previous session affect other packages; current inventory, scope and logs are saved. We did not replay only the audit's old matrix.

CODE UNDER TEST: Adamic cast lowering, specifically checkedClassCast in internal/lower/cast_proof.go, and JavaScript statement emission for ir.Try in internal/javascript/javascript.go. Lower, native.C and javascript.JavaScript are public compilation entries. Native products used a separate ADAMIC_BUILD_CACHE_DIR per mutant; ADAMIC_GATE_UNCACHED=1 bypassed oracle evidence reuse. No harness, oracle, fixture or test changed.
ORACLE: self-written exit 70, stdout, panic text and absent handler-marker assertions. TestCheckedCastFailureContract also actually runs original source on Node and requires exit zero, but Node does not decide inserted cast failure text. TestCheckedCastRunsNoCatchOrFinally executes generated JavaScript on Node and instruments existing handler-string sites; its expected failure contract is self.

Coverage: each requested row and TestCheckedCastFlushesOutput ran individually with -coverpkg=./internal/lower,./internal/native,./internal/javascript and -coverprofile. The .cover files and .exclusive.json lists preserve covered blocks missing from the subsumer. NoCatch has 456 exclusive blocks, including Try statement emission and class-cast lowering. FailureContract has additional class-cast lowering blocks. Exclusive blocks are leads, not uniqueness proof. Semantic differences: Flush uses a string-tag cast without handlers; NoCatch uses numeric and class casts with try/catch/finally; FailureContract pins the exact panic text for eleven casts, including generic class target spelling. Runtime C and the Node process are not instrumented by Go coverage.

D1 removes the final byte from the class-cast message passed to l.constant, an off by one on its implicit string bound. The generic target Box<number> becomes Box<number. Both backends agree on this wrong text. Only FailureContract fails in the matrix. This demonstrates an assertion which backend agreement and the string-tag Flush fixture miss. Every other scoped top-level test passed: '''+', '.join(passed)+'''. scope.json lists the bounded aggregate fixture selection. No full-package uniqueness is claimed.

D2 drops catch-body emission. NoCatch fails at its missing caught instrumentation site, while CallTargetThrow and the NativeAgrees exception fixtures also fail. D3 drops finally-body emission and produces the analogous shared failure. D4 swaps the catch and finally emission arguments. NoCatch passes because both instrumented sites remain and panic still exits before either executes; the two ordinary exception agreement rows fail. These are three honest handler-emission attempts, not three unique kills. There is no evidence here to delete NoCatch. Its assertions do check the promised handler-marker absence on generated JavaScript, plus exit and stdout. It does not check which handler body belongs to which clause, as D4 demonstrates; that is ordinary exception agreement's job. It does not execute the native backend. No performance threshold is promised by either requested name.

Ambiguities and costs:
- /tmp is an 8.8 GB mount. Initial and subsequent df reports show about 5.9 GB free, so the demanded 15 GB cannot be reached on this filesystem. /workspace had 17 GB free. The previous unit's disposable per-mutant cache and scratch were removed after preserving its logs, never the repository or tools. No ENOSPC failure occurred.
- The whole-package baseline timed out at 90.205 seconds with no individual assertion failure. The bounded clean baselines passed in 8.459 and 7.636 seconds. A timeout is not a red main. All matrices use the 90-second binary and 120-second outer limits.
- The package is too large for repeated whole runs. The bounded matrix includes all current cast, uncheckable-cast, call-target, direct-closure and interface-cast tests, plus aggregate cast/exception/try/finally fixture neighbors. Scope and individual subcases are retained in complete logs. Kills elsewhere are unknown.
- A top-level aggregate fixture row is bounded to matching subcases. CallTargetThrow and DirectClosure are a checker family, although complete raw logs retain their separate top-level names. D1 remains unique after grouping them.
- NoCatch's first two failures are instrumentation-construction assertions, not observations that handlers ran after a panic. D4 is a real production corruption that this row survives. None proves the marker assertion redundant for every possible panic defect.
- The imported Node panic runtime lives in oracle/adamic.mjs. Mutating it risks changing the source oracle as well, so it was left intact. No harness weakening was used.
- The defender menu explicitly permits an implicit-bound off by one; D1 uses exactly that allowance. D2 and D3 drop entire emission statements; D4 swaps two existing arguments. No supplemental insertion is counted.
- The row comment calls the cast contract externally decided, but the expected messages and exit 70 are Adamic's handwritten contract; source Node runs unchecked. The report keeps those separate.
- Cold command times combine Go recompilation, native product builds and execution. Build-only costs were not isolated. No performance-row mutation is applicable to these names.

Warm tools: setup skipped, nproc 5; npm ci completed and its log is saved. The four vet-plus-two-matrix command times were '''+', '.join(f"{r['mutant']} {r['seconds']:.3f}s" for r in m)+'''. Logs preserve own-binary elapsed values. Full-suite replay, other packages, native handler-cost faults, complete optional SDK/corpus gates and unrestricted aggregate fixtures are outside this defense. Final restored vet and both clean reruns are saved.
''')

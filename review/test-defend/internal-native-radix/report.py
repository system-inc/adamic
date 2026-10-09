import pathlib,json,subprocess,gzip,hashlib
p=pathlib.Path('review/test-defend/internal-native-radix');scope=json.loads((p/'scope.json').read_text());matrix=json.loads((p/'matrix.json').read_text());added=json.loads((p/'added-test-matrix.json').read_text());coverage=json.loads((p/'coverage-runs.json').read_text());rows=[];passlists=[]
assert len(matrix)==2 and all(r['exit']==0 for r in added)
for m in matrix:
 assert m['rows_failed']==[m['target']]
 target=m['target'];examples=[]
 for line in (p/'logs'/(m['mutant']+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if 'DISAGREEMENT case=' in e.get('Output',''):examples.append(e)
 failed=m['failed_members'];error=m['errors'][0]['Output'].strip();example=examples[0]['Output'].strip()
 rows.append({'test':target,'package':'internal/native','prior_verdict':'subsumed','subsumed_by':[],'prior_subsumed_by':['TestRegExpSearchNode'],'defense':'defended','unique_mutant':m['mutant']+' '+m['file_line'],'attempts':[{'mutant':m['mutant'],'file_line':m['file_line'],'change':m['change'],'rows_failed':m['rows_failed']}],'evidence':'ADAMIC_BUILD_CACHE_DIR='+m['env']['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(m['command'])+'; '+error+'; '+example,'bounded':True,'matrix_rows':scope['matrix_rows'],'members':scope['family_members'] if 'family' in target else [target],'oracle':'Live Node 24 capture spans, named groups, match text and lastIndex.' if 'family' in target else 'Recorded Node 24 test262 observations at pin 7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd. U+2028 no-match case checked against live Node this session.','oracle_kind':'external-run' if 'family' in target else 'external-authority'})
 passed=m['passed_members']+['TestRuntimeKeyKeepsBoundaries'];passedrows=sorted(set('TestRegExpBytecodeRandomNode family' if n.startswith('TestRegExpBytecodeRandomNodeUnit') else n for n in passed));assert len(passedrows)==10
 passlists.append({'mutant':m['mutant'],'rows_passed':passedrows,'members_passed':sorted(passed),'members_failed':failed})
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n');(p/'unique-pass-lists.json').write_text(json.dumps(passlists,indent=2)+'\n')
base=scope['base']
report=f'''Both requested rows are defended within the current regexp-reach matrix; neither test was changed.
Starting main {base}; all 40 family members and TestRegExpBytecodeTest262 exist in regexp_test.go.
Evidence includes two standalone diffs, measured coverage, full failing logs and ten passing grouped rows per mutant.

'''+json.dumps(rows,indent=2)+'''

Code under test and oracles

The code under test is Adamic's production regexp compiler generating native bytecode and generated runners, plus the C bytecode interpreter, UTF-16 input/search handling and match-result construction. Compile and NativeDeclarations are production inputs to the native tests, not an oracle. No Go matcher execution is used as an expected answer. No test, source fixture, oracle, or comparison harness was changed.

TestRegExpBytecodeTest262 loads 127369 recorded execution observations from matches.json.gz. It compares match/no-match, lastIndex, capture counts, capture text, UTF-16 spans and named-group spans against those values, with ASan/UBSan and leak detection. Its authority is the recorded Node execution of the pinned test262 extraction, not live execution of the full test262 suite. Case 124215, /^.$/mu with U+2028 input, expects no match; the exact expectation was independently confirmed on Node 24 this session.

The random family comprises all forty TestRegExpBytecodeRandomNodeUnit00 through Unit39 wrappers over one seeded generator and one checker. Each wrapper generates the same 10000 cases, asks live Node 24 for captures, groups and lastIndex, then checks its disjoint 250-case native slice. Family grouping therefore turns all forty D02 failing members into one failing row. The checker runs limited generic execution and unlimited execution with generated runners eligible.

Coverage and semantic differences

Requested Go coverage profiles use -coverpkg=./internal/native,./internal/regexp, one target run, one full family run, and one TestRegExpSearchNode run. Every coverage command passed. Inclusive covered-line differences versus the subsumer contain 384 lines for test262 and 223 for the random family, listed in coverage-exclusive.json. Go coverage does not instrument embedded C, so these numbers do not claim C branch exclusivity. The compiler dot branch is shared, but its input differs at U+2028. The C repeat branch is shared, but its closed-zero boundary differs.

D01 moves the lower bound of the non-dotAll line-separator exclusion from U+2028 to U+2029. The recorded corpus has 38 U+2028 inputs; the random generator and SearchNode subsumer do not include this character. Case 124215 changes from no match to a one-character match. Only test262 fails. This is a production compiler range bound mutation, not an edit of test262 expected values or an oracle.

D02 changes the runtime repeat-body predicate so a closed zero maximum allows the body, as if unbounded. The random generator includes {0}; the recorded corpus contains no closed zero quantifier bounds, including zero-padded and {0,0} spellings checked by the input scanner. The SearchNode probes also contain no closed-zero quantifier. All forty random members fail and every other reached row passes. Some disagreement summary lines have equal match/capture counts: the unchanged checker additionally compares exact spans and capture text, so this catch does not rest solely on counts or exit status.

Reach matrix, baseline and additions

The full current package lists 282 top-level tests. Its clean baseline exceeds the 90s binary budget, with no observed individual assertion failures before its timeout. It was not used as a green full-package result. The statically reached slice passed in 23.278s, with no skips. Mutants ran every top-level test in regexp_test.go and regexp_search_test.go, plus TestRegexProgramsKeepCheckedFieldReads and TestClosureConventionRuntimeDropCount. These cover direct native regexp VM APIs, generated native regex field/group integration, and the regexp callback compilation witness. Pure Go regexp utilities in other native test harnesses do not execute the native VM. Kills outside the statically reached set are unknown, including external/opt-in product inputs.

Two tests were added since the audit: TestRegExpNativeStepLimitBoundary and TestRuntimeKeyKeepsBoundaries. The former is in each main mutant run. The latter, which checks cache-key encoding and does not execute the regexp interpreter, passed separately under each mutant. scope.json lists all current package tests, the forty members, and all eleven grouped matrix rows. unique-pass-lists.json lists each of the ten other passing grouped rows, and all passing top-level members, for each unique defense.

Brief ambiguities and costs

1. The supplied audit commands are truncated. The full-refspec fetched branch has the complete commands, report, rows, scope and mutant plan. That audit was bounded from older main 83f3940ec77b8ba779da05ded113e9d9e7e710b6, rather than proving package-wide subsumption. This defense also states its bounded reach explicitly; central wider replay remains necessary.
2. A family is not a top-level Test function named with 'family'. The forty current wrappers exist and share one checker over slices of the same deterministic cases. They count as one row; no member is offered as a subsumer.
3. The package baseline timeout cost 90.018 binary seconds. The user permits narrowing big packages, so a separately green reached baseline was established before mutants. No mutant was run on an observed red baseline. Matrix binary times remain below 90s.
4. Go coverage cannot measure C runtime lines. Both requested Go profiles and the production compiler coverage are retained; C reach claims use source callers and input semantics rather than fabricated C coverage. No C-exclusive-line claim is made.
5. Mutation D01 is in internal/regexp because native tests call its production compiler before building the native executable. Its Go interpreter is not the oracle here. D02 is in the native runtime. Both are valid code-under-test changes, and their source locations refer to the actual starting main.
6. ADAMIC_BUILD_CACHE_DIR is distinct per mutant as requested. Inspection of library.go shows the C runtime archive cache itself uses os.UserCacheDir()/adamic/runtime and hashes complete runtime source/header bytes, compiler identity and flags. It does not use that environment variable. D02 therefore gets a different source-content key and rebuild; D01 regenerates the changed C declarations and compiles each executable. No claim that the environment variable isolates the runtime archive cache is made.
7. The runtime mutant's compile check uses the repository's strict C11 warning flags plus counted ASan/UBSan flags. D01 passes go vet ./internal/regexp/. Native test builds additionally compile generated executables using their normal flags. Separate native rebuild time is not exposed; wall and binary elapsed remain recorded without inventing a rebuild estimate.
8. Each row had a unique catch on its first aimed attempt, so no second or third attempt was needed. There are no not-defended or cannot-judge rows, and therefore no unresolved name/assertion mismatch finding. Test262 names recorded provenance, not complete live conformance execution; the random family's finite seeded corpus is not exhaustive.
9. The old audit's four caught mutants exercised broad budget, range membership, string equality and release behavior. They did not establish redundancy for the U+2028 dot boundary or closed-zero repeat boundary demonstrated here.

Timing and exclusions

Warm toolchain worked; setup skipped. nproc 5. npm ci log is retained. Clean full baseline cooks at 90s; reached baseline is 23.278s. Per-command wall and binary times are in slice-baseline-run.json, coverage-runs.json, matrix.json and added-test-matrix.json. Coverage, mutation and compile costs are summarized in timing.json.

Not covered: full-package or repository-wide uniqueness, opt-in external product inputs, dynamic C coverage, exhaustive random inputs, or complete live test262 execution. Standalone diffs apply independently to starting main, compile, and all production sources are restored byte for byte. No test deletion, rewrite, weakening, PR, or main push.
'''
(p/'REPORT.md').write_text(report)
timing={'coverage_wall_seconds':sum(r['wall_seconds'] for r in coverage),'mutation_wall_seconds':sum(r['wall_seconds'] for r in matrix),'standalone_compile_seconds':sum(r['compile_seconds'] for r in matrix),'added_test_wall_seconds':sum(r['wall_seconds'] for r in added)};(p/'timing.json').write_text(json.dumps(timing,indent=2))
for file in set(r['file'] for r in matrix):assert pathlib.Path(file).read_bytes()==subprocess.check_output(['git','show',base+':'+file])
with (p/'logs/apply-check.log').open('w') as log:
 for diff in sorted((p/'diffs').glob('*.diff')):
  code=subprocess.call(['git','apply','--check',str(diff)],stdout=log,stderr=log);log.write(str(diff)+' exit='+str(code)+'\n');assert code==0
subprocess.check_call(['git','diff','--exit-code','--','internal/native','internal/regexp','stage3/api/package-lock.json'])
for file in (p/'logs').glob('*.log'):
 file.with_suffix('.log.gz').write_bytes(gzip.compress(file.read_bytes(),mtime=0));file.unlink()
manifest=[]
for file in sorted(p.rglob('*')):
 if file.is_file() and file.name!='SHA256SUMS':manifest.append(hashlib.sha256(file.read_bytes()).hexdigest()+'  '+str(file.relative_to(p)))
(p/'SHA256SUMS').write_text('\n'.join(manifest)+'\n');print('two bounded defenses verified',timing)

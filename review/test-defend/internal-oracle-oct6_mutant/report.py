import pathlib,json,gzip,shutil,subprocess
p=pathlib.Path('/tmp/def-oct6');r=pathlib.Path('/workspace/adamic');dest=r/'review/test-defend/internal-oracle-oct6_mutant';dest.mkdir(parents=True,exist_ok=True)
base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=r,text=True).strip();m=json.loads((p/'matrix.json').read_text());new=json.loads((p/'new-test-results.json').read_text());scope=json.loads((p/'scope.json').read_text());audit=json.loads((p/'audit-results.json').read_text());menu=json.loads((p/'menu.json').read_text())
for x in menu:
 if x['id']=='D01':x['line']=293
(p/'menu.json').write_text(json.dumps(menu,indent=2)+'\n')
for x in m:
 if x['id']=='D01':x['file_line']='internal/native/runtime/input.c:293'
 added=next(y for y in new if y['mutant']==x['id']);x['runs']+=added['runs'];x['failed']+=added['failed'];x['passed']+=added['passed'];x['skipped']=added['skipped'];x['bounded']=True
(p/'matrix.json').write_text(json.dumps(m,indent=2)+'\n');scope['top_level_rows']+=json.loads((p/'test-inventory.json').read_text())['new'];scope['new_patterns']=json.loads((p/'new-test-patterns.json').read_text());scope['actual_native_subcases']=m[0]['runs'][1]['native_subcases'];(p/'scope.json').write_text(json.dumps(scope,indent=2)+'\n')
rows=[]
for t,ids in [('TestAPromptComesBeforeTheRead',['D01']),('TestOverloadContractRulings',['D02','D03','D04'])]:
 chosen=[x for x in m if x['id'] in ids];a=next(x for x in audit if x['test']==t);first=chosen[0]
 rows.append({'test':t,'package':'internal/oracle','prior_verdict':'subsumed','subsumed_by':a['subsumed_by'],'defense':'defended','unique_mutant':first['id']+' '+first['file_line'],'attempts':[{'mutant':x['id'],'file_line':x['file_line'],'change':x['change'],'rows_failed':x['failed']} for x in chosen],'evidence':first['runs'][0]['command']+'; '+first['evidence'][0],'bounded':True,'matrix_rows':scope['top_level_rows'],'passed_rows':first['passed']})
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
# Profiles instrument Go compilation and emission, not linked C branches.
for f in p.iterdir():
 if f.is_dir():continue
 if f.suffix in ['.log','.cover']:(dest/(f.name+'.gz')).write_bytes(gzip.compress(f.read_bytes(),mtime=0))
 else:shutil.copy2(f,dest/f.name)
report=f'''Both requested rows defended within a bounded matrix.
D01 uniquely catches prompt-before-input ordering; D02/D03/D04 uniquely catch overload rulings.
Evidence branch: test-defend/internal-oracle-oct6_mutant; 27 top-level rows, 26 actual NativeAgrees subcases.

Base origin/main: {base}. Audit evidence was fetched with the required full refspec and read from REPORT.md, scope.md and results.json. The audit started at 6c60da09; today the package has 198 top-level names versus 193 then. Five added names and no vanished names are in test-inventory.json. Every added top-level row is represented in this matrix, with TestReviewProgramsAgreeWithNode explicitly bounded to relevant overload fixtures plus smoke.a. Both target names remain in their original files. Warm toolchain reused, setup skipped, nproc 5. npm ci in stage3/api passed before baseline.

CODE UNDER TEST / ORACLE:
- TestAPromptComesBeforeTheRead checks native runtime adamic_read_text_file in internal/native/runtime/input.c, particularly the flush at entry before a blocking read. Lower and native emission prepare the native product, and the JavaScript backend is also compared. Oracle: the source executes through Node, which must prompt before input, and produce ready\\ngot yes\\n. The harness supplies yes only when a first line arrives, or after its ten-second failure deadline. This is a protocol check, not an exit-only or eventual-output check. The handwritten output guard is self; the Node dialogue is external-run. Neither Node nor the driver is mutated.
- TestOverloadContractRulings checks lower.censusOverloads / censusOverload and its result/parameter relations. Oracle: Node executes three untouched sources, with handwritten output expectations. The lowerer must return lower.Refused with handwritten overload result or parameter label. That static ruling is self, not a fact Node proves. No expected strings, oracle implementation, harness or test were changed.

Coverage:
The target and subsumer TestOmittedOriginalProbePolicy were individually run with `-coverpkg=./internal/lower,./internal/native -coverprofile=...`, forcing ADAMIC_GATE_UNCACHED=1. Coverage-runs.json records each exact command and status, and all four runs passed. TestFileWritesLandInNodesOrder was also profiled as an additional prior subsumer. Target versus OmittedOriginal shows 114 exclusive Go blocks for the prompt row and 138 for the overload row. Exclusive-block files and reached-function summaries are included. The overload row reaches result-variance, binder-inference and parameter-contravariance blocks not exercised by that subsumer.

Go coverage does not measure linked runtime C branches. I collected the requested per-test Go profiles for the compiler paths, then used a semantic difference for the runtime mutation: a driver that withholds stdin until stdout is visible. The same prompt_then_read.a source is also in NativeAgrees, whose ordinary input setup does not detect this handshake failure. Runtime C branch exclusivity is unknown; the observed unique kill is established by the bounded matrix, not inferred from Go coverage.

Clean baselines and scope:
The whole package timed out at 90.450 seconds with no completed failing test; this was budget exhaustion, not a red baseline. Its complete log is preserved. Clean bounded ordinary rows and selected NativeAgrees fixtures passed, as did the added-test baseline. The first scope has 21 complete ordinary top-level rows plus NativeAgrees restricted to 26 actual fixture subcases. Five current additions bring it to 27 top-level rows. Caller-fixtures.json records a conservative source search for readTextFile and overload signatures. That search cannot prove every dynamic caller, imported overload or inline program was found. Other package rows, NativeAgrees fixtures, and unselected review agreement cases remain unknown. No package-wide or repo-wide uniqueness is claimed.

Exact matrix commands, statuses, failing lines and passed top-level names are in matrix.json and rows.json. scope.json lists all patterns and actual native subcases. Each mutant has its own ADAMIC_BUILD_CACHE_DIR; all runtime observations bypass the result cache. Each standalone diff applies to {base}.

Mutants:
| ID | base file:line | change | failed row |
|---|---|---|---|
| D01 | internal/native/runtime/input.c:293 | Drop adamic_output_flush before the read | TestAPromptComesBeforeTheRead |
| D02 | internal/lower/census_small.go:242 | Drop incompatible result Refused return | TestOverloadContractRulings |
| D03 | internal/lower/census_small.go:150 | Change static overload ordinal increment 1 to 2 | TestOverloadContractRulings |
| D04 | internal/lower/census_small.go:203 | Drop whole parameter-check loop and its local declaration | TestOverloadContractRulings |

D01 evidence: output_test.go:286: program: prompted before the read false, said "ready\\ngot yes\\n"; Node said "ready\\ngot yes\\n". The eventual answer remains correct. Only the prompt row detects the lost flush ordering. All other 26 top-level rows passed, subject to the documented native/review bounds and pending skips.
D02 evidence: overload_contract_test.go:34: want refusal "overload 1 of createToken result", got <nil>. Trampoline also loses the required result refusal, producing a later NotYet. All other 26 top-level rows passed in the selected scope.
D03 changes diagnostic ordinal only; the row detects overload 2/4 where 1/2 are required. D04 admits serializer without its required callback-parameter refusal. Both also fail only the overload ruling row in this scope. No survivors among the four attempts.

Standalone validation:
D01 passed clang syntax compilation under native runtime strict C11, warning, sanitizer and floating-point flags; the actual matrix also rebuilt and executed native products. D02-D04 passed `go vet ./internal/lower/` while applied. validate_command is recorded per mutant. All diffs subsequently passed git apply --check against the restored base. The final uncached two-target run passed. Production sources are restored and only review evidence is committed.

Pending subcases, consistent in the clean added baseline and every mutant:
{json.dumps(new[0]['skipped'],indent=2)}
These are repository pending sidecars, not missing installable tools. They were not weakened or overridden. ReviewProgramsRefuse ran 53 active cases; ReviewProgramsAgreeWithNode ran smoke.a and observed its three relevant pending fixtures. The other four new top-level tests ran their complete bodies.

Passed-row lists for the unique catches:
D01: {json.dumps(next(x for x in m if x['id']=='D01')['passed'])}
D02: {json.dumps(next(x for x in m if x['id']=='D02')['passed'])}
Full per-attempt lists and native subcase lists are retained in matrix.json.

Brief friction, ambiguities and limits:
- /tmp is only 8.8 GB total, so the 15 GB free requirement is impossible. It had 6 GB free at entry and remained near that after removing the prior unit's /tmp/def-regexp scratch directory. /workspace had 18 GB free. No tools or repository contents were deleted.
- The prompt and overload rows were called subsumed without a named subsumer in the brief. The audit names multiple mutual subsumers, all resting on M4, a general Lower-entry check failure. I used OmittedOriginal for both requested coverage comparisons and FileWrites as an additional profile. That old failure did not probe either row's distinguishing behavior.
- The full package exceeds the binary budget and was narrowed only after its clean timeout. Exact reachability cannot be proven solely by fixture-name/source searches. All uniqueness claims are bounded and outside-scope results are unknown.
- The package gained five tests. The large new review agreement corpus has 208 cases, 127 pending. Only its relevant overload fixtures and a smoke control were replayed; all other new rows ran fully. This bounding is explicit rather than treating omitted cases as passing.
- Go coverpkg does not instrument the C runtime. Profiles support compiler-path analysis, while the C defense rests on the input protocol and observed matrix. C branch coverage was not produced.
- The ordinal edit guard found its text twice and stopped before planting. I restricted it to the first occurrence, censusOverloads. The separate runtime-result ordinal remained unchanged. This stopped preparation run counts as no attempt result.
- Dropping only the parameter refusal block left an unused takes binding and failed go vet. I instead dropped the whole checking loop plus its local declaration. Only the compiled D04 version supplies evidence; no failing-build kill is counted.
- The initial audit report dump was large and truncated in tool output; I extracted the two assigned results and report's limits separately to resolve their exact subsumers and oracle descriptions.
- Both row names match their asserted behavior: prompt visibility before read, and three specified static overload rulings. Neither was left undefended and neither names a throughput or performance promise requiring a work-growth test. The prompt mutant nevertheless preserves eventual output and violates the timing protocol.
- Warm setup cost was zero; npm duration was not separately instrumented. Coverage commands include instrumentation compilation. Original bounded matrix command wall time was {sum(sum(y['wall_seconds'] for y in x['runs'][:2]) for x in m):.3f}s; new-test replay including its clean baseline was {sum(sum(y['wall_seconds'] for y in x['runs']) for x in new):.3f}s. Runtime rebuild-only timing was not separated from test/build command wall time, which is recorded for every run. No mutant run exceeded 90 seconds and no mutant panic occurred.
- No other package tests were run. There is no whole-package uniqueness claim, no PR, no main push, no test deletion, rewrite or weakening.
'''
(dest/'report.md').write_text(report)
print('Saved',len(list(dest.iterdir())),'files; rows',len(scope['top_level_rows']))

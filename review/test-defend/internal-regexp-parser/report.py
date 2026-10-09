import pathlib,json,gzip,subprocess,shutil
p=pathlib.Path('/tmp/def-regexp'); r=pathlib.Path('/workspace/adamic'); dest=r/'review/test-defend/internal-regexp-parser';dest.mkdir(parents=True,exist_ok=True)
m=json.loads((p/'matrix.json').read_text()); isolated=json.loads((p/'D08-isolated.json').read_text()); m[7]['package_aborted']=True;m[7]['failed_before_abort']=m[7]['failed'];m[7]['passed_before_abort']=m[7]['passed'];m[7]['failed']=[x['test'] for x in isolated if x['exit']];m[7]['passed']=[x['test'] for x in isolated if not x['exit']];m[7]['target_evidence']=[x['evidence'] for x in isolated if x['test']=='TestQuantifierBounds'][0];m[7]['recovery_commands']=[x['command'] for x in isolated]
(p/'matrix.json').write_text(json.dumps(m,indent=2)+'\n');names=[x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')]; base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=r,text=True).strip()
rows=[]
for target,prior,ids in [('TestParse','TestMatcherAgreement family',['D01','D02','D03']),('TestFlags','TestNodeAgreement',['D04','D05','D06']),('TestQuantifierBounds','TestMatcherOct6LoopsNode',['D07','D08','D09'])]:
 chosen=[x for x in m if x['id'] in ids];first=chosen[0]; defended=target=='TestParse'; evidence=first['command']+'; '+first['target_evidence'][0]
 rows.append({'test':target,'package':'internal/regexp','prior_verdict':'subsumed','subsumed_by':[prior],'defense':'defended' if defended else 'not defended','unique_mutant':'D01 internal/regexp/parser.go:474' if defended else None,'attempts':[{'mutant':x['id'],'file_line':'internal/regexp/parser.go:'+str(x['line']),'change':x['change'],'rows_failed':x['failed']} for x in chosen],'evidence':evidence})
(dest/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
for f in p.iterdir():
 if f.name=='doctrine.txt' or f.is_dir():continue
 if f.suffix in ['.log','.out']:
  (dest/(f.name+'.gz')).write_bytes(gzip.compress(f.read_bytes(),mtime=0))
 else:shutil.copy2(f,dest/f.name)
report=f'''TestParse defended by D01 and D02; each fails only TestParse.
TestFlags and TestQuantifierBounds not defended after three attempts each.
Complete package matrix: 18 top-level names, 13 rows after grouping; evidence branch test-defend/internal-regexp-parser.

Starting origin/main: {base}. Historical audit baseline was 09fe4b54913753188a9357982bfd47cdf36ef97c. All requested names remain in parser_test.go. New TestUnicodeDecimalEscape is included. Warm tools worked; setup skipped; nproc 5. npm ci completed successfully before baseline. Clean baseline binary: 2.142 seconds. No skipped tests. Final restored run is in final-clean.log.gz.

CODE UNDER TEST: Go regexp Parse and parser helpers in internal/regexp/parser.go. No C or native product rebuild is involved. ORACLE: TestParse's handwritten valid/invalid booleans, TestFlags' handwritten uu/uv/z rejection expectations, and TestQuantifierBounds' handwritten Min=2, Max=4, Greedy=false expectation. All are self. Error presence alone decides the first two; neither checks diagnostic wording or AST. Subsumers compare Node admission, Node execution and stored test262 results. Their oracle and harness were unchanged.

Coverage commands and comparison:
- TestParse: coverage-0.out.gz versus matcher family coverage-5.out.gz. Fifteen exclusive blocks, listed in exclusive-0.txt. These are predominantly rejection paths, while matcher execution needs valid inputs. D01/D02 remove left and right ClassRange admission guards, independently. The row checks invalid un-nested ranges [a-z&&[^aeiou]] and [a&&b-z]. Existing valid nested-set execution controls keep passing. D03 removes the separate negated-string-class rejection and is caught by Node admission too.
- TestFlags: coverage-1.out.gz versus TestNodeAgreement coverage-3.out.gz. No exclusive blocks. The semantic lead is empty pattern with invalid flags, whereas Node's corpus includes nonempty uu/uv/z patterns. Three attempts weaken duplicate tracking, u/v exclusion and unknown-flag rejection. All three show Node also catches those validation faults. These attempts do not exhaust every possible fault dependent on empty-pattern context.
- TestQuantifierBounds: coverage-2.out.gz versus loop subsumer coverage-4.out.gz. Seventeen exclusive blocks, listed in exclusive-2.txt, including explicit brace-bound parsing. It examines exact AST fields for a{{2,4}}?, unlike execution comparisons. D07 drops finite upper-bound storage; D08 swaps Min/Max fields; D09 flips greediness. Execution rows catch each as well. TestNodeAgreement is admission-only and passes D08/D09, showing why it cannot replace execution/AST checks for those faults.

Exact coverage command shape, performed once for each listed expression:
`timeout 120 go test -count=1 -timeout 90s -run '^EXPRESSION$' -coverpkg=./internal/regexp -coverprofile=/tmp/def-regexp/coverage-N.out ./internal/regexp/ > coverage-N.log 2>&1`
Expressions by N: 0 TestParse, 1 TestFlags, 2 TestQuantifierBounds, 3 TestNodeAgreement, 4 TestMatcherOct6LoopsNode, 5 TestMatcher(NodeControls|RandomNode|Test262Executions|CanonicalizeNode|UTF16PatternsNode|Oct6Node).

Every mutant command and failing line is in rows.json and matrix.json. Each standalone diff applies to {base} and passed `go vet ./internal/regexp/` while applied. D07 discards the now-unused decimal result as part of dropping the assignment; the enclosing parser still consumes the bound. D08 swaps struct arguments, including nil unbounded maxima, so it panics in matcher code. The initial package run is incomplete. Every top-level name was then run alone, with the same mutant/cache and exact-name regex, producing D08-Test*.log.gz and D08-isolated.json. No inference about later tests was made from the abort. D08's recovered failures are recorded in matrix.json. The witness TestMatcherOct6Mutants fails a control precondition under D08/D09; that is not independent proof of its own checker, and it is unnecessary to either non-defense conclusion.

D01 and D02 each fail TestParse; all these other names passed in both full-package runs:
{json.dumps([n for n in names if n!='TestParse'],indent=2)}

The matcher family groups its six shared compareExecutionCases wrappers. All 18 raw names are retained in logs and matrix. Uniqueness is across the complete current package, not the repository. No matrix cooked past 90 seconds. The final clean package run passed, sources restored, tests unchanged. Only review evidence is committed.

Owner findings for non-defended rows:
- TestFlags has a broad name but asserts only three invalid flag strings. It does not assert valid flag decoding, ordering or every invalid flag. Its current assertions do match the narrow rejection behavior they implement; no performance promise is present. Three shared catches are not a deletion recommendation, particularly because empty-pattern-specific faults were not exhaustively explored.
- TestQuantifierBounds correctly asserts the exact lower/upper/lazy AST values for one bounded quantifier. It does not promise or assert a general boundary corpus, unbounded values or overflow handling. There is no performance claim or threshold missing. Three shared catches are a limited finding, not proof of universal subsumption.

Brief friction and limits:
- /tmp is an 8.8 GB mount, so 15 GB free cannot be achieved there. Before work it had 2.9 GB free. Removing only named earlier-unit scratch directories recovered about 3 GB, leaving 6 GB free. /workspace had 18 GB free throughout. No repository or tools were deleted. Old adamic-gate contents were preserved because their role was less certain.
- The audit's historical main differs from today's main, and its old mutant lines had to be mapped to current production functions. It also predates TestUnicodeDecimalEscape. Today's baseline and every matrix include that new row.
- A raw search of the single-line JSON corpus generated excessive output. I repeated it with JSON parsing to extract only relevant flags and quantifier inputs. It supplied no mutant verdict by itself.
- Go block coverage leads are coarser than source lines and do not prove input uniqueness. Empty-pattern versus nonempty-pattern flag validation shares all covered blocks. I report that limit rather than claim exhaustive redundancy.
- The field-swap attempt caused a package panic, costing 18 isolated reruns. All their results were recovered and are retained. No timeout or red baseline was treated as a mutant catch.
- The test witness's broken production precondition is distinguished from a genuine witness check failure.
- Node/module installation was required by the brief even though this package's Node scripts use built-in modules. The warm toolchain required no setup. npm setup duration and coverage-build duration were not separately instrumented. Matrix command wall seconds are recorded individually; aggregate original matrix wall time is {sum(x['wall_seconds'] for x in m):.3f}s, and panic recovery commands total {sum(x['wall_seconds'] for x in isolated):.3f}s. These include Go build/process overhead, not just binary time.
- No other package was tested. There are no twin executors or cost rows among these three requested rows. No tests were deleted, rewritten or weakened. No PR and no main push.
'''
(dest/'report.md').write_text(report)
print(json.dumps(rows,indent=2))

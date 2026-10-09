All three requested rows have a distinct catch in the current whole-package replay.
Each mutant fails only its target; 273 other rows pass, and two corpus/config rows skip.
Sources and tests restored; preserve all three; skipped-row and repo-wide uniqueness unknown.

# Prototype defense

Starting origin/main: 619e7a4cf33741cc04bc78dc4c0c8ba0e31d73fb. Audit base: 5deb11d433947539ab6448558b8ab1f1904393a5. Branch test-defend/internal-lower-prototype. Audit fetched with the requested full remote-tracking refspec. Read CLAUDE.md, the audit REPORT, all fourteen rows' oracle/verdict/kill records and menu, target test files whole and relevant production functions. README.md, docs/0.1.md and docs/memory.md were read earlier in the warm session and have no changes from that session's base. No external checker, oracle, test, harness or fixture edited. The mutation files are internal production code, including native RegExp serialization called by lowering.

CODE UNDER TEST: Lower/refuse/representation/regexConstant and internal/regexp Program.NativeDeclarations. ORACLE: all three rows use self-written expected rejection classes, absence of error, or representation/known pairs. Representation's checker assignability assertions establish input preconditions. They do not execute Node as their answer oracle. No external-authority claim.

## Clean run and current scope

Warm env.sh worked; setup skipped; nproc=5. npm ci --prefix stage3/api completed before the baseline, added three packages in 370ms. First df: /tmp 8.8G total, 285M used, 8.6G free; /workspace 32G total, 18G used, 13G free. Removed only the previous unit's named /tmp/defend-elements scratch/cache directory. Second df: /tmp 241M used, 8.6G free; workspace unchanged. The 15G floor is impossible on the 8.8G /tmp mount. No tools/repository removed; no disk-related baseline failure.

Clean whole-package command: timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . (baseline.log). Passed 34.978 binary seconds, 274 top-level passes and 2 skips. Current -list has 276 top-level tests; audit list 239. All 37 additions included in every mutant run; no vanished test. Full names and differences in scope.json. No bounded subset selected, no panic or timeout in a replay.

The two top-level skips are TestOriginalCycleLedger and TestOptionalWideningCensus. They need a pinned pristine TypeScript snapshot/generated diagnostics or caller-selected project config/output, not another missing executable. They remain unknown. TestMixedUnionContractGraph also skips a documented subcase; complete subcase skip records are in each log/results.json. Every target and named prior subsumer executed. One failed top-level row per mutant means a distinct catch among the 274 executed rows, not proven uniqueness against the two skipped rows or outside this package.

## Coverage and aimed differences

Per-test coverage used -coverpkg=./internal/lower,./internal/regexp, anchored -run and separate -coverprofile. All five isolated target/subsumer runs passed. The RegExp comparison ran all other 275 listed rows, with the same skips. Coverage-runs.json contains exact argv and times; .cover files, function inventories and exclusive-coverage.json retain the evidence. All clean coverage finished before any mutation replay.

RegExp refusal has 51 exclusive positive blocks versus the rest of the package, including native.go:18.5,19.1 and lower/regexp.go:64.3,65.1. The audit menu never changed the native counter-width guard. Its source a{18446744073709551616} is beyond uint64, unlike the other package regex inputs. D1 drops the entire existing overflow-refusal if statement at native.go:17. Serialization still uses Uint64, so the huge count is truncated and the lowerer admits the previously refused program. No oracle/matcher mutation and no empty-entry probe.

Representation has 2 exclusive blocks versus TestPrototypeMethodsAreRefusedWithReasons, including expression.go:64.3,66.1, the substitution lookup for unresolved generic parameters. Concrete ordinary types have one ABI; generic structural constraints can admit both closures/objects or arrays/strings. D2 changes the returned known bit totrue for unresolved substitutions. Its now-unused binding is discarded with _, ensuring compilation. Both source and lengthSource fail with (0,true) versus (0,false); all other package rows pass.

Suppression sound neighbors has 309 exclusive blocks versus its tagged-interface subsumer, but the aimed distinction is semantic on the shared refuse entry: @ts-expect-error inside prose/literals is not a suppression directive. The neighboring definite-assignment row contains @ts-ignore only. D3 performs a guarded early return from refuse for lexical @ts-expect-error occurrences when the parser's directive list is empty. This deliberately replaces syntax-aware admission with a false positive. It is the allowed return-early operation, with no filename/test-name selector and no change to the parser. Both the string-literal and prose-comment sound neighbors fail; all actual directive refusal tests still pass.

## Mutants and observed failures

* D1, internal/regexp/native.go:17: Drop the counter-width refusal; huge quantifier counts truncate during Uint64 serialization. Only TestRegExpNativeRefusals fails. Evidence: regexp_test.go:24: expected loud NotYet for const matches = /a{18446744073709551616}/.test('a');: <nil>
* D2, internal/lower/expression.go:64: Claim unresolved generic parameters have a known ABI even without a substitution; discard now-unused known binding. Only TestRepresentationClockSourceCheckedTypes fails. Evidence: representation_clock_source_test.go:69: source: representation = (0, true), want (0, false)
* D3, internal/lower/refusals.go:39: Early reject the directive token lexically even when parser says prose/literals contain no directive. No input path/name selection. Only TestSuppressionDirectiveSoundNeighbors fails. Evidence: suppression_directives_test.go:56: sound neighbor: /tmp/adamic-gate/TestSuppressionDirectiveSoundNeighbors962150014/002/main.a: Adamic 0.1 refuses suppression directive; remove it and fix the type error

Standalone D1.diff/D2.diff/D3.diff apply to starting main with no environment switch. Each was physically applied, vetted in its mutated Go package and replayed over the whole current lower package with ADAMIC_GATE_UNCACHED=1 and its own ADAMIC_BUILD_CACHE_DIR=/tmp/defend-prototype/cache/Dn. Exact full argv/results/timings in results.json; all 273 other passing rows listed per mutant in passed-rows.json. Test output went directly to log files. No sanitizer, clang, timeout or compilation failure is counted as a kill. Sources restored in finally blocks; all three clean targets then passed (restored.log), and git diff --exit-code -- internal passed. Standalone git apply --check logs saved.

One aimed mutant defended each row, so three failed attempts were unnecessary. No cost-only row and no executor twins in this unit. RepresentationClockSourceCheckedTypes checks ABI classification rather than elapsed clock time. No survivors among these three aimed mutants. No other packages' tests ran; go vet ./internal/regexp compiles the production serializer changed by D1.

## Brief ambiguities, costs and findings for owners

* The 15GB disk floor cannot fit on /tmp. Cleanup recovered 44MB there and did not change workspace free space. No full-disk failure occurred. Earlier-unit flat /tmp/u* files were not per-mutant cache directories; no repository or tools were removed.
* The audit and current required main differ. Scope expanded from 239 to 276 top-level tests, all 37 additions replayed. Requested rows are present at the same paths, with assertion bodies checked. Audit-only untrue means no audit-menu catch, not inability to fail: D1 demonstrates the omitted counter-width behavior.
* Two opt-in corpus/config rows and one contract-graph subcase skip, so their mutant responses remain unknown. Full completed whole-package runs do not make skipped checks pass. No claim of repo-wide uniqueness.
* Exclusive coverage blocks are leads, not independent assertions. The sound-neighbor defense rests on semantic input classification at shared refusal code. Coverage concerns Go lowerer/serializer code, not execution of emitted C.
* D2 changes a result constant and discards the unused binding. D3 is a guarded early return allowed by the menu, rather than an arbitrary inserted operation. These differences are generic semantic conditions, not test/fixture identity selectors.
* All three are defended, so no undefended-name finding is required. Still, RegExpNativeRefusals accepts either Refused or NotYet without pinning why, and does not execute native code. Its catch proves the specific overflow admission boundary. SuppressionDirectiveSoundNeighbors asserts nil error but never inspects a returned IR program, consistent with its audit empty-answer finding. Defense of prose admission does not repair that assertion limit. Representation checks six concrete/unresolved pairs and checker preconditions, not timing; its clock name does not establish a performance promise.
* Warm setup skipped. npm's reported 370ms is install time, not independently measured process wall. Mutation command wall includes Go compilation and any native builds. No isolated native rebuild time is claimed. Fetch and initial baseline compilation walls were not separately instrumented.

Timing: baseline binary 34.978s; D1 wall41.722s, binary33.491s; D2 wall43.107s, binary34.817s; D3 wall42.130s, binary34.409s. Standalone vet total1.193s. Coverage command walls total55.922s. Scope/coverage/mutants completed within the requested unit budget.

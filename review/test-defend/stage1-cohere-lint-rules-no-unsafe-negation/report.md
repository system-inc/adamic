TestCompileProfiles defended as a compilation-acceptance guard.
D1 is uniquely killed by the only test in the entire current package.
Evidence includes a compiler-valid standalone diff, clean coverage, full matrix and restored control.

Base origin/main: bc9edc5560379b63d4688a21006efb6032b60434. Audit ref fetched with full refspec; audit README, rows, plans, functions, matrix, runs, survivor and application-check notes were read. Current go test -list . contains only TestCompileProfiles, matching the audit. No newly gained test, vanished row, family, twin, or skip exists.

Before mutation, CODE UNDER TEST and ORACLE were stated in commentary and saved in plan.json. The audit tried semantic mutations of the .a lint port, whereas this test checks compilation only. This defense tests its asserted compile promise by mutating the production compiler itself. No oracle, harness or test was edited. The owned port remains well-typed and unchanged.

Coverage
The target covers 4823 Go blocks and 4814 blocks absent from the empty rest-of-package selection, totaling 6998 statements. There is no actual competing row, so this exclusivity is structural rather than evidence of rare coverage across the repository. Compiler coverage includes builtin's known-string-method recognition and stringCall argument lowering. No .a runtime coverage is claimed. The selected semantic lead is that the compile gate needs represented string codePointAt calls accepted, while all audit mutants preserved compileability.

Mutant
D1 internal/lower/object.go:1690 changes the existing stringMethods map key "codePointAt" to "codePointAtDisabled", a change-constant menu mutation. This removes recognition of a supported string operation. The unchanged scanner.ts dependency calls it at 69:20, and profile.a calls it at 37. The compiler now rejects the scanner call during Lower. This is an acceptance regression, not malformed Go source, a clang warning, or a changed expected answer.

Failure: build_test.go:28: /workspace/adamic/stage1/typescript/scanner/scanner.ts:69:20: Adamic 0.1 refuses inherited library member codePointAt read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)
Command: ADAMIC_BUILD_CACHE_DIR=/tmp/no-unsafe-negation-defense/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/rules/no-unsafe-negation/ -run . > D1.log 2>&1
Rows failed: TestCompileProfiles. Rows passed: [] because no other row exists. No bounded selection was necessary. The compiler mutant passed go vet ./internal/lower/ in 0.789 seconds, and the mutant test binary compiled and ran. D1's whole-package command took 13.811 wall seconds; test binary 0.204 seconds. It stops before native product rebuilds because lowering fails. The standalone diff applies to the untouched origin/main base. Production source was restored in a finally block.

Oracle and owner findings
Keep this row for compiler acceptance. This defense does not turn it into semantic lint verification: it never executes either native product or emitted JavaScript, and does not compare findings, messages, ranges, suggestions or edits. The audit's semantic survivors and empty-profile probe remain valid findings for that broader claim. Its name TestCompileProfiles accurately describes the native compile checks. Its log overstates emitted JavaScript as "compiled": the test only writes that file, without parsing or executing it. "Profiles" means the standalone lint profile program; there is no speed assertion or performance threshold, so this is not a cost row.

Unclear requirements and time costs
1. The audit confined mutations to the port while requiring every port mutation to compile. A compile-success row cannot reject such a mutation by construction. I expanded the explicitly named code under test to the production compiler whose acceptance the row actually asserts. The Go compiler mutation compiles with its own toolchain, and the valid unchanged input is rejected. This distinction is necessary to defend the row's actual promise.
2. Comparing coverage against "the rest of the package" is degenerate here: there is no other test. rest.cover comes from -run '^$', rather than inventing a subsumer or running unrelated packages. Repository-wide uniqueness is not asserted.
3. Go -coverpkg cannot instrument .a execution. The row has no port execution to instrument, so coverage is labeled compiler-pipeline coverage throughout.
4. Build and lowering times are mixed in this row; clean baseline and coverage report their own binary times rather than invented per-native-build timings. Warm env.sh worked, setup was skipped, nproc=5, npm ci stage3/api ran before baseline.
5. An execution-tool transport briefly disconnected during a read after coverage completed. Retrying the read worked; the test runs were unaffected.

Timings: clean baseline 37.121 binary seconds; clean target coverage 36.612 seconds; mutant 0.204 seconds plus compiler/test-binary build time noted above. Restored whole-package control passed in 39.672 seconds; see restored-control.log. Total session approximately seven minutes. No run exceeded 90 seconds. No other packages' tests, lint validation corpus, behavior probes or empty-answer probes were run in this defender task. One successful unique mutant was sufficient; no extra attempts were needed. No test deletion, rewriting, weakening, main push or pull request occurred.

Defense blocked at the pristine baseline, before any production mutation.
Starting main: 157a43552015f41a79331949c2e82b6f8c7caaab. All 17 family members remain present.
Verdict: cannot-judge, not a deletion recommendation.

CODE UNDER TEST: Adamic lowering, native/JavaScript emission and native runtime execution of the copied, already-mutated JSX canary. ORACLE: Node running that copied TypeScript. Go cohere output is obtained but disagreement is only logged by these shards; the separate MutantKilled witness asserts it.

Read CLAUDE.md and all prior audit report/rows/plan notes, then the family, product builder and native-emitter paths. Warm tools worked and setup was skipped. npm ci in stage3/api and lint registry generation preceded the baseline. Current full test enumeration is list.log; scope.json records all present family members and the package's current test count.

Baseline observations:
* Whole-package baseline timed out in TestLegacyMutants at 90 seconds, without an individual assertion fail event.
* First narrowed coverage build failed before testing because /tmp was full. It held scratch products from earlier units. No other unit's data was deleted; TMPDIR moved to /workspace/u113-defend-tmp, mode 1777.
* Pristine instrumented family run timed out at 90.075 binary seconds. load=0.161s and lower=4.631s completed; C emission did not. Its stack reached native.emit_objects.go dynamicProperties via repeated walkExpressions.
* Pristine family without instrumentation timed out at 90.059s. load=0.171s and lower=5.061s completed; C emission did not. Its stack reached fieldTypesNeeded via repeated walkExpressions. Coverage overhead alone therefore does not explain the block.
* TestNestedOutsideModuleCopy passed in 0.580s with coverage. This comparison row builds a small program, not the large lint graph.

Coverage: family.cover was flushed by the timed-out process and is partial construction coverage. other.cover is completed coverage of TestNestedOutsideModuleCopy, not coverage of every other package row. coverage-exclusive-blocks.txt contains leads relative to that one comparison row, not a uniqueness proof. Native/JavaScript/Lower were all in -coverpkg. No compiled canary assertion is claimed to have executed.

No mutants or standalone production diffs were created. The rule against defending without a clean baseline precludes interpreting any later mutant timeout or setup error as an assertion catch. Three honest production attempts were not possible, so not-defended would be unsupported. No tests, harnesses, or oracles were edited.

Owner finding: the family's name and mutant-detected log can suggest it proves the built-in JSX rule mutant differs from Go. Its leaf assertions instead prove emitted JavaScript and sanitized native output equal Node running that same mutated source. The Go-versus-mutated-Node disagreement is logged, not asserted here. The separate MutantKilled test supplies that check. This is a compiler-fidelity family; port-only survivor evidence does not prove it cannot fail under a compiler defect.

Brief friction/limits:
* A prior untrue audit was scoped to port mutants, whereas these leaves compare compiled products to the same mutated source on Node. The defense needs compiler/runtime mutations.
* A clean whole-package timeout permits narrowing, but the narrowed family's own shared setup remains monolithic: all setup aliases call yepestaFetch and rebuild the same native product. Selecting fewer leaves does not split C emission.
* /tmp exhaustion required another scratch filesystem and added a failed build log.
* Go coverage cannot instrument native C execution. The obtained family profile covers only construction up to the timeout.
* The package has gained many generated tests; current enumeration is saved. No mutant matrix was run, so no package or bounded uniqueness claim is made.
* Current origin/main is newer than the prior audit's ce1c5a2f. This report uses the exact starting commit above.
* The 90-second binary limit was retained for every test run. No attempt waited for a longer cold build to finish.

Existing source and tests are untouched. No tests in other packages ran. All test output was logged. rows.json is the requested per-row deliverable.

Additional healthy-port baseline TestShardsAgree_000 also exceeded 90 seconds; healthy-baseline.log records that result. It is not a clean comparator baseline and was not used for mutation conclusions.

Added a-check headers to slice 4 and removed obsolete generator refusal headers, preserving all program bodies.
Source 9655fccfbcaf22d6d16931e164a150338228d3ae; fixed base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA is in the handoff.
All 82 changed Adamic programs were compiled in place by this branch's own compiler: 61 checked, 5 NotYet, 16 refused, 0 type errors; all pass the pinned a-check predicate.
Removed-header and wrong-reason mutants on internal/lower/testdata/generic_body_relations/refused.a were each rejected by unchanged Gate.aCheck and restored.
Not covered: full-tree a-check, full gate, whole-package suites or a repeat of the previous commit's 49 compiler mutants. This pass tests header effects only.

The explicit fixed base takes precedence over the generic current-main instruction. After git fetch origin, created compiler/loom-r-slice-4-ac at dbdf977f5b and applied only 9655fccf's own change with git cherry-pick --no-commit. Main's newer tip was deliberately not used. No merge conflicts occurred. The final commit contains the original change plus this header pass; no compiler implementation or runtime change was added.

Inventory is git diff --name-only --diff-filter=AM dbdf977f 9655fccf -- '*.a', not a directory-limited scan. All 82 paths are present in statuses.json and statuses.md. There are no changed review/*.a files in that exact source patch; review evidence was included in the selection. Every path was compiled with adamic c in its actual tree location, before and after annotation. Diagnostics retain actual NotYet reasons separately from checked programs. Each compiler invocation has a 45-second limit; none timed out. build.log and scan.log record the build and scan; scan.py reproduces the procedure after rebuilding its compiler binary.

Four files gained their first header:

- internal/lower/testdata/generic_body_relations/refused.a
- stage3/fixtures/generics/11_indexed_result.a
- stage3/fixtures/generics/12_mixed_indexed_result.a
- stage3/fixtures/generics/13_constrained_local.a

Twelve existing refusal headers were expanded to the full actual diagnostic reason: notyet/generator_cycle.a; generics/06_polymorphic_recursion.a, 07_generic_cast.a, 08_readonly_view.a; iteration-dispatch/non_callable_next.a, non_object_close.a, non_object_iterator.a, non_object_step.a; iteration/array_view_variance.a, intrinsic_iterator_alias.a, intrinsic_iterator_prototype.a, intrinsic_iterator_write.a. Full paths, exact before/after first lines and unchanged body SHA-256 values are in headers.json.

The initial exact predicate also found two stale headers on clean programs. Removed '// a-check: refused a generator function' from stage3/fixtures/iteration/generator.a and delegated_generator.a, then compiled both again. Clean programs and NotYet stops now have no expected-error header. In total 18 first lines changed: four additions, twelve replacements and two removals. Every byte below the old/new a-check header is identical to 9655fccf; this was independently compared using git show for all 18 files. No program body was modified.

Position expectations changed only in internal/oracle/step16_generics_test.go's step16BodyRefusals table: 11_indexed_result.a 1:68 to 2:68; 12_mixed_indexed_result.a 1:87 to 2:87; 13_constrained_local.a 2:11 to 3:11. Exact-path and basename searches of *_test.go and testdata are recorded in path-readers.json. Existing historical review snapshots are retained as historical observations; statuses.json is the current compiler observation and does not relabel old observations as current. No other live diagnostic position expectation needed an update.

Predicate proof follows stage3/a-check-headers/README.md and its REPORT.md: fetched fbac28c62493f27a788edc02a18bc8edb68de5da, saved its exact cloud/fast-gate/run.py as gate.py, and invoked its unmodified Gate.aCheck using captured real compiler diagnostics. The build step is already completed separately; only process delivery is replaced by replay in the proof harness. Gate.aCheck still opens each real source and reads its actual first line. All 82 paths pass. Removing one real refusal header fails; replacing its reason with a string absent from the diagnostic fails. The program is restored in finally and verified byte-identical afterwards. predicate-results.json records the complete ordinary result and both rejection messages. This is a header-predicate proof, not a claim that a compiler was executed for each mutant; real compiler executions are the separate scan.

Validation used source /workspace/adamic-tools/env.sh, unset GOCACHEPROG, GOMAXPROCS=4. All test output went directly to logs. Commands:

- timeout 120 go build -o review/compiler/loom-r-slice-4-ac/adamic ./cmd/adamic
- timeout 300 python3 review/compiler/loom-r-slice-4-ac/scan.py
- timeout 30 python3 review/compiler/loom-r-slice-4-ac/predicate.py
- timeout 180 go test ./internal/lower ./internal/oracle -run '^(TestGenericBodyRelationsRefuse|TestGenericBodyRelationsWitness|TestStep16GenericOutcomes|TestStep16IdentityMutant|TestStep16OptionalResultMutant|TestGeneratorCycleFrame|TestIterationDispatchPendingStops|TestStep20ArrayViewVariance|TestStep20IntrinsicIteratorWrites)$' -count=1 -parallel=4 -timeout 90s -json
- timeout 150 go test ./internal/oracle -run '^TestStep20IterationOutcomes$' -count=1 -parallel=4 -timeout 90s -json, ADAMIC_GATE_UNCACHED=1
- timeout 150 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s -json
- timeout 150 go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/load ./internal/lower ./internal/native ./internal/oracle ./stage3/project-loader/probe

All required tests above and vet passed. Oracle path readers retain Node comparisons and native sanitizer/leak checks. TestStep20IterationOutcomes also checks JavaScript and release execution. TestCallTargetReaders passed in 2.38 seconds; the longest ordinary test was TestStep20IterationOutcomes at 10.12 seconds. test-times.json lists every observed test and seconds, including the infrastructure failure below. No test was removed, weakened, split or rerun after exceeding 60 seconds; none exceeded 60 seconds.

Counts: internal/oracle/counts.md is byte-identical to 9655fccf. No regeneration was performed. An extra focused TestCountsAreRecorded family invocation failed because its shard preparation filled the 32 GB filesystem (13 GB Go cache and 13 GB product cache); it printed 'no space left on device', followed by missing measurement messages. That run is retained in counts-check.jsonl and marked as infrastructure failure in test-times.json. It is not reported as a passing count-family test. To work around this without changing code or counts, moved this pass's disposable compiler to /tmp, removed four verified untracked ELF binaries from my previous pass's review scratch, and built the two affected generators directly with --count using scratch for cache products. Both exact recorded rows passed: generator.a 14/14/33/35/13/0; delegated_generator.a 20/20/70/74/15/0. Direct build-and-run seconds were 4.447 and 1.019. direct-counts.json records commands and measured/recorded values. No cache source, tracked artifact or evidence was deleted.

Setup: GOPROXY=https://proxy.golang.org|direct; ADAMIC_GOCACHE_OFF=1 timeout 600 bash cloud/setup.sh succeeded. Timing lines in setup.log: node .029s, Go .044s, submodules .060s, markdown .091s (step .012s), clang .218s, cache off .220s, Go build 11.373s, test binaries deferred 11.541s, cache warm 11.543s, done 11.569s. nproc=5, cgroup quota four CPUs; Go 1.27.1, Node 24.19.0, house clang 20.1.8. Setup itself needed no workaround. No cohere code was copied, no PR was opened and Loom was not invoked.

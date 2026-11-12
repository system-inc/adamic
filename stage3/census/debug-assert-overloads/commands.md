# Task 68x1t2y: reproduction

Main is pinned at `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`, detached in `/workspace/orm-main`. The report branch starts at `7d73b282da946b85a73d3e77b2aea0e3f891f30d`. No new fetch or scratch merge changes the measured main. `/workspace/orm-adapted` is the identical adapted tree from the preceding census, and all 747 hashes are rechecked. See the preceding `overload-results-main` directory for the own-tree overlay/build and immutable upstream TypeScript source pins.

Every potentially long child below is wrapped with `run.py`: terminate at 1,740 seconds, then at most ten seconds to kill. It prints and propagates its child's actual status. Shell wrappers print and propagate that status. Go evidence files end in `.go.txt`. Actual Go scratch inputs and executable live outside this directory.

Set:

```sh
source /workspace/adamic-tools/env.sh
scripts=/workspace/op2-report/stage3/census/debug-assert-overloads
```

1. Run `bash "$scripts/measure.sh"`. It clears all inherited LATENT_/CENSUS_ variables, sets FULL=1, ONLY_FILE to adapted destructuring.ts, ASSERT_NO_OUTPUT=1, and uses the original entry-census binary with no shard selection. The entry root remains adapted src/tsc/tsc.ts. It retains the full JSONL stream, all attempted-unit timings, child log and resources. Full mode yields nested units in addition to outer units; do not compare its unit count to first-error mode's top-level catalog.

2. Run the checker:

```sh
python3 "$scripts/run.py" "$scripts/checker.log" "$scripts/checker.time.json" node "$scripts/checker.cjs"
rc=$?
echo "checker child_exit=$rc"
exit "$rc"
```

This is a standalone command block, not the parent of subsequent work. The script requires stock TypeScript 6.0.3. It uses the same strict compiler options and entry root plus main's bundled Node 25.3.3 index/type root, keeping the source files unchanged. It records zero diagnostics, all six debug assertion calls in destructuring.ts, declarations, predicates, source hashes, exhaustive any-reachability with the production library-member exemption, and a first-any witness search with library members included. The earlier exploratory full-library enumeration stopped at 100,000 types; its nonzero log/statistics are retained separately and it does not supply report counts.

3. Build the scratch probe:

```sh
python3 -B "$scripts/instrument.py"
# In /workspace/orm-main:
python3 "$scripts/run.py" "$scripts/probe-build.log" "$scripts/probe-build.time.json" go build -buildvcs=false -overlay=/workspace/dao-probe/overlay.json -o /workspace/dao-probe/entry-census ./stage3/census/latent/tool
rc=$?
echo "probe build child_exit=$rc"
exit "$rc"
```

`instrument.py` retains the exact generated Go as `.go.txt`. It overlays two production files only in scratch, layered over the previously built entry overlay. It preserves the walk's nil/seen checks, any-or-bound condition, original traversal order, library exclusion and returns. A helper carries an immutable copy of each edge path; every successful any/bound event prints branch, flags, depth and all 33 node labels. Outer parameter checks and fallback doors are logged with their mapped types. The final CALLGATE logs whether holdsAny short-circuited the relation; `relatedIfEvaluated=false` is not a measured false relation when holdsAny=true.

4. Run `bash "$scripts/probe-run.sh"`. This uses exactly the same FULL/ONLY_FILE/ASSERT_NO_OUTPUT settings as the stock run. All stock and instrumented stream records are identical, including the relation report. Tracing does not change the observed answer.

5. In measured main, check the existing small twins:

```sh
python3 "$scripts/run.py" "$scripts/small-shape-tests.log" "$scripts/small-shape-tests.time.json" go test ./internal/lower -run '^TestOverloadCallback(AssertNode|AssertEachNode)$' -count=1
rc=$?
echo "small shape tests child_exit=$rc"
exit "$rc"
```

Both tests pass. Their retained original source documents the shallow N/Id types and the omitted NodeArray/stackCrawlMark pieces. These are the repository's test programs, not a claim to have the separate unpublished Optimizer sample.

6. Generate/recheck the report and evidence:

```sh
python3 "$scripts/run.py" "$scripts/report.log" "$scripts/report.time.json" python3 -B "$scripts/report.py"
rc=$?
echo "report child_exit=$rc"
exit "$rc"
```

`report.py` asserts the exact main, pristine tracked production sources, all 747 adapted hashes, checker source hashes, zero diagnostics, stock/probe stream equality, exactly eight selected overload-parameter observations, twelve bound events with 33-node paths and no Any flag, and zero/time-bounded exits for every final child. It derives `overload-parameters.json`, `depth-traces.json`, `validation.json`, provenance and REPORT.md. The report shows all eight observations and all four unique depth paths for the two requested calls.

Only this directory is committed and pushed to typescript/overload-proof-census. No binaries, production changes, or adapted-source changes are committed.

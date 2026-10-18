Built: loom-r-string-brand-ac-m, preserving current main and applying the source tip's own change as one commit.
Commits: source 539788fef585ab910f78f5df41463ccb48902e0b; base 5a94465370ea0172e25aae8c13478e59e40c2bb4; delivery SHA in final response.
Commands and outputs: final whole touched-package suites, counts regeneration, reader/map guard, parallel guard and vet passed; initial failures and exact argv/exits are retained in checks.json.
Mutants: current-tree lowering and/or a-check header proofs are recorded below; no compiler-build failure counts as a catcher.
Not covered: full repository gate, ahra, Loom submission or opt-in lanes that the package suites skipped (skips.json).

Fetched main and source refs. The newest fetched origin/main at branch creation is the base above. Applied only the source tip's own diff using git cherry-pick --no-commit, and independently captured the requested explicit merge-base merge-tree result. No diff against the source branch's whole history was applied. All main work outside the branch's required changes is retained.

Conflict: internal/oracle/counts.md, the insertion after static_constructor_spread_owned.a. Main added string_destructuring.a; the branch added six string_brand_ rows at the same location. Replaced the conflicted file with main's table and regenerated it once. The resulting diff adds precisely the six original brand rows and retains main's string_destructuring.a row unchanged. All six values match the source. No manual row merge.

Main's newer string destructuring and implied-target cache code stay intact. The source's Go, fixture and docs added/deleted lines are identical on this base (source-delta.json). No main line was reverted. There was no compiler-code conflict or manual lowering fix. Merge-tree and conflicts.patch evidence preserve the original conflict. Runtime source and recut diffs are both empty.

All 12 source-added .a files, including the review void-view probe, are scanned with the current branch compiler by the unchanged README-pinned Gate.aCheck. The five refusal headers remain intact. Removed-header and wrong-reason mutants each fail the predicate, and the original programs are restored. The original five lowering mutants are regenerated as overlays over current-main files, so none replaces a file with an older-main snapshot.


The first complete oracle run found a main expectation that this lowering intentionally advances: internal/oracle/element_access_boundaries_test.go's paths case expected NotYet for a function returning Path | undefined. The branch now proves that representation; the branded array still has no element-access lowering. Updated only that expectation to the exact remaining NotYet, an ElementAccessExpression, retaining the Node witness and all refusal checks. The original failure is in oracle-all.log; oracle-all-fixed.log is the complete rerun. This is the only extra source/test change beyond the original commit and regenerated counts. It is required by the branch's lowering, not a reversion of main's infrastructure. A current-source compiler mutant disables only the string-brand representation hook and is caught by this paths boundary check when the old representation stop returns.


Counts: took the complete main table, never hand-merged it, then ran the full TestCountsAreRecorded -update-counts on Linux with house clang 20.1.8 after setup compiled the resolved tree. counts.patch and counts.json list the entire delta. Only this change's rows moved; every other main row is preserved. The precise added/deleted lines are below.

Runtime clearance: runtime.json compares source-own and recut added/deleted lines under internal/native/runtime. Additional resolution lines in runtime *.c or *.h: NONE. No runtime approval is needed for an extra resolution line.

Validation: whole-package go test commands are in checks.json. Each uses GOMAXPROCS=4, -parallel 4, -p 1, -count=1, -timeout 12m and JSON output, with an outer per-command hard deadline. Focused guards use -timeout 90s. ADAMIC_GATE_UNCACHED=1 bypasses oracle observation caches. Every requested test command actually ran; no tests were removed. Output goes directly to evidence logs. All measured leaf tests are below 60 seconds; test-times.json/tsv retains every leaf and grouping parent, including failures intentionally produced by mutants. Native/oracle suites include Node agreement, sanitizers, leaks and runtime witnesses. gofmt and source whitespace checks pass. Vet covers the changed compiler packages. Opt-in skips are preserved and explicitly listed, not claimed covered.

Timing detail: on this worker's first native run, TestSplitTSGoAgreesUnit00 was a 163.58s grouping parent because it built the shared checker C archive before its child comparisons. Its child leaves were under 60s. The initial collector mistakenly classified that grouping parent as a leaf; corrected the classification and retained both observations. The whole native suite exited zero. The group was verified again with its products prepared, matching Loom's product ordering; native-split-after-products in checks.json records the successful bounded run. No test was deleted, weakened or given a larger leaf budget.

Setup: GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh under a 600-second hard limit; sourced /workspace/adamic-tools/env.sh. nproc=5 and cpu.max=400000 100000 (four CPUs). House clang 20.1.8, Go 1.27.1 and Node v24.19.0. Timing lines:

```text
setup: node ready (0.024s)
setup: go ready (0.028s)
setup: submodules ready (0.053s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.010s
setup: markdown dependencies ready (0.080s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.208s)
setup: shared cache ready (0.915s)
setup: go build ready (47.838s)
setup: test binaries deferred (use --warm-tests) (47.910s)
setup: build cache warm (47.911s)
setup: build-flags commit=5a94465370ea0172e25aae8c13478e59e40c2bb4 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.06 0.01 0.00 1/140 1082 load-after=2.80 0.75 0.25 1/146 1964
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (47.935s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.ozQ9Qt
```

Assumptions: use the latest fetched origin/main separately for each branch; retain main's newer infrastructure and apply only the source-own change. The newest request explicitly requires whole touched-package suites, superseding the earlier focused-only worker instruction. No cohere code was copied. No PR, ahra command or Loom submission was made. One commit is pushed only to this new own branch, without force. The commit message preserves the original message and adds the requested re-cut line and Ahra co-author trailer.

Commands and observed outcomes:

- counts-regenerate: exit 0, 84.357s.
- lower-all: exit 0, 137.777s.
- native-all: exit 0, 416.252s.
- native-split-after-products: exit 0, 8.376s.
- oracle-all: exit 1, 427.796s.
- oracle-all-fixed: exit 0, 340.614s.
- readers-maps: exit 0, 35.653s.
- parallel-guard: exit 0, 2.222s.
- vet: exit 0, 7.5s.
- mutant-real-field-phantom: exit 1, 13.189s.
- mutant-drop-undefined: exit 1, 16.105s.
- mutant-read-brand: exit 1, 13.202s.
- mutant-literal-refinement: exit 1, 9.916s.
- mutant-unchecked-missing-cast: exit 1, 9.819s.
- mutant-paths-representation-regression: exit 1, 16.043s.

Counts added lines:

```text
| internal/oracle/testdata/string_brand_escape.a | 14 | 14 | 38 | 43 | 5 | 0 |
| internal/oracle/testdata/string_brand_unescape.a | 7 | 7 | 42 | 39 | 4 | 0 |
| internal/oracle/testdata/string_brand_map.a | 9 | 9 | 23 | 28 | 7 | 0 |
| internal/oracle/testdata/string_brand_optional.a | 10 | 10 | 27 | 38 | 4 | 0 |
| internal/oracle/testdata/string_brand_plain.a | 3 | 3 | 4 | 7 | 3 | 0 |
| internal/oracle/testdata/string_brand_value.a | 3 | 3 | 9 | 12 | 3 | 0 |
```

Counts removed lines:

```text

```

Lowering mutant catchers:

- mutant-drop-undefined.log: TestStringBrandOptional, 2.92s.
- mutant-literal-refinement.log: TestStringBrandCastKeepsLiterals, 0.03s.
- mutant-paths-representation-regression.log: TestElementAccessCompilerAreaBoundaries/paths, 2.6s.
- mutant-paths-representation-regression.log: TestElementAccessCompilerAreaBoundaries, 0s.
- mutant-read-brand.log: TestStringBrandReadRefused, 0.03s.
- mutant-real-field-phantom.log: TestStringBrandRealRefused, 0.03s.
- mutant-unchecked-missing-cast.log: TestStringBrandCastChecksMissing, 0.03s.

Header mutant results:

```json
[
  {
    "file": "internal/oracle/testdata/string_brand_refused/destructure.a",
    "mutation": "removed-header",
    "exit": 1,
    "caught_by": "unchanged README-pinned Gate.aCheck with current branch compiler",
    "failure": "a-check failed for 1 file:\ninternal/oracle/testdata/string_brand_refused/destructure.a: refused, expected checked (adamic: /workspace/adamic/internal/oracle/testdata/string_brand_refused/destructure.a:2:48: Adamic 0.1 refuses a string brand member __brand read at runtime; keep the brand phantom and unread, or use )"
  },
  {
    "file": "internal/oracle/testdata/string_brand_refused/destructure.a",
    "mutation": "wrong-reason",
    "exit": 1,
    "caught_by": "unchanged README-pinned Gate.aCheck with current branch compiler",
    "failure": "a-check failed for 1 file:\ninternal/oracle/testdata/string_brand_refused/destructure.a: refused, expected refused deliberately unmatched brand reason (adamic: /workspace/adamic/internal/oracle/testdata/string_brand_refused/destructure.a:3:48: Adamic 0.1 refuses a string brand member __brand read at runtime; keep the brand phantom and unread, or use )"
  }
]
```

Measured leaf count: 11757; maximum observed leaf seconds: 23.85.

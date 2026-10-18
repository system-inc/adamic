Built: loom-r-string-brand-ac-m-fixed, preserving current main and applying the source tip's own change as one commit.
Commits: source 539788fef585ab910f78f5df41463ccb48902e0b; base 9f7c0b5bbe75e8a1593cd3c5d104a49289d1486e; delivery SHA in final response.
Commands and outputs: final whole touched-package suites, counts regeneration, reader/map guard, parallel guard and vet passed; exact argv/exits are retained in checks.json.
Mutants: current-tree lowering and/or a-check header proofs are recorded below; no compiler-build failure counts as a catcher.
Not covered: full repository gate, ahra, Loom submission or opt-in lanes that the package suites skipped (skips.json).

Fetched main and source refs. The newest fetched origin/main at branch creation is the base above. Applied only the source tip's own diff using git cherry-pick --no-commit, and independently captured the requested explicit merge-base merge-tree result. No diff against the source branch's whole history was applied. All main work outside the branch's required changes is retained.

Conflict: internal/oracle/counts.md only. Main added string-destructuring observations and other newer table rows. Took the complete main table, then regenerated once with house clang 20.1.8; all existing main rows remain byte-identical and only the six source string-brand rows are added. No production-code conflict or runtime resolution was needed.

Main consumer expectation: internal/oracle/element_access_boundaries_test.go still expected Path | undefined return lowering to stop. This change proves that representation and gets as far as unsupported branded array indexing. Preserved Node validation and changed the precise expected stop to an ElementAccessExpression; a current-tree representation mutant proves this expectation can fail.

Full stage3 consumer audit found three stale main records, all in unchanged program files. assertions/17_brand_upcast.a still refuses, now at 11:73 for a non-phantom __escapedIdentifier runtime layout, replacing its old cast-stop diagnostic at 14:70. namespaces/02_jsx_names.a and 03_react_names.a now compile; TestFixturesNamespaces -update admitted these only after native output with sanitizers/leaks equaled recorded Node output. Exactly these three stage0 observations changed; Node records and program bodies are unchanged. A stale-diagnostic mutant and a stale-refusal mutant test those records. Full stage3/fixtures/... passed.

Delivery assumption: the earlier compiler/loom-r-string-brand-ac-m missed these stage3 consumers and is superseded. Use compiler/loom-r-string-brand-ac-m-fixed for the complete one-commit replacement without rewriting any pushed history. Base is the newest origin/main fetched when this replacement started. Chain-followers uses its independently validated compiler/loom-r-chain-followers-ac-m-fixed.

All source-added/modified .a programs, including review evidence, retain their existing correct a-check headers and unchanged bodies. The actual compiler and pinned README predicate are rerun, including removed-header and wrong-reason mutants. No pre-existing unrelated main program was rewritten.

Guards: full lower tests include native-agreement program registration; full oracle includes exact runtime-set witnesses. Explicit IR reader/SortedMap and cmd/adamic-gate parallel guards passed. No new compiler-consuming Go package or IR-held map is introduced, so dependency declarations and map types need no changes. No production/test recipe adds a build-directory key. All original source .a bytes match after header mutants were restored.

Environment: placed new products and Go-cache writes on the separate temporary volume using main's supported ADAMIC_BUILD_OUTPUT_DIR and ADAMIC_GOCACHE_DIR. Preserved the existing ordinary Go cache by relocating it to /tmp/adamic-gate/standard-go-cache-preserved and symlinking its original path when the root volume approached full; no cached contents were deleted. These are environment settings, not source or product-key changes.

Final proofs: all six lowering mutants failed with their named test catchers, both stale stage3 record mutants failed the exact gap comparison, and restored stage3 records passed. All 12 changed .a programs passed the actual branch compiler and pinned a-check predicate: seven checked, five refused, zero NotYet, zero type errors. Both removed-header and wrong-reason mutants failed that predicate, and all source program bytes were restored.


Counts: took the complete main table, never hand-merged it, then ran the full TestCountsAreRecorded -update-counts on Linux with house clang 20.1.8 after setup compiled the resolved tree. counts.patch and counts.json list the entire delta. Only this change's rows moved; every other main row is preserved. The precise added/deleted lines are below.

Runtime clearance: runtime.json compares source-own and recut added/deleted lines under internal/native/runtime. Additional resolution lines in runtime *.c or *.h: NONE. No runtime approval is needed for an extra resolution line.

Validation: whole-package go test commands are in checks.json. Each uses GOMAXPROCS=4, -parallel 4, -p 1, -count=1, -timeout 12m and JSON output, with an outer per-command hard deadline. Focused guards use -timeout 90s. ADAMIC_GATE_UNCACHED=1 bypasses oracle observation caches. Every requested test command actually ran; no tests were removed. Output goes directly to evidence logs. All measured leaf tests are below 60 seconds; test-times.json/tsv retains every leaf and grouping parent, including failures intentionally produced by mutants. Native/oracle suites include Node agreement, sanitizers, leaks and runtime witnesses. gofmt and source whitespace checks pass. Vet covers the changed compiler packages. Opt-in skips are preserved and explicitly listed, not claimed covered.

Setup: GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh under a 600-second hard limit; sourced /workspace/adamic-tools/env.sh. nproc=5 and cpu.max=400000 100000 (four CPUs). House clang 20.1.8, Go 1.27.1 and Node v24.19.0. Timing lines:

```text
setup: node ready (0.018s)
setup: go ready (0.021s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.006s
setup: markdown dependencies ready (0.060s)
setup: submodules ready (0.061s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.156s)
setup: shared cache ready (9.116s)
setup: go build ready (30.237s)
setup: test binaries deferred (use --warm-tests) (30.322s)
setup: build cache warm (30.324s)
setup: build-flags commit=9f7c0b5bbe75e8a1593cd3c5d104a49289d1486e nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.31 1.32 1/515 307224 load-after=0.95 0.54 1.36 1/521 309363
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (30.350s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.5Kxdo2
```

Assumptions: use the latest fetched origin/main separately for each branch; retain main's newer infrastructure and apply only the source-own change. The newest request explicitly requires whole touched-package suites, superseding the earlier focused-only worker instruction. No cohere code was copied. No PR, ahra command or Loom submission was made. One commit is pushed only to this new own branch, without force. The commit message preserves the original message and adds the requested re-cut line and Ahra co-author trailer.

Commands and observed outcomes:

- stage3-update: exit 0, 25.999s.
- stage3-all: exit 0, 49.612s.
- counts-regenerate: exit 0, 49.604s.
- lower-all: exit 0, 42.867s.
- native-all: exit 0, 181.668s.
- oracle-all: exit 0, 312.19s.
- readers-maps: exit 0, 6.85s.
- parallel-guard: exit 0, 1.167s.
- vet: exit 0, 1.627s.
- mutant-real-field-phantom: exit 1, 3.33s.
- mutant-drop-undefined: exit 1, 6.192s.
- mutant-read-brand: exit 1, 3.387s.
- mutant-literal-refinement: exit 1, 3.523s.
- mutant-unchecked-missing-cast: exit 1, 3.677s.
- mutant-paths-representation-regression: exit 1, 6.735s.
- mutant-record-assertions: exit 1, 5.893s.
- mutant-record-namespaces: exit 1, 5.744s.
- stage3-records-restored: exit 0, 11.922s.

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

- mutant-drop-undefined.log: TestStringBrandOptional, 2.91s.
- mutant-literal-refinement.log: TestStringBrandCastKeepsLiterals, 0.03s.
- mutant-paths-representation-regression.log: TestElementAccessCompilerAreaBoundaries/paths, 2.76s.
- mutant-paths-representation-regression.log: TestElementAccessCompilerAreaBoundaries, 0s.
- mutant-read-brand.log: TestStringBrandReadRefused, 0.03s.
- mutant-real-field-phantom.log: TestStringBrandRealRefused, 0.03s.
- mutant-record-assertions.log: TestFixturesAssertions/assertions/17_brand_upcast.a/stage0, 0s.
- mutant-record-assertions.log: TestFixturesAssertions/assertions/17_brand_upcast.a, 0.09s.
- mutant-record-assertions.log: TestFixturesAssertions/assertions, 0s.
- mutant-record-assertions.log: TestFixturesAssertions, 3.04s.
- mutant-record-namespaces.log: TestFixturesNamespaces/namespaces/02_jsx_names.a/stage0, 0s.
- mutant-record-namespaces.log: TestFixturesNamespaces/namespaces/02_jsx_names.a, 0.09s.
- mutant-record-namespaces.log: TestFixturesNamespaces/namespaces, 0s.
- mutant-record-namespaces.log: TestFixturesNamespaces, 2.9s.
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

Measured leaf count: 8065; maximum observed leaf seconds: 16.72.

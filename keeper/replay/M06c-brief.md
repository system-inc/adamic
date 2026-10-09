New work for you, from @system_adamic_tests (the keeper of Adamic's test audit, #xphstyt): a mutant replay, not an audit. You run one planted mutant against a named set of packages and report exactly which tests fail. Nothing gets judged here; the keeper reads the failing set.

Before anything else: your workspace is warm and keeps earlier units' build caches. Run `df -h /tmp /workspace`; if either has under 15 GB free, delete earlier units' scratch and per-mutant cache directories under /tmp (never /workspace/adamic or your tools).

## The replay

- Base: `bc9edc5560379b63d4688a21006efb6032b60434`. Fetch it and work on a detached checkout of exactly that sha, clean (`git status` empty), with submodules at its pins. Skip setup if /workspace/adamic-tools/env.sh works.
- The mutant, `u045 M06 (element_borrow.go:276, captured roots may borrow), covered packages only` (diff hash e742af58576a), applied with `git apply` at the repository root:

```diff
--- a/internal/native/element_borrow.go
+++ b/internal/native/element_borrow.go
@@ -273,7 +273,7 @@
 		return false
 	}
 	held := program.Locals[root]
-	if held.Global || held.Captured || held.Function != function || assigned[root] {
+	if held.Global || held.Function != function || assigned[root] {
 		return false
 	}
 	return chainUnchanged(program, program.Functions[function].Body, names)
```

- Packages (run every one, whole, nothing else):
  - `./internal/native`
  - `./internal/oracle`
  - `./stage1/cohere/css`
  - `./stage1/cohere/estree`
  - `./stage1/cohere/formatfiles`
  - `./stage1/cohere/graphql`
  - `./stage1/cohere/graphql/printer`
  - `./stage1/cohere/json`
  - `./stage1/cohere/lint`
  - `./stage1/cohere/lint/helpers/comments`
  - `./stage1/cohere/lint/rules/no-unsafe-negation`
  - `./stage1/cohere/lint/rules/no-unsafe-optional-chaining`
  - `./stage1/cohere/lint/rules/typescript-no-this-alias`
  - `./stage1/cohere/markdownblocks`
  - `./stage1/cohere/mediaquery`
  - `./stage1/cohere/values`
  - `./stage1/cohere/yaml`
  - `./stage1/typescript/parser`

Run them with the mutant applied, each mutant run on its own fresh caches so no earlier result stands in:

    export ADAMIC_BUILD_CACHE_DIR=/tmp/replay-e742af58576a/cache ADAMIC_GATE_UNCACHED=1
    go test -json -count=1 -timeout 30m <the packages> > /tmp/replay-e742af58576a/mutant.jsonl 2>&1

(One go test over all the packages is fine; it runs them in parallel. If it would take longer than 40 minutes, split by package and say so.)

Then, for every test that failed under the mutant, rerun just those tests on the clean base (`git apply -R`, a fresh cache directory) to show they pass without it. A test that fails on the clean base too is not a catch; list it separately as red at base.

## Report

One short paragraph: the diff applied or didn't, wall times, anything that broke. Then exactly one fenced json block:

```json
{"mutant": "u045 M06 (element_borrow.go:276, captured roots may borrow), covered packages only", "hash": "e742af58576a", "base": "bc9edc5560379b63d4688a21006efb6032b60434", "applied": true,
 "packages_run": ["..."], "packages_failed_to_build": ["..."],
 "caught_by": [{"package": "internal/oracle", "test": "TestNativeAgreesWithNode", "subtest": "internal/oracle/testdata/x.a", "line": "file_test.go:123: the failure's first line"}],
 "red_at_base": [{"package": "...", "test": "..."}],
 "survived": false}
```

`caught_by` lists top-level tests that fail with the mutant and pass on the clean base, one entry per top-level test (name the first failing subtest). `survived` is true only when caught_by is empty and every package built and ran to completion. A package that didn't finish (timeout, killed, out of disk) goes in a note and makes `survived` false: the keeper can't call a mutant unguarded on a partial run. Restore the tree when done. Push nothing.


New work for you, from @system_adamic_tests (the keeper of Adamic's test audit, #xphstyt): a mutant replay, not an audit. You run one planted mutant against a named set of packages and report exactly which tests fail. Nothing gets judged here; the keeper reads the failing set.

Before anything else: your workspace is warm and keeps earlier units' build caches. Run `df -h /tmp /workspace`; if either has under 15 GB free, delete earlier units' scratch and per-mutant cache directories under /tmp (never /workspace/adamic or your tools).

## The replay

- Base: `bc9edc5560379b63d4688a21006efb6032b60434`. Fetch it and work on a detached checkout of exactly that sha, clean (`git status` empty), with submodules at its pins. Skip setup if /workspace/adamic-tools/env.sh works.
- The mutant, `u045 M11 (arguments_length.go:179, packed closure count + 1)` (diff hash 94c965473851), applied with `git apply` at the repository root:

```diff
--- a/internal/native/arguments_length.go
+++ b/internal/native/arguments_length.go
@@ -176,7 +176,7 @@
 
 func (e *emitter) packedClosureCall(call ir.CallClosure, closure, packed, count string) string {
 	if e.program.PackedCountNeededFromCall(call) {
-		return fmt.Sprintf("adamic_closure_call(%s, %s, %s)", closure, packed, count)
+		return fmt.Sprintf("adamic_closure_call(%s, %s, %s)", closure, packed, count+" + 1")
 	}
 	return fmt.Sprintf("%s->code(%s, %s)", closure, closure, packed)
 }
```

- Packages (run every one, whole, nothing else):
  - `./internal/native`
  - `./internal/oracle`
  - `./stage1/cohere/estree`
  - `./stage1/cohere/lint`
  - `./stage1/cohere/lint/helpers/comments`
  - `./stage1/cohere/lint/rules/no-unsafe-negation`
  - `./stage1/cohere/lint/rules/no-unsafe-optional-chaining`
  - `./stage1/cohere/lint/rules/typescript-no-this-alias`
  - `./stage1/typescript/parser`
  - `./cmd/adamic`
  - `./cmd/adamic-test262`
  - `./internal/fuzz`
  - `./internal/lower`
  - `./stage1/cohere/css`
  - `./stage1/cohere/cssstrings`
  - `./stage1/cohere/graphql`
  - `./stage1/cohere/graphql/printer`
  - `./stage1/cohere/json`
  - `./stage1/cohere/lint/helpers`
  - `./stage1/cohere/lint/regex`
  - `./stage1/cohere/lint/rules/no-underscore-dangle`
  - `./stage1/cohere/markdownblocks`
  - `./stage1/cohere/selector`
  - `./stage1/cohere/static_single_assignment`
  - `./stage1/cohere/tsprinter`
  - `./stage1/cohere/typeaware`
  - `./stage1/cohere/yaml`
  - `./stage1/typescript/scanner`

Run them with the mutant applied, each mutant run on its own fresh caches so no earlier result stands in:

    export ADAMIC_BUILD_CACHE_DIR=/tmp/replay-94c965473851/cache ADAMIC_GATE_UNCACHED=1
    go test -json -count=1 -timeout 30m <the packages> > /tmp/replay-94c965473851/mutant.jsonl 2>&1

(One go test over all the packages is fine; it runs them in parallel. If it would take longer than 40 minutes, split by package and say so.)

Then, for every test that failed under the mutant, rerun just those tests on the clean base (`git apply -R`, a fresh cache directory) to show they pass without it. A test that fails on the clean base too is not a catch; list it separately as red at base.

## Report

One short paragraph: the diff applied or didn't, wall times, anything that broke. Then exactly one fenced json block:

```json
{"mutant": "u045 M11 (arguments_length.go:179, packed closure count + 1)", "hash": "94c965473851", "base": "bc9edc5560379b63d4688a21006efb6032b60434", "applied": true,
 "packages_run": ["..."], "packages_failed_to_build": ["..."],
 "caught_by": [{"package": "internal/oracle", "test": "TestNativeAgreesWithNode", "subtest": "internal/oracle/testdata/x.a", "line": "file_test.go:123: the failure's first line"}],
 "red_at_base": [{"package": "...", "test": "..."}],
 "survived": false}
```

`caught_by` lists top-level tests that fail with the mutant and pass on the clean base, one entry per top-level test (name the first failing subtest). `survived` is true only when caught_by is empty and every package built and ran to completion. A package that didn't finish (timeout, killed, out of disk) goes in a note and makes `survived` false: the keeper can't call a mutant unguarded on a partial run. Restore the tree when done. Push nothing.


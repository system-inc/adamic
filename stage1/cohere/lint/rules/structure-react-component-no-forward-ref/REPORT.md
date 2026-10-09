Port of `structure/react-component-no-forward-ref` for Kirk and Ahra. Only this rule directory changes. The adapter returns the pinned Go rule unchanged; the rule has no options, fixes or suggestions.

The rule reports React named imports by imported name, including aliases; bare calls named `forwardRef`; `React.forwardRef` calls; and `React.ForwardRefExoticComponent` or `React.ForwardedRef` qualified types. The upstream case "a bare call with no import in the file" deliberately reports a locally declared function of that name. The port preserves that behavior. An aliased import reports the import while its differently named call stays silent.

The shared `bindingsOf`, `importedNameOf`, `isNamespacedMember` and React `project` helpers perform their existing operations. The owned import projection supplies named parser links. Registration's prepare hook builds both projections once per file. A compiler refusal during the initial projection attempt was isolated with an unchanged-base control, which passed; using the shared React projection and existing parser child arrays resolved it without changing shared code.

Validation commands, from the repository root with `/workspace/adamic-tools/env.sh` sourced:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/no-forward-ref-setup.log 2>&1
go run ./cmd/lint-registry > /tmp/no-forward-ref-registry.log 2>&1
ADAMIC_LINT_RULES=structure-react-component-no-forward-ref GOMAXPROCS=4 go test -overlay=/tmp/no-forward-ref-overlay.json ./stage1/cohere/lint -run 'TestRulesAgree|TestOwnedWitnesses|TestMutants/structure-forward-ref-import-module|TestForwardRef' -count=1 -v -timeout=40m > /tmp/no-forward-ref-selected-final.log 2>&1
```

The overlay maps the virtual `stage1/cohere/lint/forward_ref_validation_test.go` to this directory's `validation.go.txt`, both as absolute paths, using Go's `{"Replace": {...}}` format. It reads the guard fixture strings from the unchanged upstream Go AST and replays all 11 shapes at each of the guard's three filenames. Its two top-level tests call `t.Parallel`. It also counts the 17 captured rule cases and builds the owned semantic mutant under ASan and UBSan, requiring its output to disagree with Go and agree with mutated Node.

Selected validation passed in 91.413 seconds: 17 captured rule cases, 33 upstream absent-node cases, three owned witnesses, and the inherited generated corpus. Source Node, emitted JavaScript and sanitized native all matched the independent Go output byte for byte. The whole discovery capture contained 5,747 unique source/rule/options combinations; the scoped rule-case comparator produced 505,841 identical output bytes. Every owned witness produced a Go finding. `testdata/imports.tsx.txt` covers named aliases and member calls, `testdata/types.ts.txt` covers both qualified types, and `testdata/parentheses.ts.txt` covers parenthesized bare and member calls.

The mutant changes the import module gate from `react` to `other`, compiles and executes, and loses the aliased import finding. All three backends caught that semantic disagreement. See `evidence/mutant.log` for exact catch lines and `evidence/selected.log` for the successful comparisons and mutant execution.

Setup passed in 0.943 seconds on five reported processors, with a four-CPU cgroup quota. The complete step timing and load observations are in `evidence/setup.log`; registration output is in `evidence/registry.log`, and the unchanged-base control is in `evidence/baseline.log`.

Whole-package measurements are recorded in `evidence/full-metrics.json`; the complete JSON test stream is compressed without changing its bytes in `evidence/full.jsonl.gz`. The required TypeScript checkout is clean at `050880ce59e30b356b686bd3144efe24f875ebc8`, the WASI sysroot is supplied, benchmarking is enabled, and both profile variables point to the same newly created directory. No scoped-rule filter is set for the whole package.

The whole lint package passed: **159 passes, zero failures, one skip**, including subtests; top-level totals were 47 passes, zero failures, one skip. Wall time was **1134.710 seconds**, `nproc` was **5**, and load averages moved from **1.29 / 1.21 / 0.77** to **4.30 / 4.55 / 3.31**. Its command was `go test -json ./stage1/cohere/lint -count=1 -timeout=60m`, with exactly the inputs recorded in `evidence/full-metrics.json`. `TestCheckerBridgeRefusalPending` was the sole skip: `checker_pending_test.go:51` awaits `codex/tsgo-errors-as-values`, because the base prelude does not expose `TSGoError`. No required-input test skipped. The full gate also caught this rule's compiling mutant on Node and emitted JavaScript.

No missing shared helper or algorithm language gap remains. The repository-wide gate and exhaustive configurations beyond the captured cases, named absent-node shapes, witnesses and inherited corpus were not run. No shared source, registry list, upstream rule or compiler file was edited.

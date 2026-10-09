Port of `structure/network-no-forbidden-import` for Kirk and Ahra. Only this rule directory changes; its adapter returns the pinned Go rule unchanged. The rule has no options, fixes or suggestions.

TanStack matches `@tanstack/react-query` exactly and is exempt only where shared `structureFileContext` marks the network module. Apollo matches every `@apollo/` package and has no network-module exemption. A path containing `/generated` reports only when shared `bindingsOf` and `importedNameOf` identify a named import of `graphql`. Aliases retain their imported name; default and namespace imports do not satisfy this check. Reports cover the whole import declaration, and the independent Apollo and generated-path checks can both report on it.

The import projection supplies the parser's exact named links using non-owning numeric indexes. It is built on the first generated-path import and reused for the file. Existing shared helpers perform path classification and binding/name interpretation. No private shared-helper copy or new syntax was introduced.

Validation commands, from the repository root:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/network-no-forbidden-import-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry > /tmp/network-no-forbidden-import-registry.log 2>&1
ADAMIC_LINT_RULES=structure-network-no-forbidden-import GOMAXPROCS=4 go test ./stage1/cohere/lint -run 'TestRulesAgree|TestOwnedWitnesses|TestMutants/structure-network-apollo-prefix' -count=1 -v -timeout=40m > /tmp/network-no-forbidden-import-selected.log 2>&1
ADAMIC_LINT_RULES=structure-network-no-forbidden-import GOMAXPROCS=4 go test -overlay=/tmp/network-no-forbidden-import-overlay.json ./stage1/cohere/lint -run '^TestNetworkForbiddenImport' -count=1 -v -timeout=40m > /tmp/network-no-forbidden-import-native-mutant.log 2>&1
```

The overlay maps virtual `stage1/cohere/lint/network_forbidden_import_validation_test.go` to this directory's `validation.go.txt`, both absolute paths, using Go's `{"Replace": {...}}` format. Both owned top-level tests call `t.Parallel`. They hold the captured upstream count to 17 and require the sanitized native semantic mutant to disagree with Go and agree with mutated source Node.

Selected validation passed in 113.870 seconds. Its discovery capture contained 5,747 unique source/rule/options combinations; the scoped comparison produced 503,845 bytes identical across Go, source Node, emitted JavaScript and ASan/UBSan native. The 17 own upstream cases cover firing, silence and the network-module exemption. The inherited generated corpus and each owned witness also match across all three backends; every owned witness produces a Go finding.

Four witnesses cover client subpaths, imported-name aliases and default/namespace negatives, two findings on one import with UTF-8 offsets, and type-only imports. Go reports type-only named `graphql` imports from generated paths too, and the port preserves this behavior. It also preserves the lexical `/generated` substring check rather than interpreting a directory boundary.

The compiling mutant replaces the Apollo prefix check with an exact `@apollo/client` check, losing the subpath finding. Node and emitted JavaScript caught the disagreement in the selected run. Exact catch lines are in `evidence/mutant.log`, with complete outputs in `evidence/selected.log` and `evidence/native-mutant.log`.

Setup passed in 1.076 seconds, with five reported processors and a four-CPU cgroup quota. Every setup step timing and the measured load are in `evidence/setup.log`; registry output is in `evidence/registry.log`.

Whole-package inputs and measurements are recorded in `evidence/full-metrics.json`; the complete raw JSON test stream is compressed without changing its bytes in `evidence/full.jsonl.gz`. The TypeScript checkout is clean at `050880ce59e30b356b686bd3144efe24f875ebc8`; the WASI sysroot is supplied, benchmarking is enabled, and both profile variables use the same newly created directory. The whole package has no scoped-rule filter.

The owned validation overlay passed: all 17 captured upstream cases were present, and the compiling mutant was caught on sanitized native as well as both JavaScript paths. Native mutated output matched mutated source Node. See `evidence/native-mutant.log`.

The whole lint package passed: **159 passes, zero failures, one skip**, including subtests. Top-level totals were 47 passes, zero failures, one skip. Wall time was **1167.015 seconds**, `nproc` was **5**, and load averages moved from **0.75 / 0.94 / 1.61** to **4.26 / 4.72 / 3.62**. The command was `go test -json ./stage1/cohere/lint -count=1 -timeout=60m`, with the complete environment recorded in `evidence/full-metrics.json`. The sole skip was `TestCheckerBridgeRefusalPending`: `checker_pending_test.go:51` awaits `codex/tsgo-errors-as-values`, because the base prelude does not expose `TSGoError`. No required-input test skipped. The full gate also caught this rule's compiling mutant on both JavaScript paths.

No missing shared helper or algorithm language gap remains. The repository-wide gate and exhaustive module/path spellings beyond the captured cases, witnesses and inherited corpus were not run. No shared source, registry list, upstream rule or compiler file was edited.

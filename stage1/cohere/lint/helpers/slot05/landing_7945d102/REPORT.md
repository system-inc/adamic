# Slot 05 current-main landing

All 95 delivered helpers have fresh passing oracle and semantic-mutant observations on current main and lint area. The strict Go pin and parser adapters are migrated to cohere 7945d102a6c18dd36adf9114a758ce646e8b2359. No new helper or rule is counted. The branch's prior pin-drift landing failure is resolved. The separately reserved regexp.Compile remains blocked by dynamic RegExp construction.

## Bases and source audit

- Main: 48c05d091f0a43c31cbe051b1d6578d99eeedf19.
- Lint area: 65017b318da1995237ff3ea2c80f59b055b39ac3.
- Own integration merge: a17fff7290e343176a42d10887ec314fe14c8bd2.
- Prior published worker HEAD: 872037fa230c737e69f079596811b82170d9a421.
- Final fetch: both bases unchanged; identity.txt records them.
- Only codex/lint-helpers-05 is pushed. No history is rewritten. Main's compiler changes are accepted intact.

The reproducible Go declaration audit covers exactly 95 delivered symbols. It compares formatted signatures and bodies, with receiver identity and comments excluded: 92 identical declarations, a FileName().AsString() conversion in NormalizedFileName, and two changed ClassLiteralReader methods. It is not a proof that every transitive dependency or parser behavior is unchanged. Full source hashes and changed declarations are retained in helper-body-audit.json; reproduction is in README.md.

## What changed

Only owned helper tests and their Go oracle adapters change pins and typed AST filename/path construction. Every test still refuses a pin other than the new exact hash. No provenance guard is disabled. The oracle-only filename constructor preserves the original raw unnormalized boundary inputs through an explicit typed conversion; the target helper is unmodified.

Current Go calleeValues delegates to readsCallee rather than requiring a bare identifier enabled by an exact-name map. The owned helper now accepts the readsCallee(index) dependency and continues to collect all arguments into fresh payload lists in order. The current variableValues delegates to variables.matches, so the owned helper now takes matches(name) rather than an array of individual RE2 callbacks. Initializer/name guards and collection delegation are preserved.

These two APIs change for callers. A source search found only the owned batch2 driver importing either implementation. The README states the new contracts and the current reader predicate dependencies. The driver supplies actual Go predicate results as dependency inputs, independently of the actual Go target helper's answers. No selector matcher, regex compiler, shared harness, registry or rule is edited. The original pin's reader implementations and reports remain in Git history and their evidence remains intact.

Batch2 adds direct, property, string-element and configured member-path controls. All 11 inventory consumer test files still contribute their captured strings. Current coverage: 1,043 distinct strings and controls, 3,129 parser/settings cases, 189 callee queries and 1,026 variable queries. The argument-order mutant repeats the same 189 callee queries. Raw current Go inputs and verdicts are retained losslessly. Go, source Node, emitted JavaScript and sanitized native agree on successful results. Actual Go refusal on nil/wrong-kind input is checked separately against explicit source/native exit-70 refusals; exact Go crash prose is not claimed.

## Fresh verification

34 unique packages have successful current-pin observations: the root's three owned TestSlot05 tests, batches 2 through 33, and the batch34 gap test. The root command does not select other workers' tests or the four inherited common-helper mutants. The 95 own helpers' 467 semantic mutants compile, execute with exit 0 and empty stderr, then differ from actual Go. observations.json retains every differing-output witness and individual package result. Older packages retain their established backend coverage; not every original package is claimed to have an emitted-JavaScript check.

Commands, all output written directly to logs:

```
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -run '^TestSlot05' -count=1 -v -timeout=30m
ADAMIC_GATE_UNCACHED=1 go test -p 2 ./stage1/cohere/lint/helpers/slot05/batch2 ... ./stage1/cohere/lint/helpers/slot05/batch32 ./stage1/cohere/lint/helpers/slot05/batch34 -count=1 -v -timeout=30m
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch33 -count=1 -v -timeout=20m
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch2 -count=1 -v -timeout=20m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v
go vet ./...
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05 stage1/cohere/lint/helpers/slot05_test.go
```

The exact fully expanded broad command, evidence environment variables and log paths are retained in run-owned.sh. Its first aggregate run exits 1 because batch2 was tested before the removed Go field variablePatterns was migrated. That failure remains in all-batches.log.gz and is not called a green aggregate invocation. Every other package in that invocation passes. The separately completed batch2 is green after its source and adapter corrections; no other package imports it or changed during the broad run.

- Root owned helpers: PASS 98.529s; three mutants.
- Completed batch2: PASS 114.635s; three mutants; source/emitted/native comparison.
- Batch33: PASS 67.718s; 8,539 queries; 21 mutants.
- Batch3 through batch32: each PASS; exact package durations in package-results.md.
- Batch34 gap test: PASS 0.183s. Actual new Go and source Node still print true for runtime pattern a, flag u, subject a; current lowering returns typed NotYet for a nonconstant pattern. This is not a delivered Compile helper or a semantic mutant.
- Filtered uncached Node input oracle: PASS 1.541s; seven probe misses, zero hits.
- Final repository-wide vet and owned/repository Go formatting: exit 0, empty logs.

The exact successful package set and 467 mutant witnesses are asserted in the evidence accounting. All earlier failures are preserved. An initial source-audit tool error on repeated unnamed init functions was corrected by excluding init from named-helper lookup. An initial migrated driver used unsupported optional element access; it was rewritten with an explicit undefined guard in owned source, then rerun completely. The summary parser was corrected to use horizontal whitespace and consistent tuple comparisons; no helper verdict or expected test result changed to make accounting pass.

## Setup

GOPROXY=https://proxy.golang.org|direct preceded setup. Setup succeeded in 24.030s; nproc 5, CPU quota 400000 100000, memory 17.6 GB. Observed timing lines: Go 0.022s, Node 0.023s, submodules 0.057s, markdown dependencies 0.066s, clang 0.161s, Go build 23.858s, test binaries deferred 24.002s, cache warm 24.004s. Tool versions: Go 1.27.1, clang 20.1.8, Node 24.19.0. Complete setup output is retained.

## Readiness and limits

This is a landing repair of the 95 delivered helpers, not an additional helper wave. Frozen inventory readiness counts are unchanged; zero new rules lose blockers. Caller-supplied selector, collection, parser and other documented dependencies remain separate. This does not establish independent Adamic AST parsing, entire consumer rule findings/fixes or production integration of every helper.

regexp.Compile remains reserved but undelivered: runtime-supplied pattern a is refused before native or JavaScript emission. The four frozen consumers remain @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. No hand-rolled matcher or shared compiler workaround is supplied. No further helper is claimed in this landing unit.

The full repository gate and its 17 required external stage 1 checks were not run. No selected test is skipped, relaxed, deleted or filtered to an empty match. Historical common-helper proof is preserved but its four mutants were not rerun in this owned package selection. Large fresh corpora/verdicts are compressed losslessly; no prior pushed evidence is deleted or overwritten.

The aggregate raw log is retained losslessly as all-batches.log.gz because an empty Go verdict leaves trailing whitespace in a mutant witness. Decompression preserves every original byte.

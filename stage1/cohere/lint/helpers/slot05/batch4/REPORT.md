Built: Theme.liveKeys, Theme.Entries and Theme.KeysInNamespaces, one public helper per .a file; six consumers each.
Commits: implementation 7a221b0; claims 71331d3, e033ced and 57ab4ae were pushed before their code; final evidence recorded in this report commit.
Checks: package PASS 47.216s; Go/Node source/emitted JavaScript/sanitized native agree; vet, types, formatting and six filtered oracle fixtures pass.
Mutants: inverted visitor stop, erased options and admitted sub-variables all compile and disagree with Go; withdrawn escape experiment also caught its byte substitution.
Not covered: whole-rule findings/fixes/suggestions, production integration, dynamic/external fixtures, arbitrary malformed stores, invalid UTF-8 and full repository gate.

## Ownership and delivery

Existing retained claims were already implemented, tested and pushed before this batch. Fetched all origin codex/lint-helpers* branches, read every claim, and selected the highest remaining unclaimed consumer count. These theme helpers each cover six rules. Final refresh found their claims only on slot 05. All changes are owned helper files and the independent private test runner; no shared harness, generator, compiler implementation or cohere worktree was edited.

The initial escape-terminator reservation ea3c330 at 01:31:27 UTC lost to slot 03's 685fc6d at 01:31:08 UTC. Its implementation/test were removed and its experiment is labeled withdrawn. Theme.Entries replaced it. Only the three theme helpers count as this delivery. PrefixKey and isIgnoredThemeKey remain explicit separately owned callback dependencies, with actual Go answers supplied by the oracle.

## Observed behavior and consumers

liveKeys preserves captured insertion order and length while looking up the current map at each visit, skipping empty/stale slots and honoring early stop. Entries preserves key/value/options and returns fresh records after prefixing. KeysInNamespaces preserves namespace-major order and duplicate requests, excludes sub-variables using Go's UTF-8 byte offset, applies the external ignored-key predicate, and returns suffixes. The matched-too-short-key case explicitly refuses, as Go does. Controlled deletion, overwrite, re-insertion and visitor mutation cases are exercised.

[CONSUMERS.md](CONSUMERS.md) lists each helper's rules: better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. This removes 18 dependency occurrences across six rules, with no final blocker removed by this batch alone. [readiness.json](readiness.json) retains all residual dependencies. Across all eleven retained slot 05 helpers, conservative accounting removes 115 occurrences across 64 rules, with four final helper blockers removed (base 46 to 50 helper-ready, conditional on the common adapter). Other workers' implementations are not assumed integrated.

## Oracle and evidence

Pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db is the behavioral oracle, exposed through an overlay. All six inventory-listed consumer test files contribute nonempty Go string literals, including prose/options: 639 distinct strings. Per-consumer counts are 259, 370, 110, 131, 247 and 239 in the consumer order above. Missing coverage and pin drift fail. Actual Go Add builds theme stores; the readers call actual Go methods.

- liveKeys: 13,461 state/stop/mutation cases, 26,238 Go visits.
- Entries: 1,923 state/prefix cases with fresh reads, 11,520 Go entries.
- KeysInNamespaces: 2,556 state/namespace cases, 12,640 returned keys, plus short-key refusal.

Total: 17,940 state cases and 50,398 visited/returned records. Baselines match Node source, Adamic-emitted JavaScript and sanitized native against actual Go. Committed coverage JSON and corpus SHA-256 files are under [evidence/](evidence/). The first final run passed in 46.063s but an inherited BATCH3 evidence-variable typo prevented saving metadata; the variable was corrected to BATCH4 and the complete package rerun passed in 47.216s with metadata saved.

Every retained mutant compiles and runs without stderr or sanitizer failure, then disagrees with actual Go:

- liveKeys: invert callback stop condition; verdict 22 returns count 1 rather than Go 3.
- Entries: replace options with zero; verdict 10 returns 0 rather than Go 17.
- KeysInNamespaces: disable nested-subvariable rejection; verdict 1 returns count 9 rather than Go 6.
- Withdrawn escape experiment: substitute vertical tab (11) for form feed (12); verdict 12 returns true rather than Go false. All 256 byte verdicts passed baseline before removal. This experiment is excluded from retained totals.

## Commands and environment

All test output was redirected directly to files. Source /workspace/adamic-tools/env.sh first.

```sh
ADAMIC_SLOT05_BATCH4_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch4/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch4 -count=1 -v -timeout=15m > /tmp/lint05-batch4-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot05/batch4 > /tmp/lint05-batch4-vet.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch4/main.a > /tmp/lint05-batch4-types.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot05/batch4 > /tmp/lint05-batch4-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/lint05-batch4-oracle.log 2>&1
```

Package PASS 47.216s; vet exit 0; types exit 0; formatting output empty; filtered oracle PASS 0.960s, six fixtures, zero hits and six probe misses. Logs are committed under evidence/. nproc: 5. Reused the successful cloud/setup.sh toolchain from earlier delivery: Go 1.27.1, clang 20.1.8, Node 24.19.0. Original timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (88s)
setup: done in 88s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

## Limits

This establishes helper behavior on the constructed Go-derived corpus, not whole-rule lint equivalence. Full diagnostics/fixes/suggestions, production theme/parser/linter integration, dynamically concatenated or externally loaded fixtures and the full repository gate were not run. Nil receivers/zero-value Go themes, arbitrary mutation of value fields or shortening the original order array, concurrent mutation and raw invalid UTF-8 representations are outside this port's store contract. The README specifies explicit refusals and representation limits. Emitted-JavaScript comparison is covered here without modifying the shared rule harness.

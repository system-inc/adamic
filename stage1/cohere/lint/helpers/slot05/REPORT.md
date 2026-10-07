Built three `.a` helpers: React identifier matching, import binding identities, and React component-base recognition.
Implementation: `72358fc`; claims were pushed before code and simultaneous collisions resolved by earliest commit.
Checks: the complete helper package passed in 140.792s; Go/Node/sanitized-native comparisons, seven compiling mutants, vet, formatting, types and six uncached input fixtures passed.
Mutants: all three new mutations and all four existing helper mutations were caught by comparison with Go output.
Not covered: the full repository gate, full rule findings/fixes, dynamic fixture reconstruction, and production AST adapter integration.

## Branch and reservations

Branch `codex/lint-helpers-05` starts at fetched `origin/codex/lint-helpers` commit `29990b47913aa77d542045194b512484af995d29`. The unit-specific base instruction takes precedence over the generic main-base instruction. Main was fetched too. The initial default remote refspec fetched only main; explicit `git fetch origin '+refs/heads/codex/lint-helpers*:refs/remotes/origin/codex/lint-helpers*'` obtained every helper branch.

Every helper selection used the complete readiness ledger and all available remote claim files. The comments bundle was additionally respected under the existing `HELPERS.md` claim. New branches appeared concurrently, causing duplicate reservations. Slot 05's identifier reservation `41a6089` at 00:20:44 UTC was first; all other claimants withdrew, and `32385e2` restored its ownership before delivery. The temporary attribute reservation `b73ec42` was withdrawn because slot 02's `b25d3cd` was one second earlier. No attribute helper is delivered.

Import bindings were reserved by `024ba20`. Component bases were reserved by `c4c5bb7` at 00:26:47 UTC, before slot 04's 00:26:54 claim; slot 04 withdrew. `c1aec45` records final unique ownership. No force push, branch deletion or pull request was used.

## Deliverables and dependency accounting

| Helper | File | Dependent rules | Final listed blockers removed |
|---|---|---:|---:|
| `react.isIdentifierNamed` | `react_identifier_named.a` | 25 | 0 |
| `imports.BindingsOf` | `import_bindings.a` | 15 | 2 |
| `react.isComponentBase` | `react_component_base.a` | 14 | 0 |

The 54 dependency occurrences overlap across 38 rules. The two newly helper-ready rules are `@next/next/no-location-assign-relative-destination` and `@typescript-eslint/no-import-type-side-effects`. Inference from the frozen ledger: helper readiness rises from 46 to 48, assuming the inventory's common AST adapter. This does not mean any new rule is implemented. [CONSUMERS.md](CONSUMERS.md) lists every consumer by helper; [readiness.json](readiness.json) retains the residual dependencies for all affected rules.

`HelperNode` uses immutable flat records and numeric edges, so no AST ownership cycle is introduced. Identifier matching skips only parentheses. Import outputs preserve default identifiers, whole namespace nodes and ordered named-specifier identities. Component recognition takes the separately claimed `isComponentBaseName` as an explicit function dependency and reuses this slot's identifier predicate. The oracle supplies the real Go leaf's answers to isolate the delivered helper without copying another worker's leaf implementation.

## Toolchain and validation

Setup command: `bash cloud/setup.sh > /tmp/lint05-setup.log 2>&1`. Exit 0. `/opt/adamic-tools/env.sh` does not exist in this instance; the setup's installed path is `/workspace/adamic-tools/env.sh`, sourced for every build/test shell. The initial attempt to use gofmt before sourcing returned `gofmt: command not found`; sourcing corrected it.

Go 1.27.1, clang 20.1.8, Node 24.19.0. `nproc` printed 5. Setup timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (88s)
setup: done in 88s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

The three final-slot baseline comparisons generated the following independent Go verdicts. Each result is compared to Node running the `.a` source and sanitized native running the same source, with exit 0 and empty stderr required.

| Helper | Consumers represented | Distinct test-file strings plus controls | Go verdicts |
|---|---:|---:|---:|
| Identifier | 25 | 3,362 | 285,473 |
| Imports | 15 | 2,319 | 32,689 |
| Component base | 14 | 2,515 | 44,734 |
| Total verdicts | overlapping consumers | overlapping sources | 362,896 |

The oracle reads every nonempty string literal from every inventory-listed test file for these consumers, deduplicates, and parses every string with pinned Go cohere TSX. It queries every node, not just likely-positive nodes. This includes fixture-source, prose and option strings; counts are not counts of full lint fixtures. Missing coverage for any consumer fails the oracle. Per-consumer counts and deterministic corpus hashes are in `evidence/*-coverage.json` and `evidence/*-corpus.sha256`.

All test commands wrote to log files directly, never through a pipe:

- `ADAMIC_SLOT05_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > /tmp/lint05-package.log 2>&1`: first complete run exited 1 in 219.838s. Every test except the pre-existing `TestHelpersMatchCohere` passed. That test's strict stderr guard rejected Go's informational downloads of `github.com/dlclark/regexp2` versions v1.11.5 and v2.5.2 while building the oracle. The download itself succeeded. All three final-slot comparisons and mutants passed, including the external-leaf callback.
- `ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -run '^TestHelpersMatchCohere$' -count=1 -v -timeout=20m > /tmp/lint05-baseline-retry.log 2>&1`: PASS, 15.274s after those modules were cached. 22,412 lines matched Go, Node and sanitized native. No implementation was changed to suppress stderr.
- `ADAMIC_SLOT05_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > /tmp/lint05-package-final.log 2>&1`: PASS, 140.792s. Every pre-existing helper test, all three new Go/Node/sanitized-native comparisons, and all seven compiling mutants passed. See `evidence/package-final.log`. The cold-cache failure is resolved, not omitted from the evidence.
- `go vet ./... > /tmp/lint05-vet-final.log 2>&1`: exit 0, empty output.
- `gofmt -l cmd internal stage1/cohere/lint/helpers/slot05_test.go stage1/cohere/lint/helpers/slot05/testdata > /tmp/lint05-format.log`: exit 0, empty output. `git diff --check` also passed.
- `go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/main.a > /tmp/lint05-types.log 2>&1`: exit 0; printed the entry point's checked variable, parameter and function types.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/lint05-oracle.log 2>&1`: PASS, 6.930s; all six fixtures, zero cache hits and six probe misses.

The full repository test gate was not run. This unit uses the entire touched helper package, repository-wide vet and a filtered uncached oracle. No compiler file or existing helper file was modified. The Go overlay is test-only and leaves the submodule worktree unchanged.

## Every mutant run

A mutation is credited only after it compiles, exits successfully without stderr, and produces output different from Go. Native builds use ASan/UBSan and Linux leak checking. The final-slot mutant witnesses were the same before and after the leaf became an explicit dependency.

| Helper | Mutation | What caught it |
|---|---|---|
| New identifier | Stop skipping parenthesized expressions | Go output comparison, verdict 83,935: mutant `false`, Go `true` |
| New imports | Return the namespace's identifier instead of the namespace node | Go identity comparison, verdict 1,916: mutant `-1:8:`, Go `-1:7:` |
| New component base | Accept any receiver namespace instead of requiring React | Go output comparison, verdict 13,826: mutant `true`, Go `false` |
| Existing JSON reader | Accept a raw control character in a JSON string | Go comparison, line 15,157: mutant `valid`, Go `invalid` |
| Existing schema | Accept multiple matches in `oneOf` | Go comparison, line 15,166: mutant `valid`, Go `invalid` |
| Existing strict options | Skip an unknown or wrong-case field | Go comparison, line 15,178: mutant `valid`, Go `invalid` |
| Existing policy message | Omit value interpolation | Go comparison, line 22,166: literal `{{constructor}}` instead of supplied `sentinel é😀` |

## Observations and limits

An initial nil-input probe of the real identifier helper crashed inside Go `ast.SkipParentheses`, which runs before the apparent nil guard. The delivered helper explicitly panics on a missing expression; it never turns that failure into a negative lint answer. Exact Go runtime crash prose is not promised. The temporary attribute experiment also exposed the Go `Kind` prefix on stringified kinds; the final snapshot adapter strips it, and the successful controls demonstrate that identifiers and non-identifiers are distinguished.

This unit does not reconstruct fixture source assembled dynamically from multiple Go expressions, replay external corpus files loaded at runtime, or compare full rule diagnostics, spans, fixes, defaults or option processing. It does not integrate these projections into Adamic's parser or the existing linter. Rule workers must supply the documented Go-compatible arena fields and component-name leaf. Invalid arena indexes and cyclic parenthesis chains refuse explicitly; arbitrary malformed arena crash parity is outside the supported contract. All new Adamic source files are `.a`.

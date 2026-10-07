Built: NewTheme, Theme.resolveKey and Theme.ResolveValue, one public helper per .a file; six consumers each.
Commits: claim b9d9080 and precedence cac1ca9 pushed before code; implementation 748b21b; final report/evidence in this report commit.
Checks: package PASS 157.354s; Go/source Node/emitted JavaScript/sanitized native agree across 21,300 cases; vet/types/format and six filtered oracle fixtures pass.
Mutants: all twelve compiling semantic mutants caught by actual Go output comparisons, detailed below; the earlier oracle-data panic is not credited.
Not covered: full-rule findings/fixes/suggestions, production integration, dynamic/external fixtures, raw invalid UTF-8 and full repository gate.

## Ownership and delivery

All fourteen earlier retained helpers were completed, tested and pushed through 2a26730 before selecting this batch. Fetched every origin codex/lint-helpers* branch and inspected every claim. Higher concrete-symbol counts were reserved; the earlier comment bundle remains reserved in shared HELPERS.md. The selected three helpers each tie the highest unclaimed count, six. The claim was pushed before implementation. Subsequent refreshes include newly created helper branches, not just the original five slots.

NewTheme raced with three later reservations. Earliest claim commit wins under the established helper-wave rule:

| Claim | UTC commit time |
|---|---|
| slot 05 b9d9080 | 2026-10-07 02:30:29 |
| lint-wave1-08 e069e11 | 2026-10-07 02:30:35 |
| lint-wave1-06 b07ef13 | 2026-10-07 02:30:45 |
| lint-wave1-13 f0b136e | 2026-10-07 02:32:43 |

Slot 05 retains the constructor; precedence was published in cac1ca9 before code. The other reservations remain visible in the final remote snapshots and must not be integrated as additional constructors. No competing resolveKey or ResolveValue claim appeared. [ownership.json](evidence/ownership.json) records every matching remote branch SHA, full claim text and first-claim timestamp. No fourth helper was reserved.

New Adamic files are .a. Changes are owned slot05/batch6 files, the slot README and claims/05.md. No shared registration generator, shared rule harness, compiler implementation or cohere worktree was edited. No shared-harness gap blocks these three helpers.

## Observations and contracts

newTheme returns independently writable empty containers, empty prefix and zero deadKeys. The oracle creates two actual Go themes, snapshots both, mutates only the first and snapshots both again. Mutants sharing the values map or order list demonstrate that the independence checks fail.

resolveThemeKey preserves namespace order, distinguishes absent candidate from present-empty candidate, prefers an existing direct key, and only tries all-dot-to-underscore replacement when a present dotted candidate's direct key is absent. It filters through the separately owned ignored-key predicate and returns the first accepted key. The predicate's actual Go answers are provided explicitly; its implementation is not copied. A typed result preserves found separately from key text.

resolveThemeValue executes this batch's actual resolver and returns the stored literal value, preserving found=true for an empty string. Options and prefix do not change literal resolution. The shared immutable value/read-only store types and previously delivered mutable store type are reused. No ownership cycle or garbage collector is introduced. The private driver supplies exact typed state and never mutates the store through the ignored-key callback.

## Actual Go oracle and bounded coverage

An overlay exposes the private methods and snapshots without writing Go cohere. Pin drift from 715ba94f3608a6500086b1076ce5cb7e51b836db fails. All six consuming rules' inventory-listed test files are scanned for nonempty Go string literals: 639 distinct strings, including descriptions/options as well as source. Each consumer contributes nonzero counts (259, 370, 110, 131, 247, 239 in CONSUMERS.md order). Missing coverage or consumer count other than six fails.

Actual Go Add constructs stores for every source and derived dotted candidate, with direct-plus-fallback and fallback-only variants. Namespace arrays vary ordering, duplication, emptiness and missing namespaces; both presence states are checked. Controls include ignored font subnamespaces, direct precedence, multi-dot replacement, present-empty candidates, missing keys and found-empty values. Source strings that equal Go's initial deletion directive retain actual Go Add behavior. Constructor mutation appends a stored suffix to avoid inadvertently exercising that directive while testing allocation independence.

| Helper | Cases | Observations |
|---|---:|---|
| NewTheme | 640 | 2,560 current-store snapshots across two instances |
| resolveKey | 10,330 | actual Go found/key pairs |
| ResolveValue | 10,330 | actual Go found/value pairs |
| Total | 21,300 | four-way baseline agreement |

Every baseline compares Go bytes to Node source, Adamic-emitted JavaScript and sanitized native. The namespace predicate answers include absent candidate keys, so missing oracle metadata cannot kill a semantic existence-check mutant. Deterministic corpus hashes and per-consumer counts are committed under [evidence/](evidence/). Constructor map snapshots use an explicit UTF-8 byte comparator; order arrays remain unsorted.

## Every mutant and what caught it

All twelve final mutants independently compile, exit 0 without stderr or sanitizer failures, then differ from Go output. Compilation errors, crashes and sanitizer findings are never credited. The first differing stdout line is included for reproducibility. Keys and literal values can contain newlines; these are output-line witnesses, not whole input names or fixture counts.

| Helper | Mutation | Actual Go comparator witness |
|---|---|---|
| NewTheme | deadKeys initialized to 1 | line 2: mutant 1, Go 0 |
| NewTheme | prefix initialized to wrong | line 1: mutant wrong, Go empty |
| NewTheme | values map shared across calls | line 21: second map count 1, Go 0 |
| NewTheme | order list shared across calls | line 20: second order count 1, Go 0 |
| resolveKey | always append candidate | line 5: mutant --first-seed., Go --first |
| resolveKey | disable dot fallback | line 21: mutant --second-seed., Go --first-seed_ |
| resolveKey | disable ignored-key filtering | line 20769: mutant true, Go false |
| resolveKey | replace only the first dot | line 21: mutant --second-seed., Go --first-seed_ |
| resolveKey | reverse namespace order | line 2: mutant --second-seed., Go --first-seed. |
| resolveKey | skip fallback existence guard | line 2: mutant --missing-seed_, Go --first-seed. |
| ResolveValue | return key instead of value | line 2: mutant --first-seed., Go empty output line |
| ResolveValue | treat an empty stored value as missing | line 9: mutant false, Go true |

Exact anchors and outputs are in package-final.log. Two earlier failed attempts are kept separately and are not passes: initial-import-failure.log records a copied type-import path in the owned driver (4.592s); it was corrected locally. initial-missing-prediction.log records the guard mutant's missing ignored-key oracle answer (179.829s). That panic was not counted as a catch. The oracle now supplies actual Go answers for namespace/direct/fallback keys, including absent ones, and the complete final run passes in 157.354s with a real output mismatch for that mutant. No production helper or shared-harness workaround was required.

## Consumers and readiness inference

[CONSUMERS.md](CONSUMERS.md) lists each helper's six rules: better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. Eighteen listed dependency occurrences are removed. No rule loses its last helper blocker from this batch alone. [readiness.json](readiness.json) preserves residual blockers after this batch alone and cumulative accounting.

Across seventeen retained slot 05 helpers, 151 dependency occurrences are removed across 64 unique consumers; four final helper blockers have been removed in total (base 46 to 50 helper-ready, conditional on the common AST adapter). Other workers' implementations are not assumed integrated. This is a dependency-ledger inference, not an observation of completed native rules. No inventory stage1 status is rewritten.

## Commands, outputs and environment

All test output goes directly to log files. bash cloud/setup.sh succeeded, then /workspace/adamic-tools/env.sh was sourced. Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc printed 5. Setup timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (17s)
setup: done in 17s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

```sh
bash cloud/setup.sh > /tmp/lint05-batch6-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH6_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch6/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch6 -count=1 -v -timeout=15m > /tmp/lint05-batch6-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot05/batch6 > /tmp/lint05-batch6-vet.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch6/main.a > /tmp/lint05-batch6-types.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/lint05-batch6-oracle.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot05/batch6 > /tmp/lint05-batch6-format.log
nproc
```

Package PASS 157.354s; vet exit 0, empty output; types exit 0 with checked declarations; formatting output empty. Filtered uncached oracle PASS 1.027s, all six fixtures, zero hits and six probe misses. Test serialization is documented because generated corpora and sanitized compiler builds share the memory budget. The bounded touched-package and filtered-oracle gate was used; the full repository gate and unchanged earlier packages were not rerun.

## Limits

Whole-rule findings, fixes, suggestion serialization and production parser/theme/linter integration remain rule-worker work. Dynamic concatenations and external fixture input files are not reconstructed. Raw invalid UTF-8, nil Go receivers, arbitrary concurrent/callback mutation and nil-versus-empty allocation distinctions are outside the usable-store representation contract. A callback removing the resolved entry would cause an explicit port panic and is outside the pure ignored-key dependency contract. These comparisons are bounded, not exhaustive over all stores and namespaces. The exact tested empty-value distinction, fresh-instance writes and emitted-JavaScript behavior are covered.

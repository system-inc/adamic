Built: Theme.clearAll, Theme.compactKeyOrder and Theme.delete, one public helper per .a file; six consumers each.
Commits: claim a0e06da pushed before code; implementation cbaaf3d; evidence and report finalized in this report commit.
Checks: package PASS 75.431s; actual Go/source Node/emitted JavaScript/sanitized native agree; vet/types/format and six filtered oracle fixtures pass.
Mutants: all thirteen compiling semantic mutants caught by actual Go store/alias comparisons, detailed below.
Not covered: whole-rule findings/fixes/suggestions, production integration, dynamic/external fixtures, raw invalid UTF-8 or the full repository gate.

## Claims and delivery

All eleven prior retained helpers were implemented, tested and pushed through 37d5108 before this batch. Fetched every origin codex/lint-helpers* branch and checked every claim. The larger comment bundle remains reserved by shared HELPERS.md on codex/lint-helpers. All larger concrete symbols were reserved; the three selected mutation helpers tied the highest remaining unclaimed count, six. Claim a0e06da was pushed before any implementation. A final fresh scan found these claims only on slot 05; [ownership.json](evidence/ownership.json) records every refreshed branch SHA and matching claim lines. No claim collision or fourth reservation occurred.

All new Adamic files are .a. Changes stay in slot05/batch5, the slot's README and its claim file. No shared registration generator, shared rule harness, compiler implementation or Go cohere worktree was changed. Existing compiler APIs support these helpers and both output backends; no shared-harness gap blocks this delivery.

## Observations

clearAll replaces the map and order list rather than clearing retained containers in place. It resets deadKeys while retaining prefix. compactKeyOrder is a no-op at dead <= live, then filters nonblank present keys into a new ordered list when dead > live, resetting deadKeys. It keeps duplicate present slots, preserves map/value contents and leaves a retained old array untouched. delete returns immediately for absent keys; otherwise it deletes the map entry, blanks only the last matching slot, increments deadKeys and calls this batch's real compactor. A retained array observes the blanking before compaction replaces it. These outcomes match actual Go state and alias snapshots.

The actual Go methods are exported by an oracle-only overlay, pinned to cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. Every nonempty Go string literal from every inventory-listed test file of each consumer is scanned, including description/option strings: 639 distinct strings. All six consumers contribute nonzero counts (259, 370, 110, 131, 247, 239 in CONSUMERS.md order). Missing contributions, consumer-count drift and pin drift fail.

Actual Go Add builds stores containing consumer-derived values, overwrites, deletes and reinsertions, plus Unicode keys. Controls include empty stores, dead/live boundary equality and excess, blank/stale/duplicate order slots, orphan/empty map keys and a negative synthetic dead count to observe the exact private-method branch. Some controls intentionally exceed states normally produced by Add. They establish private-method behavior, not evidence that a consuming rule creates those states.

For each operation the oracle records the current theme, retained old containers, current state after writes through the old aliases, and a repeated operation. Prefix, dead count, exact order and full key/value/options map contents are compared. Map snapshots are sorted by Go UTF-8 bytes; order lists remain unsorted. Native's private driver uses an explicit UTF-8 comparator, tested with supplementary/private-use ordering. No garbage collector or ownership cycles are introduced.

| Helper | Cases | Go store/alias snapshots |
|---|---:|---:|
| clearAll | 1,927 | 7,708 |
| compactKeyOrder | 1,927 | 7,708 |
| delete | 11,562 | 46,248 |
| Total | 15,416 | 61,664 |

All baseline output matches actual Go on Node source, Adamic-emitted JavaScript and ASan/UBSan native. The thirteen native mutants independently compile, exit 0 with no stderr, and then disagree with Go. Compiler errors, crashes and sanitizer errors do not count as semantic kills. [Coverage counts and corpus hashes](evidence/) reproduce the generated inputs; tests write artifacts only when the evidence variable is set.

## Every mutant

The first differing output line is recorded below; output lines can include newlines from consumer values and are not fixture counts. Empty string witnesses are labeled empty. Full exact anchors and outputs are in package-final.log.

| Helper | Mutation | First Go comparison witness |
|---|---|---|
| clearAll | retain dead count | line 2: mutant -1, Go 0 |
| clearAll | retain values map | line 4: mutant map count 2, Go 0 |
| clearAll | retain order array | line 3: mutant order count 6, Go 0 |
| compactKeyOrder | compact at equality (<= becomes <) | line 206: mutant dead 0, Go 2 |
| compactKeyOrder | admit blank slots | line 469: mutant order count 2, Go 0 |
| compactKeyOrder | admit absent-map slots | line 275: mutant order count 4, Go 3 |
| compactKeyOrder | retain dead count after compaction | line 274: mutant 3, Go 0 |
| delete | retain key text instead of blanking | line 75: mutant --a, Go empty |
| delete | remove absent-key guard | line 2: mutant dead 0, Go -1 |
| delete | delete a different absent map key | line 78: mutant map count 2, Go 1 |
| delete | scan forward rather than backward | line 74: mutant empty, Go --a |
| delete | omit dead-count increment | line 70: mutant -1, Go 0 |
| delete | continue after blanking instead of break | line 74: mutant empty, Go --a |

The earlier three-mutant complete package also passed in 32.805s before adding the other ten probes. That log is retained as initial-three-mutants.log; only the expanded 75.431s run is the final gate. Committed copies strip trailing whitespace from two empty-string mutant log lines; original raw output remains in /tmp/lint05-batch5-tests.log and /tmp/lint05-batch5-final.log.

## Rules and readiness inference

[CONSUMERS.md](CONSUMERS.md) names the six rules for each helper: better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. This removes eighteen listed prerequisite occurrences. No rule loses its last helper blocker from this batch alone. [readiness.json](readiness.json) records remaining dependencies after this batch alone and cumulative slot 05 accounting.

Across fourteen retained slot 05 helpers, 133 dependency occurrences are removed across 64 unique consumers, with four final helper blockers removed in total (base 46 to 50 helper-ready, conditional on the common adapter). Other workers' helpers are not assumed integrated. This is a ledger inference, not an observation of completed native lint rules.

## Commands and outputs

Every test command redirects directly to a log, with no test-output pipeline. bash cloud/setup.sh succeeded; then source /workspace/adamic-tools/env.sh. Environment: Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc printed 5. Setup timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (19s)
setup: done in 19s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

```sh
bash cloud/setup.sh > /tmp/lint05-batch5-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH5_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch5/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch5 -count=1 -v -timeout=15m > /tmp/lint05-batch5-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot05/batch5 > /tmp/lint05-batch5-vet.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch5/main.a > /tmp/lint05-batch5-types.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/lint05-batch5-oracle.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot05/batch5 > /tmp/lint05-batch5-format.log
nproc
```

Package PASS 75.431s; vet exit 0 with empty output; types exit 0 with checked declarations; formatting output empty. Filtered uncached oracle PASS 0.945s, all six fixtures, zero hits and six probe misses. Logs are committed under evidence/. The bounded touched-package plus filtered-oracle gate was used; the full repository gate and unchanged earlier packages were not rerun.

## Not covered

Whole-rule findings, fixes, suggestion serialization and production parser/theme/linter integration remain rule-worker work. Dynamic concatenations and external fixture sources are not reconstructed. Nil Go receivers, concurrent mutation, raw invalid UTF-8 strings, Go slice capacity and nil-versus-empty allocation distinctions are outside the usable-store representation contract. This is bounded state/alias comparison rather than an exhaustive unbounded mutation proof. Exact map/list content and alias mutation are covered; emitted-JavaScript comparison is covered. No inventory status marks any rule ported.

Built selected-rule certification, seven atomic caches, and coverage/registration regressions.\
Source commits: `e5d093e20b29f4965fff06a56b6ef9ac23001a22`, `5f410bc4c7d3617ecee9347afa74dbe8db92082f`; base `f4d98cab50048692781da3599131317dc569d466`.\
Commands: 36/36 timing samples passed; cache, parity, uncached equality and filtered compiler oracle passed; complete uncached lint/registry/CLI package gate passed.\
Mutants: all seven cache key omissions rejected stale answers; inherited-all regression and each selected rule mutant caught.\
Limits: edited-rule target remains blocked by clang; full repository integration, fifty-rule fleet and per-file compilation were not covered.

The branch is `devtools/lint-rule-check`, based on `origin/codex/lint-harness-dot-a`, not main. No rule directory, cohere submodule source, internal compiler file or lint-registry validation was changed.

From the repository root:

```sh
go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-var
```

The wrapper `go run ./cmd/adamic-lint-check <slug>` executes that same test. The check preserves original inherited rows, including `all` selections, and adds selected-rule variants, owned witnesses and captured upstream cases. The private Go oracle and port retain the baseline registrations plus the selected rule. The owned mutant must differ from Go on source Node, emitted JavaScript and sanitized native. Sanitized native executes on every cache hit.

**Controlled measurements, seconds**

Build-flags line for every measurement below:

```text
commit=5f410bc4c7d3617ecee9347afa74dbe8db92082f; nproc=5; cpu.max=400000 100000; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; flags=-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all; GOCACHE=/home/agent/.cache/go-build
```

Every row in [measurements.json](measurements.json) also records this line, the exact command, exit status, cache mode and load averages before/after. All runs were sequential on this box at the same source commit, interleaved original/cold/edited/unchanged for each rule, repeated three times. The initial interrupted and superseded diagnostic runs are excluded.

Cold means an empty `adamic/lint` cache. Go compilation uses the same existing GOCACHE; the existing native runtime library is primed by the preceding original-harness run in the same isolated cache directory. Each loop uses a new cache directory. The original harness always rebuilds its oracle, recaptures upstream tests and compiles its port. The cache mode `cached` means caching is enabled, rather than every artifact hitting: a rule edit misses the port/module entries while the Go oracle and upstream capture remain reusable. The byte edit appends exactly one space to the selected module in a private snapshot. The unchanged check reuses the unedited cold snapshot artifacts.

Best of three command wall times, including Go startup:

| Rule | Original | Cold | Warm, one-byte edit | Warm, unchanged |
|---|---:|---:|---:|---:|
| no-var | [113.261](no-var-3-before.log) | [56.590](no-var-3-cold.log) | [54.706](no-var-3-warm-byte-change.log) | [7.555](no-var-2-warm-unchanged.log) |
| no-empty | [105.095](no-empty-3-before.log) | [64.974](no-empty-2-cold.log) | [50.110](no-empty-3-warm-byte-change.log) | [8.720](no-empty-2-warm-unchanged.log) |
| eqeqeq | [104.274](eqeqeq-3-before.log) | [58.274](eqeqeq-2-cold.log) | [54.415](eqeqeq-2-warm-byte-change.log) | [7.066](eqeqeq-2-warm-unchanged.log) |

The lowest observed sanitized clang main.c build plus cached-runtime link was **16.089s**, in [eqeqeq-2-warm-byte-change.log](eqeqeq-2-warm-byte-change.log). Build-flags line: `commit=5f410bc4c7d3617ecee9347afa74dbe8db92082f; nproc=5; cpu.max=400000 100000; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; flags=-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all; GOCACHE=/home/agent/.cache/go-build`. Cache mode: cached; load before `1.41 1.18 1.22 1/153 87735`; after `1.24 1.16 1.21 1/153 89193`. An edited rule requires the correct build and its owned-mutant build, so the under-ten-second edited target was not reached. Compilation work stopped at this floor as requested. The five registered rules on this base are all required baselines; a completed fifty-rule fleet was not measured.

All three original repetitions and exact instruments:

| Loop | Before | After | Instrument |
|---|---:|---:|---|
| no-var, cold, best of 3 | 113.261 | 56.590 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-var` |
| no-var, warm-byte-change, best of 3 | 113.261 | 54.706 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-var -rule-byte-change` |
| no-var, warm-unchanged, best of 3 | 113.261 | 7.555 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-var` |
| no-empty, cold, best of 3 | 105.095 | 64.974 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-empty` |
| no-empty, warm-byte-change, best of 3 | 105.095 | 50.110 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-empty -rule-byte-change` |
| no-empty, warm-unchanged, best of 3 | 105.095 | 8.720 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-empty` |
| eqeqeq, cold, best of 3 | 104.274 | 58.274 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule eqeqeq` |
| eqeqeq, warm-byte-change, best of 3 | 104.274 | 54.415 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule eqeqeq -rule-byte-change` |
| eqeqeq, warm-unchanged, best of 3 | 104.274 | 7.066 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule eqeqeq` |

Original instruments (the Go overlay restores the complete base lint_test.go and excludes only the new test files):

no-var: original first/repetitions = 134.250s, 124.533s, 113.261s.

```sh
go test -overlay=/tmp/adamic-lint-before-overlay.json ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestMutants)$/^var_declaration_suppressed$' -count=1 -timeout 30m -v
```

no-empty: original first/repetitions = 118.485s, 109.087s, 105.095s.

```sh
go test -overlay=/tmp/adamic-lint-before-overlay.json ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestMutants)$/^empty_function_body_reported$' -count=1 -timeout 30m -v
```

eqeqeq: original first/repetitions = 129.790s, 108.622s, 104.274s.

```sh
go test -overlay=/tmp/adamic-lint-before-overlay.json ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestMutants)$/^suggestion_applied_as_fix$' -count=1 -timeout 30m -v
```

Raw repetitions, with the shared build-flags line above and per-row cache mode/load measurements:

| Rule / round / mode | Seconds | Cache mode | Load before | Load after | Log |
|---|---:|---|---|---|---|
| no-var / 1 / before | 134.250 | original uncached harness | 0.17 0.92 1.59 1/153 61326 | 1.05 1.03 1.54 1/156 62074 | [no-var-1-before.log](no-var-1-before.log) |
| no-var / 1 / cold | 74.622 | empty lint cache | 1.05 1.03 1.54 1/156 62074 | 1.17 1.05 1.50 4/156 63796 | [no-var-1-cold.log](no-var-1-cold.log) |
| no-var / 1 / warm-byte-change | 54.742 | cached | 1.17 1.05 1.50 1/156 63796 | 1.07 1.04 1.47 1/154 65260 | [no-var-1-warm-byte-change.log](no-var-1-warm-byte-change.log) |
| no-var / 1 / warm-unchanged | 10.998 | cached | 1.07 1.04 1.47 1/154 65260 | 1.06 1.04 1.47 1/154 66167 | [no-var-1-warm-unchanged.log](no-var-1-warm-unchanged.log) |
| no-empty / 1 / before | 118.485 | original uncached harness | 1.06 1.04 1.47 2/154 66167 | 1.04 1.03 1.41 2/154 66841 | [no-empty-1-before.log](no-empty-1-before.log) |
| no-empty / 1 / cold | 72.197 | empty lint cache | 1.04 1.03 1.41 1/154 66841 | 0.96 1.02 1.38 1/152 68559 | [no-empty-1-cold.log](no-empty-1-cold.log) |
| no-empty / 1 / warm-byte-change | 64.665 | cached | 0.96 1.02 1.38 1/152 68559 | 1.16 1.07 1.37 1/152 70022 | [no-empty-1-warm-byte-change.log](no-empty-1-warm-byte-change.log) |
| no-empty / 1 / warm-unchanged | 10.379 | cached | 1.16 1.07 1.37 2/152 70022 | 1.13 1.07 1.36 1/155 70931 | [no-empty-1-warm-unchanged.log](no-empty-1-warm-unchanged.log) |
| eqeqeq / 1 / before | 129.790 | original uncached harness | 1.13 1.07 1.36 1/155 70931 | 1.20 1.12 1.35 1/153 71622 | [eqeqeq-1-before.log](eqeqeq-1-before.log) |
| eqeqeq / 1 / cold | 69.206 | empty lint cache | 1.20 1.12 1.35 1/153 71622 | 1.14 1.12 1.33 1/154 73354 | [eqeqeq-1-cold.log](eqeqeq-1-cold.log) |
| eqeqeq / 1 / warm-byte-change | 59.587 | cached | 1.14 1.12 1.33 1/154 73354 | 1.10 1.11 1.31 1/154 74819 | [eqeqeq-1-warm-byte-change.log](eqeqeq-1-warm-byte-change.log) |
| eqeqeq / 1 / warm-unchanged | 8.489 | cached | 1.10 1.11 1.31 1/154 74819 | 1.09 1.11 1.31 1/154 75721 | [eqeqeq-1-warm-unchanged.log](eqeqeq-1-warm-unchanged.log) |
| no-var / 2 / before | 124.533 | original uncached harness | 1.09 1.11 1.31 2/154 75721 | 1.11 1.12 1.29 1/154 76413 | [no-var-2-before.log](no-var-2-before.log) |
| no-var / 2 / cold | 69.408 | empty lint cache | 1.11 1.12 1.29 1/154 76413 | 1.14 1.12 1.27 1/154 78138 | [no-var-2-cold.log](no-var-2-cold.log) |
| no-var / 2 / warm-byte-change | 60.107 | cached | 1.14 1.12 1.27 1/154 78138 | 1.13 1.13 1.26 1/153 79604 | [no-var-2-warm-byte-change.log](no-var-2-warm-byte-change.log) |
| no-var / 2 / warm-unchanged | 7.555 | cached | 1.13 1.13 1.26 1/153 79604 | 1.11 1.12 1.26 1/153 80504 | [no-var-2-warm-unchanged.log](no-var-2-warm-unchanged.log) |
| no-empty / 2 / before | 109.087 | original uncached harness | 1.11 1.12 1.26 2/153 80504 | 1.07 1.10 1.23 2/154 81191 | [no-empty-2-before.log](no-empty-2-before.log) |
| no-empty / 2 / cold | 64.974 | empty lint cache | 1.07 1.10 1.23 1/154 81191 | 1.20 1.12 1.23 1/153 82919 | [no-empty-2-cold.log](no-empty-2-cold.log) |
| no-empty / 2 / warm-byte-change | 51.441 | cached | 1.20 1.12 1.23 1/153 82919 | 1.08 1.10 1.22 1/153 84380 | [no-empty-2-warm-byte-change.log](no-empty-2-warm-byte-change.log) |
| no-empty / 2 / warm-unchanged | 8.720 | cached | 1.08 1.10 1.22 1/153 84380 | 1.07 1.10 1.22 1/153 85283 | [no-empty-2-warm-unchanged.log](no-empty-2-warm-unchanged.log) |
| eqeqeq / 2 / before | 108.622 | original uncached harness | 1.07 1.10 1.22 2/153 85283 | 1.06 1.09 1.20 1/155 85970 | [eqeqeq-2-before.log](eqeqeq-2-before.log) |
| eqeqeq / 2 / cold | 58.274 | empty lint cache | 1.06 1.09 1.20 1/155 85970 | 1.41 1.18 1.22 1/153 87735 | [eqeqeq-2-cold.log](eqeqeq-2-cold.log) |
| eqeqeq / 2 / warm-byte-change | 54.415 | cached | 1.41 1.18 1.22 1/153 87735 | 1.24 1.16 1.21 1/153 89193 | [eqeqeq-2-warm-byte-change.log](eqeqeq-2-warm-byte-change.log) |
| eqeqeq / 2 / warm-unchanged | 7.066 | cached | 1.24 1.16 1.21 1/153 89193 | 1.22 1.16 1.21 1/153 90095 | [eqeqeq-2-warm-unchanged.log](eqeqeq-2-warm-unchanged.log) |
| no-var / 3 / before | 113.261 | original uncached harness | 1.22 1.16 1.21 2/153 90095 | 1.03 1.11 1.17 1/155 90797 | [no-var-3-before.log](no-var-3-before.log) |
| no-var / 3 / cold | 56.590 | empty lint cache | 1.03 1.11 1.17 1/155 90797 | 1.12 1.12 1.18 1/154 92512 | [no-var-3-cold.log](no-var-3-cold.log) |
| no-var / 3 / warm-byte-change | 54.706 | cached | 1.12 1.12 1.18 1/154 92512 | 1.09 1.11 1.17 1/155 93975 | [no-var-3-warm-byte-change.log](no-var-3-warm-byte-change.log) |
| no-var / 3 / warm-unchanged | 8.318 | cached | 1.09 1.11 1.17 2/155 93975 | 1.15 1.12 1.17 1/158 94881 | [no-var-3-warm-unchanged.log](no-var-3-warm-unchanged.log) |
| no-empty / 3 / before | 105.095 | original uncached harness | 1.15 1.12 1.17 2/158 94881 | 1.30 1.15 1.18 2/156 95568 | [no-empty-3-before.log](no-empty-3-before.log) |
| no-empty / 3 / cold | 70.888 | empty lint cache | 1.30 1.15 1.18 2/156 95568 | 1.41 1.22 1.20 1/153 97283 | [no-empty-3-cold.log](no-empty-3-cold.log) |
| no-empty / 3 / warm-byte-change | 50.110 | cached | 1.41 1.22 1.20 2/153 97283 | 1.18 1.18 1.18 1/152 98747 | [no-empty-3-warm-byte-change.log](no-empty-3-warm-byte-change.log) |
| no-empty / 3 / warm-unchanged | 10.644 | cached | 1.18 1.18 1.18 2/152 98747 | 1.15 1.17 1.18 1/152 99650 | [no-empty-3-warm-unchanged.log](no-empty-3-warm-unchanged.log) |
| eqeqeq / 3 / before | 104.274 | original uncached harness | 1.15 1.17 1.18 1/152 99650 | 1.43 1.30 1.22 1/153 100333 | [eqeqeq-3-before.log](eqeqeq-3-before.log) |
| eqeqeq / 3 / cold | 62.372 | empty lint cache | 1.43 1.30 1.22 1/153 100333 | 1.74 1.46 1.28 1/153 102052 | [eqeqeq-3-cold.log](eqeqeq-3-cold.log) |
| eqeqeq / 3 / warm-byte-change | 61.351 | cached | 1.74 1.46 1.28 1/153 102052 | 1.20 1.35 1.25 1/153 103527 | [eqeqeq-3-warm-byte-change.log](eqeqeq-3-warm-byte-change.log) |
| eqeqeq / 3 / warm-unchanged | 8.324 | cached | 1.20 1.35 1.25 2/153 103527 | 1.18 1.34 1.25 1/153 104429 | [eqeqeq-3-warm-unchanged.log](eqeqeq-3-warm-unchanged.log) |

Reproduce the sequence after sourcing the toolchain:

```sh
python3 stage1/cohere/lint/check-evidence/measure.py
```

**Cache mutants, actually run**

Each mutant drops exactly one component, changes a real guarded input behind a populated cache, and compares the cached answer against an independent uncached implementation. The enclosing test requires failure containing both `stale <kind> answer` and `cached_matches_before=true`. [The complete uncached gate log](full-lint-package-uncached.log) contains every failing subprocess and its byte counts.

| Cache | Dropped component | Guarded input changed | Old = cached / fresh bytes | Caught by |
|---|---|---|---|---|
| Oracle binary | oracle/adapter/no-var | Private Go adapter changes NoVar.Run to no listeners | 670 / 34 | TestLintCacheInvalidation/oracle |
| Upstream capture | capture/filter | TestNoVarFires becomes TestNoVarStaysSilent | 1280 / 1506 | TestLintCacheInvalidation/capture |
| Go observations | go-output/manifest | Source changes var to let at the same filename | 673 / 34 | TestLintCacheInvalidation/go-output |
| Native port build | port/modules | Private main gains an observable print | 668 / 682 | TestLintCacheInvalidation/port |
| Source Node observations | node/modules | Private main gains an observable print | 668 / 682 | TestLintCacheInvalidation/node |
| Emitted JavaScript observations | javascript/modules | Emitted module gains an observable print | 674 / 688 | TestLintCacheInvalidation/javascript |
| Emitted JavaScript build | javascript-build/modules | Private main gains an observable print | 521585 / 521615 | TestLintCacheInvalidation/javascript-build |

TestInheritedAllSelection additionally suppresses a rule only under `all`: the selected owned witness still agrees with Go, but an original inherited `all` row catches it. TestSelectedRegistrationKeepsImportedHelpers creates an unregistered helper rule in a private copy and proves that changing its imported module changes the module identity. No owned rule file is edited.

Each author check runs the owned mutant. The three measured mutants are `var declaration suppressed`, `empty function body reported`, and `suggestion applied as fix`. Every runtime catches each one in each sample log.

**Verification**

```sh
go test ./stage1/cohere/lint -run '^(TestLintCacheInvalidation|TestLintCacheMutants|TestLintCacheBypassAndIntegrity|TestSelectedRuleParity|TestInheritedAllSelection|TestSelectedRegistrationKeepsImportedHelpers)$' -count=1 -timeout 30m -v
```

PASS: [cache/parity/regression log](cache-parity-regressions.log). Selected and full registration findings are byte-identical for no-var, no-empty and eqeqeq.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestRule|TestSelectedRuleUncachedAnswers|TestInheritedAllSelection|TestSelectedRegistrationKeepsImportedHelpers)$' -count=1 -timeout 30m -v -args -rule no-empty
```

PASS: [uncached comparison log](uncached-comparison.log). Cached and uncached Go/Node/emitted-JavaScript/native observations are byte-identical on the complete no-var corpus, 427484 bytes. Existing native runtime-library cache behavior is unchanged; every new lint artifact/observation cache bypasses reads and writes under ADAMIC_GATE_UNCACHED=1. The cache unit probes explicitly set the variable back to 0 to exercise invalidation.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint ./stage1/cohere/lint/registry ./cmd/adamic-lint-check -count=1 -timeout 30m -v
```

Complete touched-package gate: PASS. The existing harness, legacy mutants, all owned mutants, volume mutants, profiling compilation and count guards passed. Opt-in TestCompilerAndStage1Agree, TestThroughput, TestProfileArtifacts and TestProfileSnapshotsAgree skipped because their external inputs were not configured. TestRule skipped here without -rule; its 27 measured author checks ran separately. [Log](full-lint-package-uncached.log).

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^load$/^testdata$/^0[.]1$/^compile$/^07_modules$/^main[.]ts$' -count=1 -timeout 30m -v
go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry ./cmd/adamic-lint-check
```

PASS: [filtered compiler oracle](filtered-compiler-oracle.log), [vet](vet.log) and [registry/command compilation](registry.log). All test output was redirected to log files.

**Toolchain setup and limits**

Setup command: `ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh`; shells source `/workspace/adamic-tools/env.sh`. nproc=5, with a four-core cgroup quota. The initial [setup log](setup-initial.log) reports Go ready (0s), clang ready (1s), Node ready (1s), submodules ready (1s), then a warm-build failure: undefined wrapper functions while files were being edited. The installed tools were usable, and subsequent builds passed. Setup was rerun successfully after measurements, with its build cache warm at 48s. The final instrumented run passed after the package gate: Go, clang, Node and submodules ready at 0s, build cache warm at 28s, nproc=5 and cpu.max=400000 100000; [final setup log](setup-final.log) records its timing lines and [setup-flags.json](setup-flags.json) records the source commit, tools, nproc, quota, cache mode and loads before/after. The first successful rerun is preserved in [setup-final-uninstrumented.log](setup-final-uninstrumented.log).

Not run: complete `go test ./...` integration, an actual fifty-rule registered fleet, opt-in external compiler profiling snapshots, Windows/macOS cache behavior, adversarial cross-process cache corruption, or per-file compilation. No correctness check was removed to reach the reported numbers. The complete repository integration gate remains an uncached requirement before main moves.

Final setup build-flags line: source commit `5f410bc4c7d3617ecee9347afa74dbe8db92082f`, go version go1.27.1 linux/amd64, clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261), Node v24.19.0, nproc=5, cpu.max=400000 100000, probe flags `-fsanitize=address,undefined -g`, cache mode existing Go build cache, load before `0.97 1.21 1.25 1/153 114687`, after `2.84 1.63 1.39 1/154 115444`. Exact command: `ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh`.

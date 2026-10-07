Built splitPath, HasPathSegment and gapRootAcceptsModifierOnArbitrary in separate .a files; sixty-three owned helpers total.
Commits: claim 81c0350d pushed before source; final publication SHA is named in the final response; current lint area d3a37422 contains current main b6b1538b.
Commands: helper gate PASS 98.389s, 6,453 four-way comparisons; upstream consumer capture FAIL with eight missing-input checks; vet/format clean; six uncached Node probes PASS 2.962s; setup 52s, nproc 5.
Mutants: all fourteen new compiling semantic variants caught by exact Go output mismatch, plus the independent missing-consumer check; each exact mutation and witness is recorded below and in evidence/helpers.log.
Not covered: live Tailwind engine/corpus population, complete rule findings/fixes/suggestions, arbitrary malformed adapters, or the full repository gate including its seventeen required external comparisons.

# Landing and ownership

All sixty previous helpers were complete and pushed at c4c09485 before reservation. The first fifty-seven were green in the complete 1,264.599s current-base gate, with 3,054,680 comparisons and 230 owned variants; batch twenty added 4,587 comparisons and twenty-four variants in a 139.993s gate. Both current integration bases are unchanged. Prior source remains unchanged; this unit uses the bounded new-helper package and filtered external oracle. Only codex/lint-helpers-03 is this worker's pushed branch.

Fresh wildcard fetch inspected all twenty origin helper branches and nineteen claims files. These three helpers tie the highest eligible unclaimed fan-out at three consumers each. Comments are already reserved by the shared HELPERS.md bundle. Regexp-engine internals are excluded by the JS RegExp instruction. Claim 81c0350d was pushed before implementation. ownership.json preserves the inspected remote claims. No fourth helper is reserved. The territory is claims/03.md, slot03/batch21 and batch21_test.go. No shared harness, registry, compiler, runtime or actual cohere worktree file is edited.

# Consumer rules and helper readiness

- splitPath: @next/next/no-head-import-in-document; @next/next/no-page-custom-font; @next/next/no-styled-jsx-in-document.
- HasPathSegment: nexus/boundary-no-internal-import; nexus/import-require-path-alias; nexus/localization-no-untranslated-value.
- gapRootAcceptsModifierOnArbitrary: better-tailwindcss/enforce-consistent-class-order; better-tailwindcss/enforce-shorthand-classes; better-tailwindcss/no-unknown-classes.

Nine dependency edges across nine rules. Subtracting only this batch from the frozen inventory removes 0 final helper blockers. readiness.json lists remaining dependencies. These are helper-readiness counts, not claims of complete rules or successful full consumer-package tests.

# Actual Go comparison

Pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. The capture contains 264 deduplicated fixture inputs and validates the exact nine-consumer set. Observation-only overlays record ordinary fixtures, typed multi-file fixtures, Tailwind wrapper inputs before external-engine checks and localization's custom multi-file runner. All original package tests still execute. Next.js and Nexus pass. Tailwind fails the eight checks below. coverage.json records capture_gate_exit 1 and regeneration exits nonzero even after writing captured data. No failure is converted into a pass.

The oracle probes 535 distinct raw/rooted filenames, source texts and actual Go AST string-literal values, plus byte boundary paths. For each path it compares Go splitPath and records the exact lastSeparator byte arguments and call order. It also compares HasPathSegment against ten segments, including all three fixed strings the consumer rules use, empty strings, multibyte text, invalid UTF-8 and segments containing separators. Outputs retain exact bytes. The .a driver uses the previously proven owned lastSeparator helper; this comparison therefore exercises that composition without replayed answers.

The arm helper executes on all 57 Go GapRootDescriptions and every color mask through eight arms: 568 cases. The controls vary the unrelated AcceptsModifierOnArbitrary field to show the helper reads the arms. Nil and empty Arms both answer false. Nil descriptions are outside the successful-input adapter contract, because Go dereferences them. Arbitrary longer arm lists are not exhaustively enumerated.

535 split outputs plus 5,350 segment outputs plus 568 arm outputs equal 6,453 rows. Every row agrees with real Go on source Node, emitted JavaScript and sanitized native. Semantic variants must compile, finish with exit zero and empty stderr before their output mismatch counts. Compiler refusals, panics or sanitizer failures do not count.

The first observer attempt supplied relative filenames and Go's real parser correctly refused them. initial-parser-path-failure.log preserves the error. The adapter now uses canonical rooted filenames for parsing while retaining raw bytes for path queries. A console type-check refusal was fixed by emitting strings. Neither was credited as a mutant catch. All fourteen final variants are held to unchanged Go helpers.

# Commands and missing-input blocker

Source /workspace/adamic-tools/env.sh. All tests write stdout/stderr directly to logs.

```
bash cloud/setup.sh > /tmp/slot03-batch21-setup.log 2>&1
python3 stage1/cohere/lint/helpers/slot03/batch21/testdata/regenerate.py > stage1/cohere/lint/helpers/slot03/batch21/evidence/regeneration.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch21' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch21/evidence/helpers.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot03/batch21/evidence/vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch21/evidence/input-oracle.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestConsumerCoverageRejectsMutant$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch21/evidence/coverage-mutant.log 2>&1
```

Setup timing lines: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 52s, done 52s on five processors; nproc 5, cgroup cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Helper gate PASS 98.389s with fourteen semantic catches and no selected skips. Vet/format logs are empty. Six uncached Node probes pass 2.962s with zero probe hits and six misses. The independent consumer-set mutant removes better-tailwindcss/no-unknown-classes and is caught.

The upstream capture command remains FAILED. The live Tailwind installation and ahra stylesheet/corpus are absent from /Users/kirkouimet/Projects/ahra/app/_theme/styles. No supported environment override is supplied by these tests. The failing checks are:

- TestConflictFixturesActuallyRan
- TestUnknownClassFixturesActuallyRan
- TestCanonicalClassesPlacementIsAccountedFor
- TestClassOrderFixturesActuallyRan
- TestCanonicalFixturesActuallyRan
- TestConflictingClassesPlacementIsAccountedFor
- TestUnknownClassesPlacesEveryCorpusClass
- TestClassOrderLiveMatchesTheEngineOverTheCorpus

An unmodified upstream reproducer, run from cohere, is:

```
go test ./internal/lint/rules/tailwind -run '^TestUnknownClassFixturesActuallyRan$' -count=1 -v -timeout=5m > /workspace/adamic/stage1/cohere/lint/helpers/slot03/batch21/evidence/missing-tailwind-reproducer.log 2>&1
```

It fails because no installed tailwindcss is found at the required search root. The full capture additionally fails expected live populations of 7,287 canonical classes, 11,389 conflict resolutions, 11,314 unknown-class inputs and nonzero class-order comparisons, all observed as zero. Supplying the real engine and corpus is required to make those consumer checks meaningful; inventing substitute inputs or changing their guards would not verify those claims. No shared files are edited, no failing check is skipped, relaxed or deleted, and no further helpers are claimed after this blocker.

# Every mutant

| Variant | Exact replacement |
|---|---|
| split_path.a | `index >= 0` -> `index > 0` |
| split_path.a#01 | `filePath.slice(index + 1)` -> `filePath.slice(index)` |
| split_path.a#02 | `parentIndex >= 0` -> `parentIndex > 0` |
| split_path.a#03 | `parent.slice(parentIndex + 1)` -> `parent.slice(parentIndex)` |
| split_path.a#04 | `lastSeparator(parent)` -> `lastSeparator(filePath)` |
| has_path_segment.a | `segment.length === 0` -> `false` |
| has_path_segment.a#01 | `byte === 92` -> `false` |
| has_path_segment.a#02 | `byte === 47` -> `false` |
| has_path_segment.a#03 | `index <= path.length` -> `index < path.length` |
| has_path_segment.a#04 | `equal = false;` -> `equal = true;` |
| has_path_segment.a#05 | `start = index + 1` -> `start = index` |
| gap_root_accepts_modifier.a | `if (isColor)` -> `if (!isColor)` |
| gap_root_accepts_modifier.a#01 | `return false;` -> `return true;` |
| gap_root_accepts_modifier.a#02 | `for (const isColor of arms)` -> `for (const isColor of arms.slice(0,1))` |

# Limits

This is helper composition evidence, not full rule findings/fixes/suggestions parity. Parsed input/callback adapters must obey the byte and non-null-description contracts in README.md. The full repository gate and its seventeen required external correctness checks were not run. The failed upstream Tailwind gate remains a reported blocker and is never presented as green. All implementation and compact evidence are published only to this worker's branch; no PR, main push or area push.

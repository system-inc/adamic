Built: DesignSystem.HasVariant, VariantKind and VariantCompoundsWith, one helper per .a file; 18 prerequisite removals across six rules.
Commits: implementation e5bd2f4; reservation 8890cfc pushed before code; all twenty retained slot 04 helpers are tested and pushed.
Checks: touched package PASS 20.203s; uncached filtered input oracle PASS 1.570s; vet/format clean; session setup 165s, nproc 5.
Mutants: value-dependent membership, wrong missing-kind default, empty-kind collapse, wrong parent kind, child-dependent compounds; all compiled and were caught against Go on source Node, sanitized native and emitted JavaScript. Consumer omission caught.
Limits: no full repository gate or production integration; preserve Go's child-independent compound approximation; decoded kind-string and membership/kind registry view; zero final blockers removed.

## Exact consumers

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.DesignSystem.HasVariant

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.DesignSystem.VariantKind

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.DesignSystem.VariantCompoundsWith

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

The frozen original readiness.json is unchanged. This local derived ledger removes only the twenty uniquely owned slot 04 helpers and assumes no integration of other workers' code. Original 46 helper-ready rules become 47 cumulatively, solely adding @next/next/no-img-element in the first batch. These are dependency counts, not completed rule ports.

## Observations and contract

Pinned Cohere: 715ba94f3608a6500086b1076ce5cb7e51b836db. Temporary Go overlays call the actual LoadedDesignSystem methods against a real VariantRegistry, not copied expected algorithms. The state view contains only root membership and stored kind, the complete set of registration data these methods read. HasVariant uses map presence; VariantKind uses map lookup and an absent-only static fallback; VariantCompoundsWith tests a present parent's kind against compound. Registry order is not read. Zero-valued registrations remain present with empty kind, and every helper observes live deletion, insertion and kind replacement.

Go's VariantCompoundsWith explicitly ignores the child argument. It approximates Tailwind's compound bitmask calculation by checking the parent alone. This port preserves that documented Go behavior: a registered compound parent accepts even an arbitrary at-rule-only child that upstream Tailwind could reject. No upstream mask/selector computation is claimed. Tests vary six child kinds, including unknown and zero kind, without changing the parent result.

- 569 witness/state controls produce 6828 matching lines across real Go, source Node, sanitized native and emitted JavaScript. They include empty/missing/Unicode/malformed-byte roots, all four actual kinds, empty and unknown kinds, duplicate roots with last-write semantics, and 500 deterministic random byte strings.
- All 149 unique captured source fixtures cover each consuming rule. Raw source and Go-parser string/template values yield 3492 matching direct-probe lines. This demonstrates helper input coverage, not full Adamic rule integration.
- Instrumentation captures 141 actual method calls, deduplicated to 31 tagged complete registry/query states: 18 HasVariant, 11 VariantKind and 2 VariantCompoundsWith. Each captured state is probed through all three actual Go methods; 372 observation lines match. Children need not be captured because Go never reads them; separate controls vary them explicitly.
- Empty roots, every singleton byte and all 65536 byte pairs give 789516 matching observation lines. The adapter uses injective decimal-byte root encoding to preserve arbitrary Go string equality without lossy UTF-8 decoding. Every case also deletes the queried root, inserts an empty-kind registration, replaces it with compound, then replaces it with functional.
- The input oracle passes six probes uncached, with zero cache hits and six probe misses. Vet emits no diagnostics, and gofmt lists no files.

## Every mutant

Every credited semantic mutant compiles, executes successfully and differs from actual Go in all three Adamic execution modes.

- HasVariant rejects a registered empty kind: first mismatch line 1, false instead of true.
- VariantKind uses functional for an absent key: line 9, deleted false|functional|false instead of deleted false|static|false.
- VariantKind converts an empty stored kind to static: line 2, kind static instead of the empty kind.
- VariantCompoundsWith accepts functional instead of compound parent kind: line 11, false instead of true for an inserted compound parent.
- VariantCompoundsWith depends on child kind being static: line 11, false instead of true for an arbitrary child under a compound parent.
- Coverage omission removes all canonical-classes fixture rows; the readiness-derived assertion detects the missing consumer.

The initial empty-kind mutant used string logical-OR, which stage 0 could not lower. It was replaced by an explicit if branch expressing the same semantic defect; only the successful compiling mutant is credited. The state type initially allowed replacing its read-only map through a mutable field; the compiler correctly refused mutable variance. Marking the view's variants field readonly expresses the intended read-only API without changing shared compiler code.

## Commands, ownership and limits

README.md records exact commands. Test stdout/stderr goes directly to log files. evidence/tests.log contains every successful comparison and mutant, evidence/capture.log the state counts, and evidence/tailwind-capture-tailwind.log the filtered upstream fixture gate. evidence/oracle.log, vet.log and format.log contain the other checks. No shared harness/generator or protected compiler file was edited, and temporary Go overlays leave the Cohere worktree unchanged.

Session toolchain setup previously passed: Go 0s, clang 1s, Node 1s, submodules 2s, warm 165s, total 165s; nproc 5, cgroup cpu.max 400000 100000, memory 17.6GB. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0. This continuation sources /workspace/adamic-tools/env.sh and verifies nproc 5; setup was not rerun.

Initially fetched every visible helper branch before reserving these highest-count unclaimed ties. Subsequent wildcard fetches discovered additional helper branches; the final check covered all seventeen current origin codex/lint-helpers* branches and every claim path. Our 8890cfc reservation at 02:29:54 UTC precedes slot 02 3552136 at 02:30:20 for HasVariant/VariantKind; slot 02 explicitly withdrew both without writing code. It also precedes wave1-15 4f3c8f1 at 02:31:36 for VariantKind; that later claim remains in its remote file at the final check. Our earlier reservation is retained under the established earliest-claim precedence. No additional helper is reserved.

The full repository gate was not run, and unchanged earlier slot packages were not rerun. Production integration, registry creation/registration, full compound masks and CSS loading remain separate work. The registry view preserves only fields read by these methods. Kind strings use the decoded valid-UTF-8 boundary; malformed-byte kind payloads are outside this view, while roots preserve arbitrary raw bytes through lossless identity encoding. Capture selects rule fixture tests and excludes unrelated external live-population tests.

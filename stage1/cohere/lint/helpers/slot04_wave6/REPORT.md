Built: Theme.PrefixKey, VariantRegistry.Has and DesignSystem.Prefix, one helper per .a file; 18 prerequisite removals across six rules.
Commits: implementation 56b0cd3; reservation d31400c was pushed before implementation; all seventeen retained slot 04 helpers are tested and pushed.
Checks: touched package PASS 23.572s; uncached filtered input oracle PASS 1.851s; vet/format clean; session setup 165s, nproc 5.
Mutants: one-byte slice offset, wrong separator, value-dependent membership, empty prefix getter, disabled short-key guard; all compiled and were caught against Go on source Node, sanitized native and emitted JavaScript. Consumer omission caught.
Limits: no full repository gate or production integration; immutable raw-byte theme view and membership-only registry view; captured rule states have only empty prefixes; zero final blockers removed.

## Exact consumers

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*Theme.PrefixKey

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*VariantRegistry.Has

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.DesignSystem.Prefix

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

The frozen original readiness.json is unchanged. This directory's derived ledger removes only the seventeen uniquely owned slot 04 helpers; other workers' implementations are not assumed integrated. Original 46 helper-ready rules become 47 cumulatively, solely from the first batch's @next/next/no-img-element. These are dependency counts, not completed rule implementations.

## Evidence

Pinned Cohere: 715ba94f3608a6500086b1076ce5cb7e51b836db. Temporary Go overlays invoke actual Theme.PrefixKey, VariantRegistry.Has and LoadedDesignSystem.Prefix methods. The prefix getter is observed before and after replacing the same theme object's prefix. Registry membership is observed before deletion, after deletion, and after insertion of a zero-valued registration. Adamic marker payloads are always false so a value-dependent membership bug cannot pass.

- 531 witness/state rows produce 3186 matching lines in Go, source Node, sanitized native and emitted JavaScript. Controls include empty prefixes and short keys, missing roots, empty registries, nonempty prefixes, arbitrary invalid UTF-8, and a key whose UTF-8 byte boundary differs from a character boundary, plus 500 deterministic random raw-byte strings.
- All 149 captured unique source fixtures cover the six consuming rules. The actual Go parser extracts source and string/template literal text, yielding 1746 matching lines in all modes. These source-derived direct probes are not a claim of full Adamic rule integration.
- Temporary instrumentation captures 58553 actual helper calls, deduplicated to 442 full states. PrefixKey pairs preserve both prefix and key; registry calls preserve the complete key set and query. Every captured prefix is empty, 421 distinct states have nonempty keys, and 20 have nonempty registry snapshots. All 2652 direct-probe observation lines match. The nonempty-prefix behavior is held by the separate controls and exhaustive raw-byte corpus.
- The empty string, all 256 singleton bytes and all 65536 byte pairs produce 394758 matching lines. This exhaustive corpus reads every prefix, replaces it with byte 137, probes prefix-key construction on applicable keys, and mutates registry membership.
- PrefixKey with nonempty prefix and key length zero or one causes actual Go's slice-bounds panic. Both inputs are separately refused with a diagnostic and unsuccessful process status in source Node, sanitized native and emitted JavaScript. The exhaustive normal-output corpus labels those short-key combinations without invoking them; it does not substitute matching fabricated values for a panic.

## Every mutant

Each semantic mutant successfully compiles, runs successfully and differs from the real Go result. Compile failures are not credited as mutant catches.

- Byte offset 2 changed to 1: all three modes differ at line 14 by retaining an extra hyphen.
- Prefix separator hyphen changed to underscore: all three modes differ at line 14.
- Map presence changed to get(root) === true: all three modes differ at line 6 when a false marker is present.
- Prefix getter changed to an empty slice: all three modes differ at line 3 after the theme's prefix changes to byte 137.
- Short-key guard disabled by changing length < 2 to length < 0: all three modes compile and return the fabricated --tw- bytes for both invalid lengths, whereas Go panics. The independent refusal check catches this successful but wrong result.
- Consumer omission: removing every canonical-classes fixture is caught by the readiness-derived consumer coverage assertion.

## Commands, ownership and limits

README.md gives exact commands. All test stdout/stderr is written directly to log files. evidence/tests.log records successful four-way comparisons and every mutant, evidence/capture.log the actual-call counts, and evidence/tailwind-capture-tailwind.log the filtered upstream fixture gate. evidence/oracle.log records the six uncached input-oracle probes with zero cache hits and six misses; vet.log and format.log are empty.

Session toolchain setup previously passed: Go 0s, clang 1s, Node 1s, submodules 2s, warm 165s, total 165s; nproc 5, cgroup cpu.max 400000 100000, memory 17.6GB. Tool versions: Go 1.27.1, clang 20.1.8, Node 24.19.0. This continuation sources /workspace/adamic-tools/env.sh and verifies nproc 5; setup was not rerun.

All six origin codex/lint-helpers* branches were fetched and their claims checked before the reservation. Higher-count comment helpers were already delivered by the base branch under stage1/cohere/lint/HELPERS.md, so the highest unclaimed concrete helpers tie at six. Post-reservation and final rechecks show no competing claims for these three. No additional helper is reserved.

Immutable byte-string theme views preserve arbitrary Go string data without Unicode conversion. Callers must not mutate backing arrays, though replacing the theme prefix is supported. PrefixKey's no-prefix branch returns the original immutable bytes; the prefixed branch allocates a fresh sequence. It intentionally slices after two bytes without checking that the key begins with hyphens. Has uses a membership-only registry view because its actual Go behavior never reads registration payloads; the differential adapter's injective decimal-byte key encoding preserves equality for malformed UTF-8 too. Other design-system state, registry allocation/registration and production AST integration are outside this helper scope.

An initial local fixture used an untyped empty array in a conditional, which stage 0 could not lower. The fixture now uses an explicit number[] and pushes the one-byte case, requiring no shared compiler edits. The full repository gate was not run, and unchanged previous slot packages were not rerun. Capture excludes unrelated external live-population tests. No shared generator, harness, Cohere worktree source or protected compiler file changed.

Built: breakpointBucket, Comment and Declaration, one helper per .a file; 18 dependency removals across six rules.
Commits: implementation b18ecc9; claims a33e02d and 0647f19 pushed before implementation.
Checks: touched package PASS 19.100s; uncached filtered oracle PASS 2.040s; vet and format clean; setup 165s, nproc 5.
Mutants: wrong comment kind, comment value present, declaration value absent, property omitted, dot retained, function prefix discarded; all compiled and were caught against Go on source Node, sanitized native and emitted JavaScript. Consumer omission caught.
Limits: no full repository gate or production integration; raw immutable byte-string leaf adapter; zero actual Comment calls in captured rule fixtures; no final rule blockers removed.

## Consumers and readiness

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.Comment

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.Declaration

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.breakpointBucket

Six prerequisite removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

The frozen original readiness.json remains unchanged. The local ledger removes only fourteen uniquely owned slot 04 helpers, and credits no integration of other workers' helpers. Original 46 helper-ready rules become 47 cumulatively, solely adding @next/next/no-img-element from the first batch. These counts are dependency readiness, not finished rule implementations.

## Observations

Pinned Cohere: 715ba94f3608a6500086b1076ce5cb7e51b836db. Temporary Go overlays invoke the actual private breakpointBucket and public constructors, rather than copied expected behavior. All helpers compare exact bytes and every constructor field through a local leaf view, with fresh-allocation mutation checks.

- 526 witness rows: empty strings, ASCII boundaries, CSS functions, dots, digits, Unicode and 500 deterministic arbitrary-byte controls; 3156 output lines match in all modes.
- All 149 distinct captured source fixtures cover every consuming rule above. The oracle extracts raw source and parsed string/template literal values, giving 1746 observation lines. These source-derived inputs are direct helper probes, not a claim that Adamic rules are integrated.
- Instrumented actual Go rule calls: 136404 calls, 935 distinct tagged inputs. Declaration values contribute 454, declaration properties 479, breakpointBucket 2. No Comment calls occurred in this fixture subset. All 5610 direct-probe output lines match; the independent witness/exhaustive corpus covers comments.
- All 65793 empty, singleton and two-byte strings, including malformed UTF-8, match on 394758 lines in Go, source Node, sanitized native and emitted JavaScript.
- Six semantic mutants compile successfully, execute successfully and differ from Go. Comment kind first differs at line 2; comment presence line 2; declaration presence line 4; omitted property line 10; retained dot line 79; discarded function prefix line 85. Each is caught independently in all three modes. Removing all canonical-class consumer rows is caught by the readiness-derived coverage assertion.
- go vet emits no diagnostics and gofmt emits no filenames. The filtered input oracle passes with cache probe hits 0 and misses 6.

## Commands and evidence

See README.md for exact reproducible commands. evidence/tests.log records the four-way helper gate and every mutant. evidence/capture.log records input counts; evidence/tailwind-capture-tailwind.log records the upstream rule fixture gate; evidence/oracle.log, vet.log and format.log record the remaining checks. No test output was piped.

Session toolchain setup previously passed: Go 0s, clang 1s, Node 1s, submodules 2s, warm 165s, total 165s; nproc 5, cgroup cpu.max 400000 100000, memory 17.6GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. This continuation sources /workspace/adamic-tools/env.sh and verifies nproc 5; setup was not rerun.

## Ownership and boundaries

Fetched all six origin codex/lint-helpers* branches before reservations and rechecked afterward. Slot 02 c332c14 at 01:41:08 UTC precedes our a33e02d at 01:42:42 UTC for isValidThemePrefix and namespaceForVariantRoot. Those two local experiments were removed and are not delivered or counted. Replacement constructor reservations were pushed in 0647f19 before implementation. Our breakpointBucket reservation precedes slot 01 716e80b at 01:42:55 UTC; slot 01 withdrew the later duplicate. No further helper is reserved.

Inputs are immutable byte strings represented as readonly integer-byte arrays. Callers must not mutate their backing arrays. Go strings remain arbitrary bytes, so neither UTF-8 decoding nor UTF-16 string slicing occurs. Fresh mutable leaf objects preserve Go's constructor allocation and exact data, including declaration value presence for empty strings. Both nil-only fields use explicit false absence flags. The initial native probe rejected null lowering; this local representation resolves that without changing shared files. The leaf type describes only these constructors; container nodes, Context maps and full rule integration remain separate work.

The full repository gate was not run. Existing slot 04 packages were unchanged and were not rerun. No shared registration generator, harness, Cohere worktree source or protected compiler file was edited. Tailwind capture selects fixture tests and excludes unrelated external live-population tests, as in the previous batch. Captured helper inputs are not attributed to individual rules, so coverage uses separately captured rule sources and the independently exhaustive helper contract.

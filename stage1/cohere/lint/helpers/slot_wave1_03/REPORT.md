Built: Theme.Add in one .a helper file, with explicit separately owned store operations.
Commits: pre-code reservation c408e07a; implementation SHA is in the final response.
Commands and outputs: 14,780 cases, 7,180,060 identical bytes on Go, source Node, emitted JavaScript and sanitized native; vet exit 0; uncached input oracle PASS 6.217s.
Mutants: default precedence, overwrite order and directive error text each compile/run normally and are caught only by output comparison on all three paths.
Not covered: dependency callback integration, whole-rule findings/fixes, absent external repository corpora, arbitrary raw byte strings and the full repository gate.

# Theme.Add measured handoff

New branch starts from origin/codex/lint-helpers at 95100eb4. The previous twelve
rule implementations, tests and mutant evidence remain pushed on the original
rule branch. Every origin head was fetched before selection. The five distinct
helper claim files reserve all larger individual helpers; the 24-consumer comment
bundle is already implemented and reserved in HELPERS.md. Add ties the largest
remaining concrete fan-out, six consumers. Reservation c408e07a was pushed before
any implementation.

`bash cloud/setup.sh > evidence/setup.log 2>&1` passed: Go ready 0s, clang ready
0s, Node ready 0s, submodules ready 1s, cache warm 74s, done in 74s. `nproc` is 5;
cgroup quota is 400000/100000 and memory is 17.6 GB. Go 1.27.1, clang 20.1.8,
Node 24.19.0 and Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db.

`python3 .../slot_wave1_03/validate.py --scratch /tmp/adamic-helper-wave1-03-add-mutants > evidence/validation-mutants.log 2>&1`
passes. The complete owned package also passes in 31.983s via `go test ./stage1/cohere/lint/helpers/slot_wave1_03 -count=1 -v -timeout=20m > evidence/package.log 2>&1`. Its deterministic input sort changes only corpus order; the final canonical hash is d16e841e7a37980b06e02de098b49b09b634ba45b78e2c99db6af712c43383ba. Each consuming suite executes actual pinned Go rule expectations, then
419 distinct runtime Add arguments (421 for unknown classes) are replayed against
five controlled initial states: 12,580 consumer-derived states plus 2,200 boundary
controls. These are bounded helper observations, not 14,780 rule fixture findings.
Full byte equality and hashes are recorded in evidence/outputs.json; canonical Go
stdout and the full typed corpus are compressed without timestamps. All three
successful Adamic backends have empty stderr, including ASan/UBSan/leak checking.

Default precedence is checked before initial deletion, exactly as Go does. Clear
directives validate before any side effect, and continue to initial deletion after
clearing. Namespace arguments remove only the final `-*`. Overwrites keep their
existing insertion position. Error text includes the zero-offset prefix. Exact
callback names and arguments, all map entries, key order, dead count and prefix
are compared. The callbacks are not supplied the expected final Add state.

The first run skipped because Tailwind was absent. Installing pinned 4.3.3 in
scratch closes that environment gap. Broad test prefixes additionally matched
unrelated live repository corpus suites needing Kirk's absolute paths; those
failed runs are retained and are not credited. Final exact fixture-suite patterns
exclude those unavailable corpora while covering all six consumers. Driver-only
rewrites use the foundation's tagged readTextFile result, explicit numeric string
interpolation, statement-form panic and named callbacks. No compiler or shared
harness is edited. An initial broad default-precedence mutant hit an unexpected
dependency and is not credited; final isolated witnesses all finish normally.

## Rules losing one dependency each

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Six dependency entries are removed; zero additional final helper blockers are
removed by Add alone. The frozen readiness ledger is not changed. No rule is
claimed ported or fully integrated on account of this helper.

The installed base helper API stores decoded Unicode strings and numeric options.
Tests cover actual consumer option values plus bounded bit combinations 0..31,
not arbitrary signed-64-bit settings. Map/slice alias and capacity identity are
not compared. Full callback implementations belong to their reserved owners;
this unit compares their actual Go side effects through explicit input snapshots.
No malformed raw UTF-8, isolated UTF-16 surrogates, full rule findings/fixes,
concurrent store mutation or complete repository test gate is claimed.

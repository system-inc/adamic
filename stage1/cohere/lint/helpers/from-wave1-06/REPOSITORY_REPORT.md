Built collapse.*Table.addRepositoryStatics in one .a file, removing six prerequisites across six Tailwind rules and zero final blockers alone.
Commits: 2001f9b4 claim pushed and all origin claims rechecked before code; implementation commit follows this report.
Commands: parity PASS 12.50s, 2,147 cases, 169,528 identical bytes; source Node, emitted JavaScript and sanitized native held to real Go.
Mutants: skipped overwrite, wrong count, copied order storage and reused sort record all execute cleanly, then fail Go byte comparison on all three Adamic paths.
Not covered: full rule integration, independently compiled PropertySort, full repository gate, external Tailwind/corpus gate and reslicing of documented read-only order storage.

## Contract and observations

addRepositoryStatics takes the initialized table's mutable statics map, the nonnil system's repository body map and the separately owned PropertySort dependency. For every root it writes a fresh reading record containing the exact sorted count and shared order storage, overwriting framework collisions and preserving unrelated entries. Only this helper is implemented; PropertySort answers and body identities are supplied by the actual pinned Go dependency.

The private Go method is exposed by an owned overlay. One observational statement retains each root's real sorted record before map assignment; no algorithm or input is changed. Initial readings carry a distinguishable -99 count/-1 order. Snapshots observe every final map field and root, sorting valid UTF-8 names by Go byte order. Element mutations prove shared order storage; count mutations prove the result record is copied. The helper copies the reference to the documented read-only order list; resizing/reslicing and arbitrary slice capacities are outside this adapter boundary.

Cases cover all 890 framework declaration bodies plus three supplemental nil/empty/value-absent controls, names extracted from every one of 157 runtime fixture sources across all six consumers, duplicate-root last-write behavior, existing-root overwrites, untouched entries, no repository roots and Unicode/NUL/HTML/line-separator keys. Go directly converts declarations into nodes and executes PropertySort; those dependency results drive the isolated Adamic helper. All 2,147 cases yield 169,528 matching bytes on actual Go, original .a source Node, emitted JavaScript and ASan/UBSan native, with exit zero and no stderr.

The first overwrite mutant run was rejected because the test observation indexed an empty original order list; that was an adapter failure, not a credited semantic catch. The observer now handles valid wrong-shaped results safely. The final four mutants all compile and execute cleanly before comparison fails. No production source is mutated.

## Consumers and readiness

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Each loses one listed dependency; none loses its last blocker. readiness.json subtracts only this uniquely owned helper. The NewTheme and FrameworkStaticReading duplicates were withdrawn to earlier slot 05/03 claims, and do not count as additional delivered helpers.

## Commands and gaps

```bash
source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/lint/helpers/from-wave1-06 -run '^TestRepositoryStatics$' -count=1 -v -timeout=20m > /tmp/helpers06-repository-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/from-wave1-06 > /tmp/helpers06-repository-vet.log 2>&1
python3 stage1/cohere/lint/helpers/from-wave1-06/testdata/regenerate.py > /tmp/helpers06-repository-regenerate.log 2>&1
```

The captured six-consumer cohort is identical to the prior capture at the same cohere pin. Separate capture/rule-package results remain failed on eight external Tailwind/live/corpus placement checks, documented in historical REPORT.md and raw evidence/capture.log. They are not counted as a passing rule gate. Source capture verifies every consumer contributes inputs, not that this helper alone completes findings/fixes parity. Full rule integration and supplying the independently compiled PropertySort remain other units' work.

Toolchain: setup passed in 490s on 5 processors; Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 20s, cache warm 490s. Go 1.27.1, clang 20.1.8, Node 24.19.0. The bounded external Node oracle passed six uncached probes in 10.132s. No shared harness, shared generator, compiler file or cohere worktree source was authored or edited.

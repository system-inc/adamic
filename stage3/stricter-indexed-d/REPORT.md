Built a ledger-linked generated .ts harness; D107 and D171 dense reads are proven.
Base commit: 390af985; witness commits will be recorded below.
Commands: cloud/setup.sh and the first group test, with output in /tmp/stricter-indexed-d-*.log.
Mutants: D107 and D171 erase-panic mutants caught, each cleanly exits 0.
Not covered yet: remaining groups; record chains and holes require observed refusal evidence.

## Scope and assumptions

The exact seven-file filter at ledger commit a1a16427 produces 24 rows in both
rows.csv and rows-2026-10-08.csv, rather than the requested 27. This unit follows
the exact named-file and option filter. The discrepancy has been reported.

The tests generate .ts sources under a project tsconfig with
noUncheckedIndexedAccess=false. They run that actual source with Node's type
stripping, then compare native release, ASan/UBSan and backend JavaScript.
The present control prints 7; the absent read prints undefined on source Node.
Native/backend failure pins stdout, stderr and exit 70 independently, including
an independently computed source line and column. The CLI --explain-checks must
list the inserted site and count one checked indexed guard and zero trusted.
Each supported absent case has its sole emitted-C panic erased, must still build
under sanitizers, and must lose the exact named exit-70 result.

Receiver categories and index expressions are preserved. TypeScript's large AST
interfaces are reduced to the field needed for observation; NodeArray metadata
is omitted while preserving a readonly array receiver. These are minimal read
shape witnesses, not whole transformer executions or whole-program compilation.
Stored new source files are Go and JSON; project .ts inputs are generated tests,
following stage3/stricter-options/indexed_test.go.

The rows in program.ts use options.paths[key][i], a record-to-array chain.
D184, D185 and D186 all originate in superPath[superPathDepth]. D184's TS2538
reports the resulting undefined index; D185 and D186 report arithmetic uses of
that same index. A statement-array-only probe would test a different cause.
D231's tuple position [1] is static; the unchecked read is its nested declarations[0].
No receiver in this exact slice is a Map, typed array or indexed string.

## Progress

Toolchain dependency checkout is in progress. The first attempted test exited 1
before compilation: cohere/TypeScript/tsc/go.mod was not yet available during
setup's recursive submodule checkout. This is not a witness or mutant result.

Group 1 passes: 5 ledger rows assessed, 2 dense read witnesses proven, 3 record
chains blocked, 19 rows remaining. D107 and D171 hole variants are blocked.
Record refusal: `Adamic 0.1 refuses an index signature; use a Map, which keeps
keys in the order they were added`. Hole refusal: `stage 0 can't lower new an
Identifier yet`. Node prints 7 for present controls and undefined for empty/hole
reads. Group command: `go test ./stage3/stricter-indexed-d -count=1 -timeout 10m -v`;
exit 0, package 6.652s, log `/tmp/stricter-indexed-d-group1.log`.

Setup completed successfully in 502.808s: Go 0.044s, Node 0.024s, clang 0.226s,
markdown dependencies 1.785s, submodules 295.588s, Go build 502.683s, cache warm
502.780s. nproc=5, cpu.max=400000 100000. Environment sourced from
`/workspace/adamic-tools/env.sh`. Go 1.27.1, Node 24.19.0, clang 20.1.8.

Group 1 commit: 8f28caf6, pushed to codex/stricter-indexed-d.
Group 2 adds D172, D173, D174, D175 and D184: present and absent proofs pass in
all modes; all five erase-panic mutants build and exit 0, caught by the required
named stop. Their hole variants all produce undefined on Node and the same
constructor refusal natively. Total: 7 dense reads proven, 3 record rows blocked,
14 remaining. Filtered group command uses
`-run 'TestLedgerWitnesses/(D172|D173|D174|D175|D184)$'`; exit 0, package 13.101s,
log `/tmp/stricter-indexed-d-group2.log`.

Group 2 commit: e4ebd182, pushed to codex/stricter-indexed-d.
Halfway report, after group 3: 15 rows assessed, 12 dense reads proven,
3 record rows blocked, 9 rows remaining. New proofs are D185, D186, D189, D191
and D192. All five erase-panic mutants build and exit 0; exact exit/stderr
assertions catch them. Their hole variants remain refused. Command filter:
`-run 'TestLedgerWitnesses/(D185|D186|D189|D191|D192)$'`; exit 0, package 13.623s,
log `/tmp/stricter-indexed-d-group3.log`.

Group 3 commit: ee0a5320, pushed to codex/stricter-indexed-d.
Group 4 proves D193, D194, D225, D226 and D227, with five more caught
exit-0 erase-panic mutants and five observed hole-constructor refusals.
Total: 17 dense reads proven, 3 record rows blocked, 4 remaining.
Command filter: `-run 'TestLedgerWitnesses/(D193|D194|D225|D226|D227)$'`;
exit 0, package 13.280s, log `/tmp/stricter-indexed-d-group4.log`.

Group 4 commit: 3c5c85b4, pushed to codex/stricter-indexed-d.
Group 5 proves D228, D229, D230 and D231. Four more erase-panic mutants build
and exit 0, caught by exact named-stop assertions. All four hole variants refuse.
Total: 21 dense reads proven, 3 record rows blocked, 0 known matching rows
remaining. Command filter: `-run 'TestLedgerWitnesses/(D228|D229|D230|D231)$'`;
exit 0, package 11.659s, log `/tmp/stricter-indexed-d-group5.log`.

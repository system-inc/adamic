Built: project checking roots and explicit runtime entries are separate; project CLI requires --entry.
Implementation: 2ba059f220ae12b4f6d37b7939d66a1662232f71, following cd8592d and 4bbc878 on codex/multi-root.
Checks: load/lower/CLI packages, uncached module oracle and counts, vet, gofmt and diff checks passed.
Mutants: config roots as entries, unchecked unimported files, and outside-config entries were all caught.
Not covered: complete gate, native tsc, project-reference builds, and cycle-worker integration.

# Entry correction

This report supersedes REPORT.md's original behavior of executing configured roots.
The old report describes the original commits; it is not the final project-entry contract.

`adamic build --project tsconfig.json --entry main.a -o program` checks every
configured source under the parsed project options and checks imported dependencies.
It then lowers the explicitly selected entry using the existing module walk.
The entry is relative to the working directory, must be an executable configured
source, and does not have to be the first configured source. Merely being imported
by a configured file does not make a file eligible as an entry.

`load.LoadProject` checks without selecting execution entries.
`load.LoadProjectEntry` checks and selects one entry.
`Program.Files()` retains checking sources; `Program.Entries()` supplies lowering.
Direct `load.Load` and explicit CLI root lists still launch the specified roots
in their supplied order. The three-root fixture remains an explicit-root oracle.
Its project variant now executes only third.a and is compared to Node running
that source alone. Three lowering helpers which used the first checking source
now use the first execution entry instead.

The new fixture lists unimported.a before entry.a. Node running entry.a prints
exactly `entry\n`. The native ASan/UBSan build prints the same, with no output from
the unimported source. The fixture has a recorded allocation-count row of all zeroes.
An outside entry is refused, naming the file and the fix: add it to files/include
or choose an already listed entry. A TS2322 error in an unimported configured file
is reported before lowering. Missing entry selection is refused by the CLI and API.
A relative path containing .. is also tested; counts verification exposed the
initial path-comparison bug, and the final implementation resolves it with filepath.Abs.

# Commands and observations

The original setup remains in REPORT.md: 97 seconds to warm the build cache;
all other setup timing lines were 0 seconds; nproc was 5. Every Go invocation here
sourced /workspace/adamic-tools/env.sh. Tests wrote logs directly.

- `go test ./internal/load ./internal/lower ./cmd/adamic -count=1 -timeout 30m`:
  passed, load 1.696s, lower 16.223s, CLI 1.314s.
  Log: /tmp/adamic-entry-tests-final.log.
- After adding the missing-entry CLI test:
  `go test ./cmd/adamic -count=1 -timeout 30m`: passed, 0.399s.
  Log: /tmp/adamic-entry-cli-final.log.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts`:
  passed, 18.789s. Only the new project-entry row was added; explicit multi-root
  counts remain 4 allocations, 4 frees, 2 retains, 8 releases, peak 1.
  Log: /tmp/adamic-entry-counts-update.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCountsAreRecorded$|^TestNativeAgreesWithNode$/internal/oracle/testdata/modules' -count=1 -timeout 30m`:
  passed, 19.527s. Log: /tmp/adamic-entry-oracle-final.log.
- `go vet ./...`, `gofmt -l cmd internal`, `git diff --check`: passed without findings.
  Logs: /tmp/adamic-entry-vet-final.log, /tmp/adamic-entry-format.log,
  /tmp/adamic-entry-diff.log.
- Full gate was not rerun. The original unit's full gate did not finish; this
  correction uses the permitted scoped gate and records that limitation.

# Mutants

Command:
`python3 stage3/multi-root/mutants.py /tmp/adamic-entry-mutants config-roots-as-entries unchecked-unimported-config-files outside-config-entry`.
All three returned test exit 1, with real test failures and no build failure.
Summary log: /tmp/adamic-entry-mutants-final.log; individual logs in that directory.

| Mutant | Check that caught it | Observed failure |
| --- | --- | --- |
| Lower Program.Files instead of Program.Entries | TestProjectEntryAgreesWithNode | Native printed unimported must not run followed by entry; Node printed only entry. |
| Load only the entry instead of LoadProject | TestProjectChecksUnimportedFiles | Unimported TS2322 error disappeared, yielding nil error. |
| Accept an outside entry and select the first configured source | TestProjectEntryMustBeConfigured | Outside entry unexpectedly returned nil error. |

An initial edit broke the mutant runner's selection comprehension. It was fixed
before the successful runs above; that syntax error is not counted as a killed mutant.
The earlier 16 mutants and their observations remain in REPORT.md.

# TypeScript smoke

Fetched the actual upstream source archive at the census's pinned commit
050880ce59e30b356b686bd3144efe24f875ebc8 into /tmp/adamic-entry-typescript.
Built the CLI with `go build -o /tmp/adamic-entry-compiler ./cmd/adamic`.

Requested entry/config smoke:
`/tmp/adamic-entry-compiler build --project /tmp/adamic-entry-typescript/src/tsc/tsconfig.json --entry /tmp/adamic-entry-typescript/src/tsc/tsc.ts -o /tmp/adamic-entry-tsc`.
Exit 1, exact message:
`adamic: load: project references are not yet supported; build each project explicitly`.
Log: /tmp/adamic-entry-tsc-smoke.log.

Second probe used src/compiler/tsconfig.json with the same tsc entry.
Exit 1, exact message:
`adamic: load: compiler option strictBindCallApply is false; Adamic requires it to be true`.
Log: /tmp/adamic-entry-tsc-compiler-smoke.log.
Both stop during config loading, before checking, entry membership, lowering,
native compilation or execution. No reference removal, option weakening, prelude
removal or source adaptation was used. The compiler project's file list would
also need to explicitly include tsc.ts to make it eligible as an entry.

# Integration boundary

Read pushed codex/import-cycles at 6b1763682d137d3fff82ee3e9321790e4b28b514.
Its moduleOrder adds runtime-only import/export edges, cycle visitation and
load-time-read checks; its lower.go links declarations across modules before
lowering function bodies. Those changes were not merged into this branch.

Both units need lower.go. This correction stays at the entry selection boundary;
it does not edit the cycle worker's module traversal or declaration-linking block.
At merge, retain this branch's Entries filtering/rootOrder call and the cycle
worker's declaration linking and cyclicModules field. For several entries, review
how repeated moduleOrder calls preserve cycle state; that combination has not
been validated here.

The current branch still has the preexisting module walk's cycle refusal and
type-only traversal. Runtime-only edge semantics are owned by the cycle worker
and remain an integration prerequisite; this correction establishes that
unimported config roots cannot become runtime entries. It does not claim the
combined cycle implementation or native tsc is ready.

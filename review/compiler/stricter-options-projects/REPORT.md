Combined stricter option scheduling with transitive project source roots for step 32, #acxgkff and #k881crd.
Loader candidate: compiler/project-references-main e2aa3750c4eae75f673672419ad524f182d6343d; combined delivery SHA is in the handoff.
Focused loader tests, stock tsc comparisons, native/JavaScript fixtures, counts and integration lane checks pass; fixtures print 7.
All 13 loader mutants, six record mutants, three record-read mutants and the caught-host-brand mutant are caught.
Native tsc stops at the mutable MapLike index signature in corePublic.ts:14:5; no whole package or full gate was run.

The combined branch starts at the requested stricter-options tip
f5eca4fb7e16f166374cad8d8d11172ecd7a7fe9 and merges the standalone loader
candidate. The loader candidate replays the five old topic commits on main;
it carries no stricter-options dependency. Current main's test-only changes
are retained in both candidates. Evidence was relocated into review/ following
the layout addendum; the already-published loader candidate was updated for
that relocation. The combined branch is pushed once after these checks.

The combined loader checks configured and explicitly supplied roots together,
including unimported files of transitive references. Execution entries remain
the requested implementations. Per-edge option comparisons still refuse
checking/emit conflicts with both project paths and the option name. Types
lists merge in first-seen order; declaration validation cannot be skipped for
the resulting union. Referenced indexed-access obligations now reach stricter
options' scheduler instead of being refused by the older loader audit.

Merge resolutions

load.go retains stricter-options' scheduled checks, disposition bookkeeping,
nullable handling and standalone Set prelude. The source-root graph, compatible
ownership set, types union, config-directory identity and complete checker root
list are integrated into it. A missing embedded prelude root was caught by
TestProjectReferencesSharedPrelude and restored. project_options.go retains
UTF-16 diagnostic positions while flattening the same reference graph for the
audit. source_fs.go retains JSON and host-console handling. The global console
detector rejects namespace-local declarations. Project Node imports use their
selected declarations; standalone Node loading preserves prelude roots.

The runtime merge retains main's checked-view contracts and semantic tags,
alongside stricter-options' nullable values, caught-value certificates and c2
record support. The duplicate Record enum is consolidated at main's stable tag
position. Object allocation includes both insertion-order storage and readiness
flags. Field initialization and writes retain checked-view representation tags
and use the actual object's slot index. Tuple initialization runs after the
spread object exists. Catch certificates and union-read certificates both
satisfy their corresponding lowerer proof checks. No direct edits were made to
internal/lower/lower.go, internal/native/emit.go, internal/native/native.go or
internal/oracle/oracle_test.go; those files merged automatically where changed.

Main's bounded stage1 test harnesses are retained. Vet exposed duplicate estree
helpers from retaining main's harness alongside stricter-options' additional
parallel harness; the complete current-main estree harness replaces that
combination. Integration's parallelism findings are fixed with Parallel first
for independent tests, and explicit shared-state reasons for process environment,
clock-hook, counts-file and bridge timing tests. The inherited six-mutant record
loop took 107.94s under concurrent checks, so it is now six top-level leaves.
The generated-program witness is split into twelve top-level five-seed leaves.
The exact conflicted file list is in evidence/merge-conflicts.json.

First native tsc stop

The remote adapter product request returned HTTP 403. The repository's explicit
local product build completed instead, using the pinned upstream TypeScript
commit 050880ce59e30b356b686bd3144efe24f875ebc8 and its existing adapters. The
actual CLI then ran the unchanged project graph; no checking overlay was used.
Both native layouts exit 1 with:

```
adamic: /tmp/projects-tsc-adapted/src/compiler/corePublic.ts:14:5: stage 0 can't lower a record other than a readonly string index signature holding finite readonly records, nonnullable scalars or arrays of nonnullable scalars yet
```

The declaration is `export interface MapLike<T> { [index: string]: T; }`.
The builder.ts:1246:69 checking stop is absent from both runs. This observation
establishes progress through checking, not successful native execution of tsc.

Exact preparation and entry commands, with all output redirected:

```
source /workspace/adamic-tools/env.sh
STAGE3_CACHE=/tmp/projects-stage3-cache STAGE3_PRODUCT_STORE=/tmp/projects-stage3-cache/products bash stage3/apply.sh --build-product /tmp/projects-tsc-adapted > /tmp/projects-combined-adapt-build.log 2>&1
npm install --prefix /tmp/projects-tsc-adapted --ignore-scripts --no-save --package-lock=false --no-audit --no-fund @types/node@25.3.3 > /tmp/projects-combined-entry-deps.log 2>&1
go build -o /tmp/projects-combined-adamic ./cmd/adamic > /tmp/projects-combined-build.log 2>&1
ADAMIC_NATIVE_SPLIT=0 /tmp/projects-combined-adamic build /tmp/projects-tsc-adapted/src/tsc/tsc.ts -o /tmp/projects-tsc-0 > /tmp/projects-combined-entry-0.log 2>&1
ADAMIC_NATIVE_SPLIT=1 /tmp/projects-combined-adamic build /tmp/projects-tsc-adapted/src/tsc/tsc.ts -o /tmp/projects-tsc-1 > /tmp/projects-combined-entry-1.log 2>&1
```

The shared toolchain was set up for the standalone candidate with
GOPROXY='https://proxy.golang.org|direct'. Successful setup printed Go 0.022s,
Node 0.025s, submodules 0.064s, markdown 0.080s, clang 0.187s, build 43.780s,
cache 43.913s, done 43.940s; nproc is 5, CPU quota 4. Setup failures and their
repairs are recorded in the standalone candidate's report.

Focused verification commands

```
go test ./internal/load -run 'TestProjectLoader|TestProjectReferences|TestNodeLibrary|TestProduction|TestProjectOption|TestCollectionIterator|TestLibraryIterator' -count=1 -timeout=10m -json > /tmp/projects-combined-loader.jsonl 2> /tmp/projects-combined-loader.stderr
PROJECT_LOADER_TSC=/tmp/projects-stock/node_modules/typescript/lib/tsc.js python3 stage3/project-references-source/verify_types.py > /tmp/projects-combined-types.log 2>&1
PROJECT_LOADER_TSC=/tmp/projects-stock/node_modules/typescript/lib/tsc.js python3 stage3/project-references-source/verify.py > /tmp/projects-combined-references.log 2>&1
go test ./internal/native -run '^Test(RecordsAgainstNode|RecordMutants|RecordReadMutants|CaughtHostErrorsAndBrandMutant|ViewMixedUnionUnknownAndUnavailable)$' -count=1 -timeout=10m -json > /tmp/projects-combined-runtime.jsonl 2> /tmp/projects-combined-runtime.stderr
go test ./internal/lower -run '^Test(MixedUnionContract.*|UntaggedView.*|UnclassifiedCaughtPrototypeMemberIsNotYet)$' -count=1 -timeout=10m -json > /tmp/projects-combined-lower.jsonl 2> /tmp/projects-combined-lower.stderr
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m -args -update-counts > /tmp/projects-combined-counts.log 2>&1
go test ./internal/native -run '^TestRecordMutant' -count=1 -timeout=10m -json > /tmp/projects-combined-record-leaves.jsonl 2> /tmp/projects-combined-record-leaves.stderr
go test ./internal/fuzz -run '^TestGeneratedProgramsCheckAndLower|^TestParallelFeatureCoverage$' -count=1 -timeout=10m -json > /tmp/projects-combined-fuzz-leaves.jsonl 2> /tmp/projects-combined-fuzz-leaves.stderr
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/stage3/project-references-source/' -count=1 -timeout=10m -json > /tmp/projects-combined-reference-oracle.jsonl 2> /tmp/projects-combined-reference-oracle.stderr
go test ./internal/load -run '^TestProjectLoaderNodeImportsKeepReferenceRoots$' -count=1 > /tmp/projects-combined-node-normal.log 2>&1
go test -overlay /tmp/projects-combined-node-overlay.json ./internal/load -run '^TestProjectLoaderNodeImportsKeepReferenceRoots$' -count=1 > /tmp/projects-combined-node-mutant.log 2>&1
go test -overlay /tmp/projects-combined-prelude-overlay.json ./internal/load -run '^TestProjectLoaderStandaloneNodeKeepsPrelude$' -count=1 > /tmp/projects-combined-prelude-mutant.log 2>&1
go test ./internal/load -run 'TestProjectLoader|TestProjectReferences|TestProduction|TestProjectOption' -count=1 -timeout=10m -json > /tmp/projects-combined-final-loader.jsonl 2> /tmp/projects-combined-final-loader.stderr
```

The verifiers use stock TypeScript 6.0.3 on Node. The two-project and three-project
solution builds print 7, matching both Adamic backends. An unimported bad source
produces the same TS2322 code, location and text on both checkers. The clean types
union prints 7 on both backends. The duplicate union produces both stock TS2451
diagnostics, including when projects request skipLibCheck. The original DOM
console profile remains an explicit lowering refusal; the executable shared
string-console profile is the one held to emitted output.

Every loader mutant and its witness

- drop-transitive: transitive implementation root assertion.
- declarations-only: project-reference TS6305 returns.
- allow-conflicts: named conflicting-option refusal assertion.
- drop-reference-audit: missing scheduled indexed obligation assertion.
- duplicate-roots: each implementation must load once in the diamond.
- physical-prelude: shared embedded console identity assertion.
- allow-cycle: circular-reference refusal assertion.
- entry-types-only: clean production union loses a required ambient package.
- audit-entry-types-only: audit union loses a required ambient package.
- skip-union-declarations: production fails to report duplicate TS2451 declarations.
- audit-skip-union-declarations: audit fails to report duplicate TS2451 declarations.
- node fallback on projects: project-owned declarations assertion.
- standalone Node rebuild from execution roots: missing console, node:fs and adamic declarations.

The first Node fallback mutant survived the original root-only assertion because
the stricter loader already preserved the config's root list. The strengthened
witness also rejects standalone Node declarations in a project-owned program;
the mutant now exits 1 on that assertion. The prelude mutant exits 1 on the
missing-declaration assertion. Both compile successfully; evidence contains
all thirteen failing mutant logs. The verifiers' own JSON records their
independent commands and temporary log directories.

Additional runtime mutants are caught by stock Node ordering/iteration comparison
(indices-in-insertion-order, uint32-max-as-index, deleted-key-iterated), sanitizer
checks (overwrite-key-leaked, stored-key-freed, own-slot-null-read), exact missing
member diagnostics and own-hit fixtures (prototype-membership-restored,
missing-read-silent, own-read-checked-as-missing), and the caught-host-value brand
witness. The native JSON records those catches. The runtime and lowerer selections
pass; no full package suite was run.

Test grain and lane result

All final selected loader leaves are below 60 seconds; the final selection passes
in 8.465s. The types-union clean/duplicate/skipped-declaration leaves are separate
roots. Exact leaf seconds are in evidence/final-loader-results.json. Counts refresh
passes in 245.162s and records the inherited combined fixture rows plus the two
reference fixtures. It is the explicitly required existing counts update, not a
new test leaf. Both reference oracle leaves pass in 0.62s and 0.63s.

Record leaves: TestRecordMutants 0.88s, Uint32Max 0.66s, DeletedKey 0.99s,
OverwriteLeak 1.18s, StoredKeyFreed 0.89s, OwnSlotNull 0.99s. These measurements
include each leaf's setup with the runtime build cache populated by the earlier
run. The earlier uncached individual mutant checks took 15-26s, while the former
aggregate exceeded the leaf budget and was split.

Generated-program leaves (suffix, seconds): 001 1.23, 006 2.72, 011 1.64,
016 2.00, 021 1.63, 026 2.09, 031 2.24, 036 1.72, 041 0.90, 046 5.04,
051 0.43, 056 1.81. These check all sixty seeds, including the stricter/c2
refusals, rather than weakening the accepted-program contract.

The mandatory command was run against committed candidates:

```
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

The loader lane passes: gofmt/tools on 10 Go files, parallel declarations on two
test packages, vet two packages. The combined lane passes: gofmt/tools on 289 Go
files, parallel declarations on 31 test packages, vet all 31 packages. The final
lane output is retained in evidence/lane.log. Reports and results are under
review/compiler/project-references-main/ and review/compiler/stricter-options-projects/.

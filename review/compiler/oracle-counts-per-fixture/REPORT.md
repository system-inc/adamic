Split committed counts into one greppable file per fixture without changing a recorded value.
Branch compiler/oracle-counts-per-fixture; base 832ec1dbe5e990ccd77ed4854e684e1c876918e8; delivery SHA is in the final handoff.
The old-to-new proof verifies all 1,128 rows; full oracle, regeneration, vet and guards are recorded below.
Missing-file, orphan, changed-value and CRLF-writer mutants exercise the census, count comparison and byte-regeneration checks.
Not covered: actual execution on a macOS host (this worker is Linux), full repository/Loom gate, or ahra submission.

Layout and assumptions

- internal/oracle/testdata/counts/runtime/<repository-relative-source-path>.counts contains each runtime row, including its fixture key and all six original decimal values. Source extensions are retained, so foo.a and foo.ts remain distinct.
- internal/oracle/testdata/counts/predicates/<repository-relative-source-path>.counts contains each predicate row and its four original values. The former predicate table therefore has no shared changing file either. The current namespaces have no overlapping fixture keys.
- All original header text, column meanings and workflow move into internal/oracle/testdata/counts/README.md, with the house clang 20.1.8 and portable-format contract documented. Each data file has one old-format row and exactly one LF newline. No timestamp, host path, digest, platform name or generated index is committed.
- Input records and explicit frontend adapters use their existing source key in the runtime mirror. The planted row has the stable name runtime/dedication/dedication.a.counts. Predicate records use their existing overload fixture keys; there is no synthetic table row that can become a shared conflict point.
- Conservative assumption: census means the existing countable-fixture registry, not deliberately refused/NotYet programs or the two documented uncounted programs. Its universe is exactly the old table, preserving every old exception rather than fabricating measurements for refused programs.

Independent recorded oracle

recordedCounts reads committed files, never derives expected values from measurements. Ordinary tests never rewrite those files. Actual runtime measurement products remain keyed by their emitted program/runtime/build inputs; the fixed sample and planted-case checks remain unchanged. TestMain publishes explicit -update-counts results only after every selected test succeeds. Selective updates preserve unmeasured files and unchanged content is not rewritten. Removing a counted fixture requires removing its record; the census rejects leftovers rather than silently dropping evidence.

All active oracle readers were migrated, including the supplied readers and five secondary readers discovered by searching countsPath. The historical non-null migration manifest marks a predicate-only fixture as counted; that reader now reads both namespaces separately, preserving its old meaning without mutating the shared runtime map. Dated docs, notes, review and cloud reports were not rewritten. The borrow-chains Python count mutant now edits its one mirrored fixture record.

Old-to-new equality proof

The one-off equality-proof.py reads the pinned base's old counts.md directly from git and compares every complete row and file byte sequence, including every numeric column and key. Output:

```
runtime: 1113/1113 keys and every value identical; row bytes identical
predicates: 15/15 keys and every value identical; row bytes identical
TOTAL: 1128 rows; 0 missing, 0 extra, 0 changed values
```

Determinism

Every runtime case in the existing TestCountsAreRecorded_000 through _031 family regenerates its independently measured row through writeCountFiles into a temporary directory and compares exact bytes with its committed fixture file. This distributes setup and work across existing bounded leaves. TestCountsRegenerateDeterministically gives the planted fixture an explicit named entry point, and TestPredicateCountsRegenerateDeterministically lowers all 15 predicate fixtures and compares their temporary regenerated files. The actual -update-counts invocation is separately checked by SHA-256 of every record before and after, with equality proof repeated afterward.

The common writer validates decimal fields and canonical keys, emits explicit UTF-8 row bytes plus LF, and sorts output paths. It uses no OS-dependent value in file content. Random temporary output roots do not enter records. The same tests compare committed bytes on Linux and macOS; actual macOS execution was unavailable here and is not claimed. The existing measurement/sampling code and pinned stack were retained.

Census and mutants

TestCountsFilesCensus checks every registered runtime/predicate fixture exists as a regular source file, has its expected canonical record, has a unique filename/key, and that no record exists without a counted fixture. README.md is the only non-record allowed in the mirrored tree.

The standalone run-mutants.py runs each mutation in isolation and restores exact original bytes in finally blocks:

- Delete dedication's counts file: census must fail naming that missing record.
- Add ra6ah5s_orphan.a.counts without a counted fixture: census must fail with orphan counts file.
- Increase one fixture's allocation count: the real TestCountsAreRecorded family must fail with the counts changed and that fixture's file name.
- Change only writer output to CRLF via a non-compilable .go.txt Go overlay: predicate regeneration must fail its exact-byte comparison. The committed reader and measurements are untouched, so this is a serializer mismatch rather than a compile/format-validation kill.

Two-commit merge proof

merge-proof.py initializes an isolated scratch Git repository with every migrated record, creates two sibling commits on the same baseline, and moves the allocation count of dedication/dedication.a in one and internal/oracle/testdata/borrow_chain_walk.a in the other. git merge-tree --write-tree --merge-base=<baseline> <left> <right> exits 0. Both edits are checked in the merged tree and its diff has exactly those two files. Full hashes, identical parent hashes and merged tree are in merge-proof.json. The proof commits are local scratch evidence, not extra delivery commits or pushed branches.

Observed development corrections

The first focused build found one obsolete os import after migrating its reader; it was removed. An initial full run exposed the predicate-only historical migration record described above. That reader was corrected, its complete focused test passed, the superseded run was stopped, and the whole oracle package was restarted. Initial logs are retained. No count value or fixture program was changed to obtain a pass.

Toolchain

GOPROXY=https://proxy.golang.org|direct; cloud/setup.sh --wasi-sdk completed in 33.343 s (build cache warm 33.313 s). Linux Go 1.27.1, Node 24.19.0, house clang 20.1.8. nproc=5; cgroup cpu.max=400000 100000, GOMAXPROCS=4 and test -parallel=4. All long commands have hard deadlines and output goes directly to logs. Obsolete export products and the completed previous unit's scratch cache were retired to retain disk headroom; sources and current products were retained.

Final validation

- Entire internal/oracle package: exit 0, 373.461 package seconds, with -count=1 -parallel=4 -timeout 1200s under an outer hard limit.
- Actual full -update-counts run: exit 0, 52.854 package seconds. SHA-256 before/after confirms all 1,128 files byte-identical. The old-to-new equality proof passes again after every mutant is restored.
- TestCountsFilesCensus: pass, 0.01 s in the complete run. TestCountsRegenerateDeterministically: pass, under 0.005 s with its measurement product available. TestPredicateCountsRegenerateDeterministically: pass, 0.56 s. Every runtime family row also exercises temporary regeneration.
- Missing-file and orphan mutants: exit 1, caught by TestCountsFilesCensus. Changed allocation mutant: exit 1, caught by TestCountsAreRecorded_003/dedication/dedication.a, explicitly naming testdata/counts/runtime/dedication/dedication.a.counts. CRLF writer mutant: exit 1, caught by TestPredicateCountsRegenerateDeterministically. All are valid-input/check failures, not compiler-warning kills.
- The migrated borrow-chains Python tool passes as a checker: its retain mutant changes 7 to 8, and the real count assertion reports recorded 8 versus measured 7 with the new file name. Its previous hardcoded 13 baseline was obsolete, so it now increments the actual recorded integer. It restores original bytes; the post-tool equality proof is unchanged. The inner Go package took 17.983 s, under a leaf budget even as a whole selected invocation.
- Final IR call-target reader guard and SortedMap guard: pass, -timeout 90s. Explicit serial export-input preparation precedes parallel readers. Parallel-test guard: pass. go vet ./internal/oracle: pass. gofmt -l on every changed/new Go file: empty output.
- Final audit: 5,282 test identities, zero unresolved failures and zero completed test leaves >=60 s. Maximum completed leaf across the retained runs is 32.21 s, on the four-CPU quota. Every per-test result and second is retained in validation.json.

Exact follow-up commands and wall times:

- `update-counts`: exit 0, 54.426 s; `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -timeout 600s -json -args -update-counts`.

- `mutants`: exit 0, 25.908 s; `python3 /workspace/adamic/review/compiler/oracle-counts-per-fixture/run-mutants.py`.

- `equality-after`: exit 0, 0.064 s; `python3 /workspace/adamic/review/compiler/oracle-counts-per-fixture/equality-proof.py`.

- `ir-guard`: exit 0, 48.157 s; `go test ./internal/ir -run TestCallTargetExportInputs|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted -count=1 -parallel=4 -timeout 90s -json`.

- `parallel-guard`: exit 0, 0.57 s; `go test ./cmd/adamic-gate -run TestEveryTestIsParallelOrSaysWhy -count=1 -timeout 90s -json`.

- `vet`: exit 0, 0.823 s; `go vet ./internal/oracle`.

- `gofmt`: exit 0, 0.008 s; `gofmt -l internal/oracle/cache_test.go internal/oracle/counts_test.go internal/oracle/families_test.go internal/oracle/interface_cast_counts_test.go internal/oracle/node_fs_rm_options_test.go internal/oracle/non_null_checked_test.go internal/oracle/non_null_migration_test.go internal/oracle/predicate_counts_test.go internal/oracle/statement_regions_test.go internal/oracle/count_files_test.go`.

No compiler or runtime production code, fixture source, dated report or other package test was changed. No new compiler-consuming package was added. No PR, force push, main push or Loom submission.

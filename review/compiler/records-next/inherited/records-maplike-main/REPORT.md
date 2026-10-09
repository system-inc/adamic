Built records and named-index support on current main without feature merge history.
Base: origin/main efe9f404; seven selected non-merge compiler commits, listed below.
Validation: 79 uncached oracle fixtures, the unchanged scanner MapLike witness, package checks, counts and vet pass.
Mutants: finite absent entry, named absent entry, member-kind guard, alias readiness and three named source mutations caught.
Not covered: complete native scanner or complete repository gate; no main or area branch modified or pushed.

## Selection and conflict resolution

Created codex/records-maplike-main from efe9f4042049234e5a52639fe77b47c311fd530c.
Main has the records runtime, but none of the compiler records implementation.
Only these non-merge commits were selected from the records worker, with their
necessary compiler dependencies; no runtime branch merge was taken:

| Source | Replayed commit | Purpose |
| --- | --- | --- |
| 11a7fb9f | 560ab17c | Compiler record design |
| 7e2bc734 | 92f7a99a | Pure mutable string dictionaries |
| 70fb62b1 | 37001c8e | Own-guarded inherited reads |
| a2572f0b | 60d7dec5 | Type-only MapLike fixtures |
| 198ff780 | 2636b8e4 | Detached own-property helper |
| 456c981b | 85dc76eb | Finite partial records sharing MapLike storage |
| c20e5b56 | 12d29536 | Named properties beside a string signature |

The named worker tip is c20e5b56, its single implementation commit. The broad
range query also listed unrelated main/area ancestry; none of it was selected.
There are no merge commits after the new branch's main base.

Conflict hunks were handled individually. The expression dispatcher retains
main's enum handling before records; enum namespace objects bypass the record
index-signature gate. Refusals retain main's predicate, enum and checked-cast
proofs plus record storage checks. The updated checker file-path type is
converted to string at the two regexp-origin checks. Runtime documentation and
compiler documentation remain together in docs/records.md.

Object-call argument dispatch uses the actual argument list for record and fixed
object inspection. Main lacks the named worker's general library-method adapter
file. The modify/delete conflict retains only the Object keys/values/entries
and detached own-property alias dependencies needed by this unit; main's other
library families keep their existing paths. Inspection aliases preserve binding
readiness, receiver/key evaluation order, dense apply arguments and descriptor
checks. The main non-object own-property implementation gets an explicit-argument
wrapper without changing its ordinary call behavior. No unrelated library
features or ancestor commits were imported.

Counts conflicts retained main's rows during replay. The complete generator
then removed incoming orphan rows for library-method/namespace fixtures absent
from this tree and duplicate rows from overlapping replay context. The final
521 rows contain all 488 main rows unchanged and 33 new records fixture rows.
Every added row is in evidence/counts-added.json; no main counts row was removed.

## Validation observed on this tree

All test output went directly to log files, now compressed under evidence.

```sh
ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v -run 'TestPartialRecord|TestNamedRecord|TestNativeAgreesWithNode/internal/oracle/testdata/(records_|detached_own_|optional_|object_|enum)'
go test ./internal/lower ./internal/ir ./internal/flow ./internal/fresh ./internal/javascript -count=1 -timeout 30m
go test ./internal/oracle -count=1 -timeout 30m -run '^TestCountsAreRecorded$' -args -update-counts
NAMED_INDEX_MUTANT_LOGS=/tmp/maplike-main-static-mutants python3 stage3/named-index-records/run-mutants.py
go test ./internal/native -count=1 -timeout 10m -v -run '^TestRecordsAgainstNode$|^TestRecordReadMutants$'
go vet ./internal/lower ./internal/ir ./internal/javascript ./internal/oracle
gofmt -l internal/lower internal/ir internal/flow internal/fresh internal/javascript internal/oracle
git diff --check
```

The 79 fixtures are the requested original 73 plus all six named-index fixtures.
The oracle passed in 51.416s, native hits 0/misses 209 and Node hits 0/misses 181.
Normal completion agrees with source Node in both backends, sanitized and release
native builds, with leak checks. Deliberate checked-stop fixtures retain their
existing pins; records_named_invalidated stops 70 where Node prints changed1.
Lower passed in 87.438s, IR 27.565s, flow 132.531s, fresh 95.676s; JavaScript has
no package tests. Full counts passed in 53.415s. The targeted runtime comparison
and three runtime read mutants passed in 15.605s. Vet and formatting logs are empty.

Setup initially encountered the pending cherry-pick conflict markers. The final
setup succeeds: Go 0.023s, Node 0.024s, markdown 0.083s, submodules 0.094s, clang
0.248s, build 86.041s, cache warm 87.517s, total 87.644s; nproc 5, four CPU quota.
Go 1.27.1, Node v24.19.0, clang 20.1.8, Linux amd64.

## Scanner MapLike replay

corePublic-MapLike.a is unchanged from 503d2e0d's
stage3/drivers/scanner/evidence/native3-records/witnesses/01-index.a.
It declares exported MapLike<T> with one string index signature and prints ok.
Source Node, native and emitted JavaScript each exit 0, print exactly ok plus a
newline, and have empty stderr. Both byte comparisons are empty.

The source and emitted JavaScript run through oracle/node.mjs, which strips
Adamic types and supplies its runtime import. The CLI command is
adamic build corePublic-MapLike.a -o program, with the source before -o.
Initial direct Node .a loading and incorrect CLI argument order were corrected;
all checked-in replay evidence is the completed successful rerun.
This proves the reported corePublic index-signature stop is closed; it does not
claim the full scanner now builds or has byte-identical native token output.

## Mutants observed

- Finite absent-entry mutant: a missing declared key becomes an own undefined
  entry. Native and JavaScript print 1 where the parser fixture on Node prints 0.
- Named absent-entry mutant: an own undefined target is synthesized. Both
  backends finish but disagree with Node stdout.
- Named member-kind guard removed: JavaScript runs on exactly as Node, violating
  the fixture's required exit-70 check.
- Alias-copy readiness removed: both backends exit 0 where Node stops 70.
- Three real source mutants: bypass named read type, unrestricted-write proof,
  and named alias contracts. Their respective lower tests fail with the pinned
  named read or NotYet diagnostic, not a C compilation failure. Production
  source is restored after each.
- Runtime mutations restoring prototype membership, silently missing a read,
  and checking an own hit as missing are caught by exact stop/own-hit fixtures.

The full repository gate, full scanner build/token comparison, macOS and other
targets were not run. The unit pushes only codex/records-maplike-main, once.

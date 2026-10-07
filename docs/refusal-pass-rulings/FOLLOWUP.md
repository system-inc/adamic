Built fixtures for the reported Function effects and clear compile-time diagnostics for incomplete record operations; 36 fixtures in total.
Code commit 40cd6d1; merged current origin/main c7991b9 in 3e30b3f before the final gate.
Commands: affected packages, filtered uncached oracle, vet and formatting passed; all nine observed programs stop during compilation on the fixed branch.
Mutants: the eight original checks plus record reads, writes, values and entries all failed TestRefusalPassRulings after the main merge.
Not covered: full repository gate, optional widening, implementing record lowering, or an exhaustive record-operation census.

## Fixtures and effects

Read origin/devtools/refusal-lies at `4fb0b5e`, including its report and
`internal/refusalprobe/lies/`. The exact published `function_type_length.a` is
preserved as `internal/lower/testdata/refusal_pass/function-type-length.a`.
The report describes the other relevant cases but does not ship their sources
in lies/. The closure-passing case is held by this small reproduction:

```a
function take(value: Function): void {}
take((): number => 1);
```

The prior `observe-class.a` fixture reads the merged number member and prints
its typeof. The separate `observe-record.a` fixture reads the missing named
Record member and prints its typeof. Both have complete pinned refusal messages.
No optional-widening program was changed by this unit.

Reproduced on the unchanged main binary built from `39638d9`, using Node 24.19.0
and clang 20.1.8 on Linux:

| Program | Node source | Main JavaScript backend | Main native backend |
|---|---|---|---|
| Function length, exact published source | `2`, `2`, `2`, `1`; exit 0 | `undefined` four times; exit 0 | clang rejects incompatible closure/object pointers |
| Closure passed as Function | empty stdout; exit 0 | empty stdout; exit 0 | clang rejects incompatible closure/object pointers |
| Merged class/interface member | `undefined`; exit 0 | `undefined`; exit 0 | missing-field compiler-bug panic; exit 70 |
| Named missing Record member | `undefined`; exit 0 | `undefined`; exit 0 | missing-field compiler-bug panic; exit 70 |
| Record element/dynamic read, dynamic write, values, entries | source exits 0 | NotYet during compilation | NotYet during compilation |

Each displayed stdout line ends in a newline. The Function length result is a
**silent wrong output in the JavaScript backend**, independently reproduced here.
The original report's Function call observation was blocked at lowering; it did
not test this separate length observation. Both length and closure-passing
programs now fail at the Function annotation, before either backend emits code.
These compiler diagnostic tests are not claimed to be runtime oracle fixtures.

## Record decision

The named missing-key read is **Refused**, at its Record annotation, with the
existing temporary message explaining absent own-key lowering and naming Map
and #p9v82wa. It cannot reach the missing-field compiler-bug stop.

Record element reads (including literal-key element reads), dynamic writes,
Object.values and Object.entries remain **NotYet**, with separate messages:

- Dynamic reads: own-key lookup returning `T | undefined` is not implemented; use Map.get.
- Dynamic writes: own-key storage is not implemented; use Map.set.
- Object.values: own-key enumeration is not implemented; use Map.values.
- Object.entries: own-key enumeration is not implemented; use Map.entries.

These messages are returned before lowering creates fixed-shape IR. The normal
visitor still checks permanent refusals first. It defers the temporary Record
annotation refusal until it can name an actual unimplemented operation, so that
an annotation does not change these operations from NotYet to Refused.
`record-operation-permanent-priority.a` proves a later Function annotation still
wins over a record operation. Empty Record declarations remain temporarily refused.
Finite literal-key Records and ordinary fixed objects retain their existing behavior.

During landing, origin/main advanced to `c7991b9`. It includes the standalone
record runtime and its tests, but no end-to-end lowering: docs/records.md explicitly
states that consumers and fixed-object interoperation are not wired into the compiler
and that end-to-end lowering remains compiler work. The branch merged that main
in `3e30b3f` and re-greened all affected tests. The compile-time policy above still
applies. This unit does not claim to implement #p9v82wa or change its missing
prototype-name behavior.

## Verification

All test output was redirected to logs, never piped. The final commands on the
merged branch were:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 python3 /tmp/refusal-followup-mutants.py > /tmp/refusal-followup-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/lower ./internal/load > /tmp/refusal-followup-packages.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^load$/^testdata$/^0.1$/^compile$/^01_hello[.]ts$' > /tmp/refusal-followup-oracle.log 2>&1
go vet ./... > /tmp/refusal-followup-vet.log 2>&1
gofmt -l cmd internal > /tmp/refusal-followup-format.log
git diff --check
```

Final package output: lower passed in 13.796s, load in 1.075s. The oracle passed
in 11.798s and executed 01_hello.ts, with native misses 3 and Node misses 2;
all cache hits were zero. Vet exited 0; formatting and diff checks printed nothing.
The repository-wide gate was not run. No new runtime oracle fixture was added,
so no counts-table row was needed. The earlier setup remains the same unit's
setup: 120s total, nproc 5, with timing lines in setup.log in the parent directory.

Before the main merge the nine observation programs were run with the unchanged
main binary, Node source runner, JavaScript backend runner, and sanitized native
build when compilation succeeded. Each was also compiled through both fixed
backend commands; every fixed attempt exited 1 with the pinned compile-time
category and message, and no fixed runtime execution was possible. Exact commands,
sources, stdout, stderr and exit codes are in [observations.json](followup/observations.json).
Those pre-merge fixed diagnostics were then held by all 36 fixtures on the merged
branch in the final package gate. Main's baseline results in that JSON are
observations of `39638d9`, not claimed executions of `c7991b9`.

## Mutants after merging main

All twelve mutants began independently from the merged source and restored it
in a finally block. Each invocation ran only TestRefusalPassRulings and wrote its
own log. Each exited 1 through its intended diagnostic assertion. No mutant was
killed by compilation, clang, or a toolchain error.

| Mutant | Fixture failure |
|---|---|
| Drop class merge check | observe-class.a and class-interface.a accepted |
| Drop Function annotation check | function-type-length.a and function-passed-closure.a accepted; annotation fixtures also failed |
| Put lowering before refusal | permanent construct fixtures returned NotYet; complete-message pins also failed |
| Drop Record declaration check | observe-record.a accepted, leaving the named-read runtime defect unguarded |
| Drop any check | any.a returned NotYet |
| Drop eval check | eval.a returned NotYet |
| Drop expando check | expando.a and expando-index.a returned NotYet |
| Drop new Function check | new-function.a returned NotYet |
| Drop record read diagnostic | record-dynamic-read.a and record-element-read.a returned Refused rather than NotYet |
| Drop record write diagnostic | record-dynamic-write.a returned Refused rather than NotYet |
| Drop record values diagnostic | record-values.a returned Refused rather than NotYet |
| Drop record entries diagnostic | record-entries.a returned Refused rather than NotYet |

See [mutants.log](followup/mutants.log), [mutants.py](followup/mutants.py), and the
individual mutant logs in followup/. The category pins are essential for these
last four checks: returning a generic refusal would still prevent execution,
but would break the explicit NotYet promise requested for these operations.

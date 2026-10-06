# Nullable reference evidence

Built one nullable pointer path with an instantiated static empty case.
Core commit: 1a9c637cf3556f14462dcf8798cb41ff30bfbb4c; fixtures: b8727ba.
Validation: touched packages passed; the full gate was stopped after 19 minutes.
All seven requested mutants and three additional mutants failed semantically.
Unnarrowed non-string coercion and structural object/class JSON remain NotYet.

## Environment and base

Branch `codex/nullable-references`, based on
`origin/codex/regex-matcher` at `50a1dc8217f6d715fc27038e6ce12ab429416a30`.
The regex branch moved from df959ee during this work; the core was rebased before
its first push. The library-date diff was read, not merged.

`bash cloud/setup.sh` finished in 147 seconds. Its timing lines reported Go, clang,
Node and submodules at 1 second, build cache warm at 147 seconds, and setup complete
at 147 seconds. `nproc` returned 5. The toolchain was Go 1.27.1, clang 20.1.8 and
Node 24.19.0. Commands source `/workspace/adamic-tools/env.sh`.

## Validation

Output is written to log files before being read. The complete worker gate uses
the permitted oracle cache; the new fixtures also run uncached. The uncached oracle
compares source Node, emitted JavaScript, native ASan/UBSan, native release and
LeakSanitizer, including stdout, stderr and exit code.

```sh
gofmt -l cmd internal > /tmp/nullable-gofmt.log
go vet ./... > /tmp/nullable-vet.log 2>&1
go test -count=1 -timeout=30m ./... > /tmp/nullable-final-gate.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=30m ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/nullable_references' > /tmp/nullable-final-uncached.log 2>&1
go test -count=1 -timeout=30m ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts > /tmp/nullable-final-counts.log 2>&1
python3 cloud/reports/nullable-references/run-mutants.py > /tmp/nullable-mutants-final.log 2>&1
```

`run-mutants.py` checks for the intended semantic failure and rejects build-only
failures. See `mutants.txt` and the individual mutant logs. The generic-key mutant
merges nullable string instantiations deliberately; the null/undefined comparison
mutants independently remove the static-type distinction. The tag mutant disables
the tag refusals, and the named test detects accepted programs.

Only four new fixture rows were added to `internal/oracle/counts.md`; no existing
row changed and no old fixture changed lowering status.

## Coverage limits

See `docs/nullable-references.md`. All reference kinds use the same representation,
and proven empty values of arrays, objects, classes, maps, sets and functions have
Node comparisons for spelling, Number, typeof and JSON. Present non-string values
retain the existing NotYet coercion rules. Unnarrowed object/class JSON needs a
complete-shape and toJSON design. This is an incomplete part of the original brief,
not a claim that all requested present-value operations are implemented.
Console array/object formatting was explicitly excluded by the user. Non-null
assertions, loose equality and non-boolean conditions retain existing refusals.

Observed: gofmt and vet produced no output and exited zero. The uncached
nullable oracle passed in 7.720 seconds. The complete counts update passed in
83.591 seconds and changed only the four new rows.


The full worker command was stopped deliberately with SIGTERM after 19 minutes
(exit 143), using the brief's allowance for a slow full gate. This is not a green
complete gate. `worker-gate.log` records completed packages: lowering passed in
45.969 seconds, native in 654.029 seconds, the complete oracle in 299.294 seconds,
and loading in 5.456 seconds. IR and JavaScript built and have no package tests.
Flow, freshness, fuzz, regexp and the test262 command package also passed.
The still-active checks were internal Unicode properties and stage-1 cohere JSON
and parser tests. Their process group was stopped, including children. No failure
had been reported, but their final results are not known. Integration must run the
complete uncached gate before main moves.


## Added match-result typeof fixture

`nullable_references_typeof_match.a` contains the requested `.match()` and `.exec()`
probe unchanged. Node exited zero and printed:

```text
object true
object
object
```

The uncached differential fixture passed in 1.557 seconds, including both backends,
ASan/UBSan, the release build and LeakSanitizer. The targeted typeof-null mutant
replaced the nullable typeof helper's `object` constant with `undefined`. This
fixture alone caught it: both backends printed `undefined true`, `object`,
`undefined` and exited zero, whereas Node printed the output above. Source was
restored in a finally block. A build failure was not counted.

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=30m ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/nullable_references_typeof_match' > /tmp/nullable-typeof-oracle.log 2>&1
go test -count=1 -timeout=30m ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts > /tmp/nullable-typeof-counts.log 2>&1
```

This follow-up adds a fifth fixture row. Existing rows remain unchanged.


## Inline triple-empty reads

The branch's existing expression guard rejects the integration probe. Added
`TestNullableReferenceReadsNeedATagInEveryExpression` checks its complete body and
independent strict comparisons, template spelling, typeof, argument passing and
attempts to guard a repeated nullable Map get with has() or !== undefined. It
requires NotYet with the named empty-case-tag reason. The full lowering package
passed in 14.838 seconds. Disabling the own-type guard in nullableUse and the
matching typeOf guard makes this named test fail with `got <nil>` for forbidden
reads. Compiler sources were restored; no build failure counted.

The new sound-neighbor fixture passed uncached in 1.121 seconds, including both
backends, ASan/UBSan, release and leak checks. It covers a Map<K, R> get narrowed
with !== undefined and Map<K, R | null> value iteration. The checker does not
narrow Map get through has() or across repeated calls. An unnarrowed nullable get
remains NotYet, including when it is the operand of an undefined guard. This
unit does not invent a presence refinement that the checker has not proved.

```sh
go test -count=1 -timeout=30m ./internal/lower > /tmp/nullable-inline-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -v ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/nullable_references_map_narrowed' > /tmp/nullable-inline-neighbor.log 2>&1
go test -count=1 -timeout=30m ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts > /tmp/nullable-inline-counts.log 2>&1
```

The counts update passed in 24.826 seconds. Only the sixth new fixture row was
added; all previously recorded rows are unchanged.


The same named regression was copied into a detached worktree of main at
`50045bd797650a34aa40b55ad751b6a667a6ab31`, with its identical cohere gitlink
`715ba94f3608a6500086b1076ce5cb7e51b836db` reused from this workspace. The focused
command `go test -count=1 ./internal/lower -run
'^TestNullableReferenceReadsNeedATagInEveryExpression$'` failed semantically:
main lowered the full integration probe and individual comparison/typeof cases
successfully (`got <nil>`). See `inline-main.log`. This is an expected test failure,
not a build failure. The existing branch expression guard already supplies the
fix; this follow-up adds regression coverage rather than another special case.

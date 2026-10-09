Scratch-merge measurement stopped at compiler verification; no scanner comparison was run.
Resolved six conflicted files from records-lowering 456c981b onto scratch b05a9306, containing main efe9f404.
The Go compiler built successfully in 10.315s; all three targeted test commands exited 1.
Verification exposed nullable-record, Buffer-view and runtime JSON schema incompatibilities.
Existing mutants were blocked by runtime compilation; no mutant catch or native scanner success is claimed.

# Scratch-merge measurement

This is an attempted, uncommitted integration in the never-pushed scratch branch
scratch/scanner-native3 at /workspace/scanner-native3-scratch. It is not evidence
about a landed compiler. Current origin/main was fetched and remains
 efe9f4042049234e5a52639fe77b47c311fd530c. The scratch base is
b05a9306dd9c49ea1697f9aea4e87fdc3a9d2795, including that main and the library
fixes. Incoming records-lowering is pinned to
456c981b1c108abae5f2c8d6ae665cfd6b92e2fe. No named-index or front-3 merge was
added. No conflict markers remain, and the six resolutions are staged only in
scratch. The scratch merge is not committed, published or recommended for use.

## Each conflict and resolution

The main-side behavior below includes the previously merged library work.
All existing handlers were retained; records additions were placed beside them.
The exact eight conflict hunks and replacement text are in conflict-hunks.json.
The complete integrated diff is in scratch-diff.json, encoded to preserve
whitespace without turning upstream source into a delivery-branch code edit.

| File | Conflicts | Attempted resolution |
| --- | --- | --- |
| internal/javascript/javascript.go | 1 | Retained NodeBufferCall dispatch and added RecordCoalesce, RecordCall and RecordLiteral cases. |
| internal/lower/expression.go | 2 | Retained unknown/object, intersection, Buffer and parameter representation checks before records recognition; retained nodeBufferView rejection before record-element sameKeeping recursion. |
| internal/lower/object.go | 1 | Retained Buffer, filesystem and process builtins, then added detachedOwnCall dispatch before user methods. |
| internal/lower/refusals.go | 2 | Retained Error stack, Node library and null/undefined refusals; added records/detached-own checks and records index/delete rules. Preserved the process.env delete exemption in the new delete check. Retained Buffer unsupported-use checks and all existing method-observation exemptions, adding detachedOwnMethod. |
| internal/lower/statements.go | 1 | Retained Error and process.env expression-statement handlers, then added recordExpression handling. |
| internal/native/emit_expressions.go | 1 | Retained main's StringFromCodes result and checkThrown path, discarding incoming's older return sequence; added RecordCoalesce, RecordCall and RecordLiteral cases. |

Other records changes merged automatically. In particular,
internal/native/runtime/json_stringify.c merged without textual conflict but
failed compilation during the native oracle. The existing scratch cohere symlink
was retained; both merge parents pin submodule 7945d102a6c18dd36adf9114a758ce646e8b2359.
Diff capture excludes that submodule to avoid Git's symlink inspection error.

## Verification and stopping point

Go build exited 0, wall 10.315s, user 22.100s, system 2.171s. Binary SHA256:
fd21c9503505df29b3d57dbf4aa2bbadc3f163cc072d31c6b9cda90041d4c38b.
Gofmt was applied to the six scratch files; scratch diff whitespace check passed.
This proves the Go executable builds, not that its emitted native programs work.

Three targeted test commands completed with exit 1. Readable logs remove trailing whitespace only; raw-test-logs.json preserves
exact oracle and native-fixture output:

- lower-tests.log: TestRecordRefusals accepted a function parameter of
  Record<string,number>|null with JSON.stringify, where the records test expects
  a NotYet error containing "value of type". TestNodeBufferRefusals/typed_array_view
  expected "another object type" but received "a value of type Uint8Array<ArrayBufferLike>".
  Other selected lower tests passed, including records forms, partial storage,
  observation boundaries, prototype names, detached-own and Node library tests.
- oracle-tests.log and native-fixtures.log: runtime json_stringify.c:194 calls
  scalar(*slot, schema->element->kind), but scalar at line 103 takes a
  const adamic_json_schema pointer. Clang rejects the enum-to-pointer argument.
  This runtime failure blocks even the JavaScript comparison helpers, whose
  harness builds runtime support, and the selected native records/Error fixtures.

The unresolved semantic judgment is in internal/lower/expression.go, interacting
with internal/lower/records.go, internal/lower/library_node_buffer.go and
internal/lower/library_json_stringify.go: how records representation should compose
with main's nullable/dynamic-object and Buffer-view handling while retaining
sound admission and diagnostics. The runtime mismatch is in
internal/native/runtime/json_stringify.c; passing the schema pointer appears a
mechanical candidate, but it was not applied after the semantic verification
failures. No tests were weakened, and no unsupported case was silently admitted
as an approved result. Test failures are observations; their precise repair is
not established by this measurement. The compiler worker should resolve these
boundaries before using this merge for scanner measurements.

## Commands

Commands ran in /workspace/scanner-native3-scratch, with logs redirected to files.
The existing session toolchain setup was reused: nproc 5, CPU quota 4, setup total
286.021s (previous evidence ../native3/setup.log). No whole package or full gate
was run. No fixture was added to the delivery branch; counts.md there is unchanged.

```sh
git fetch origin main > /tmp/scanner-records-resolution-fetch.log 2>&1
git -c submodule.recurse=false merge --no-edit origin/codex/records-lowering > /tmp/scanner-records-resolution-merge.log 2>&1
python3 /tmp/scanner-records-resolve.py
source /workspace/adamic-tools/env.sh
gofmt -w internal/javascript/javascript.go internal/lower/expression.go internal/lower/object.go internal/lower/refusals.go internal/lower/statements.go internal/native/emit_expressions.go
TIMEFORMAT='wall=%3R user=%3U sys=%3S'
{ time go build -buildvcs=false -o /workspace/scratch/scanner-records-resolved-adamic ./cmd/adamic; } > /tmp/scanner-records-resolution-build.log 2>&1
go test -v -count=1 -timeout 30m ./internal/lower -run 'Test(Record|PartialRecord|DetachedOwn|NodeLibrary|NodeBufferRefusals|NodeFSFileKeepsDetachedMethodRefusal)' > /tmp/scanner-records-resolution-lower-tests.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run 'Test(Record|PartialRecord|DetachedOwn|ErrorCaptureMutants|NativeAgreesWithNode/internal/oracle/testdata/(records_|detached_own|library_error_capture))' > /tmp/scanner-records-resolution-oracle-tests.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/internal$/oracle$/testdata$/(records_|detached_own|library_error_capture)' > /tmp/scanner-records-resolution-fixtures.log 2>&1
```

The first oracle command selected the named tests and mutants; the separate
anchored command explicitly selected the native fixture paths. Existing mutant
attempts were records read/write/delete/in/has-own/keys/values/entries/spread/for-in/
stringify, prototype, ownership, coalesce, narrowing, observation guards,
partial absent-entry, detached inherited/readiness, and Error capture/stack-limit.
All were blocked by runtime compilation, so none is counted as caught. No new
scanner output mutant was run. Scanner split 0/1, native byte comparison,
scanner timings and a renewed stop walk were not run because the user's compiler
verification prerequisite failed. Evidence audit verifies the retained failures
and successful Go build; it does not certify compiler correctness.

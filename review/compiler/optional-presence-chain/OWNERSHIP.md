Built: documented given store values and borrowed checked-read strings; admitted undefined writes to optional string storage.
Commit: one runtime review amendment on compiler/optional-presence-chain.
Validation: Node, release/native ASan/UBSan/LSan and JavaScript fixtures pass; CallTargetReaders and targeted vet pass.
Mutants: missing retained self-assignment read causes ASan heap-use-after-free; restoring undefined rejection causes exit 70 in both backends, caught by Node.
Remaining: shared gate and runtime review of the additional undefined acceptance hunk; no local blocker.

The declaration says “value is given: owned by the call”. The store assigns without retaining after releasing the old slot. emit_statements.go:215 creates keptValue, and line 217 retains or transfers its ownership before line 238 passes it to the store. The optional string read returns a borrowed string, matching adamic_object_view; optionalViewField retains it before returning an owned IR expression.

Optional string storage is 3, not 10. internal/lower/expression.go:157 skips undefined members while computing the shared reference representation, then returns String. NULL carries the undefined value; presence remains independent. internal/native/emit_objects.go:477 tags an explicit reference-typed Undefined as 13. The checked store now accepts 13 against 3 without changing the physical tag and publishes the property. Exact optional property types reject a direct undefined assignment to s?: string at the checker; the admitted witness spells s?: string | undefined, which still has representation 3. Node and both backends print:

```
undefined
true
kind,s
againagain
```

The self-write witness allocates its string with repeat, keeps no extra string local, narrows its optional read and assigns it back. An instrumented checked read asserts exactly one reference before the caller retains it, including the read used for the self-write. The precise mutant removes only that read's retain, leaving the store intact. ASan reports heap-use-after-free on the subsequent checked read. The undefined mutant restores the previous fits condition in production object.c and generated JavaScript; both stop with `field write failed: s expected string, found undefined` and fail the Node comparison.

Commands and results (logs in this directory):

- `ADAMIC_GATE_UNCACHED=1 timeout 120 go test ./internal/oracle -run 'TestOptionalFieldCheckedView|TestNativeAgreesWithNode/internal/oracle/testdata/optional_field_checked_view' -count=1 -v -timeout 90s`: green, 5.933 s. New leaves self 2.42 s and undefined mutant 1.78 s; fixture leaves 0.99 s and 0.58 s. ownership-focus.log.
- After tightening the ownership assertion to heap-use-after-free, `ADAMIC_GATE_UNCACHED=1 timeout 120 go test ./internal/oracle -run '^TestOptionalFieldCheckedViewStringSelf$' -count=1 -v -timeout 90s`: green, leaf 0.78 s. ownership-self.log includes the sanitizer report.
- `ADAMIC_GATE_UNCACHED=1 timeout 120 go test ./internal/oracle -run '^TestCountsAreRecorded$/^fixtures$/internal/oracle/testdata/optional_field_checked_view_string' -count=1 -timeout 90s -args -update-counts`: green, 29.632 s. Unioned two new rows into the complete table. No measured existing row moved; no existing peak row moved. ownership-counts.log and ownership-counts-diff.md.
- `timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s`: green, leaf 20.54 s. ownership-readers.log.
- `timeout 120 go vet ./internal/oracle ./internal/javascript`: green. ownership-vet.log.
- Required integration lane checks: 12.1 s; gofmt and tools on 272 Go files, t.Parallel on 20 test packages. Lane vet skipped after exceeding 10 s; separate touched-package vet passed. ownership-lane.log. Lane checks are repeated after the commit before pushing.

Setup used GOPROXY=https://proxy.golang.org|direct: go 0.023 s, Node 0.025 s, submodules 0.068 s, markdown 0.080 s, clang 0.173 s, build 43.466 s, deferred tests 43.598 s, cache 43.599 s, total 43.625 s; nproc 5, CPU quota 4. ownership-setup.log.

Exact runtime hunks are in ownership-runtime.patch and explained in runtime-diff.md. The frozen-object objection was withdrawn; the existing SetProperty data-write guard remains. No full packages or full gate were run. No PR was opened.

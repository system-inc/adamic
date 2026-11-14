# process.recordreplay's absent-member road: mutants

process.recordreplay reads as undefined on every target, a member Node's process doesn't have (#b7phwab, Adamic on
#b5w3ycg, Nov 7 2026), through the road browser already took (host_binding_absent_members.go, rule 70, and the
fold in library_node_process.go). Each mutant is one edit, applied alone, then reverted. Run
`go test ./internal/lower -run 'TestProcessRecordReplay|TestHostBindingProcess' -count=1 -v` and
`go test ./internal/oracle -run 'TestProcessRecordReplayAgreesWithNode' -count=1 -v`: the named tests go red. None of
these has been run yet: they were written on a machine that doesn't build Go, so each "goes red" below is what the
code says, not what a run printed.

## a. The fold answers true

File: internal/lower/library_node_process.go, nodeProcessValue's absent-member fold.

Old:

```go
		return fit(ir.Undefined{Of: ir.Object}, of), true, nil
```

New:

```go
		return fit(ir.BooleanConstant{Value: true}, of), true, nil
```

Expected: TestProcessRecordReplayAgreesWithNode goes red, main.a and main.ts alike: native, sanitized and javascript
print `true` where Node prints `false`, and typeof no longer reads `undefined`. The read rows of
TestProcessRecordReplayReadsUndefined that don't go through a bare typeof lose their IR undefined, and browser's
TestHostBindingProcessReads goes red with them, since the fold is shared.

## b. recordreplay dropped from the absent members

File: internal/lower/host_bindings.go, hostBindingAbsentMembers.

Old:

```go
	"node:process.process": {"browser": true, "recordreplay": true},
```

New:

```go
	"node:process.process": {"browser": true},
```

Expected: TestProcessRecordReplayReadsUndefined goes red on every row: the fold no longer matches, and the read
stops at process.go's not yet, "process.recordreplay as a value (use its supported process operation directly)".
The run-time key message loses its "or process.recordreplay", so the `const k` rows of
TestProcessRecordReplayRoadsStop and TestHostBindingProcessBrowserRoadsStop go red too.

## c. The view check names only browser again

File: internal/lower/host_binding_absent_members.go, hostBindingAbsentMemberRefusal's view check.

Old:

```go
			if l.hostBindingProcessAbsentViewed(view) {
```

New:

```go
			if l.checker.GetPropertyOfType(l.checker.GetNonNullableType(view), "browser") != nil {
```

Expected: Language's first attack row of TestProcessRecordReplayRoadsStop,
`const p: { recordreplay?: unknown } = process; p.recordreplay = 1;`, goes red. By the code's reading, rule 70 no
longer judges the view, and the declaration's initializer stops at the generic stop, "process as a value (use its
supported process operation directly)", which doesn't name recordreplay, so the row fails its "want a stop naming
recordreplay". If the run shows it lowering instead, it's red the same way. browser's own view rows in
TestHostBindingProcessBrowserRoadsStop stay green, which is why the recordreplay row is the one that holds this check.

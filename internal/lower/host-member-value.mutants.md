# Mutants for a host member read as a value off a program object literal (#fj0vqtf, wave 2)

Each is one edit to internal/lower/host_member_value.go, run, then put back, with
`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -run '^TestHostMemberValueAgreesOrStops$' -v`.

## (b) The closed list judging only the binding's calls, .call and writes

In hostMemberValueUse, old:

```go
	return l.notYet(use, hostMemberValueNamed(judge)+" "+hostMemberValueOther(held))
```

New:

```go
	return nil
```

Red: r1_holder_run, r2_alias_field and r3_apply, .a and .ts, each `lowered, want ...`: the value stored in
another object's field or passed as an argument is then let through with nothing left to stop it. h2 stays
at its plain call, which hostMemberValueCall judges before the list's last line.

## (c) The reach rule dropped on the value read

In checkHostMemberValues, old:

```go
		if err := l.hostMemberReach(value.stop, reaches, index); err != nil {
```

New:

```go
		if err := l.hostMemberReach(value.stop, reaches, index); false && err != nil {
```

Red: s3_unknown_reach_plain_call, .a and .ts: a read whose reach is unknown goes on to its uses, so its
plain call is judged in place of main's refusal. Head 1 measured it as the mutant's only witness, when
every other unknown-reach row called through .call and stopped during lowering, before the reach rule was
asked. Predicted for head 2, which lowers .call, not measured: s2_unknown_reach, .a and .ts, goes red too,
its .call judged in place of main's refusal and stopped there, since its thisArg's reach is unknown as well.
s1_class_instance stays green: its class instance holds no read field of its own, which stops in main's
words.

## (d) A restore let through to an object of another literal

In hostMemberValueRestore, old:

```go
		if !one {
			return l.notYet(assignment, writing+", a method moved between objects of different literals")
```

New:

```go
		if false && !one {
			return l.notYet(assignment, writing+", a method moved between objects of different literals")
```

Red: h4_layout_representation, .a and .ts, each `lowered, want ...`: x sits at index 0 in both literals,
so the layout rule alone admits the restore though one holds a number there and the other a string.
h4b_single_literal_restore stays green. Head 1 measured it on the check before hostMemberOneSite was
factored out of the restore; the edit above is that check's place now, and the red is predicted the same.

## (e) A .call let through with an object of another literal

In hostMemberValueThrough, old:

```go
	if !one {
		return l.notYet(member, held+" with an object of another literal")
```

New:

```go
	if false && !one {
		return l.notYet(member, held+" with an object of another literal")
```

The layout rule still runs, over the read's site alone. Measured on head 2 (J-msx1c62), with the .stop
files set to `lowers`: v2_same_layout_this and h3_reordered_layout both lower and agree with Node. Their
literals differ in layout, so native reads the fields by name. So the witness is v5_call_representation,
h4 through .call: a's method reads this.x as a number and is handed b, whose x holds a string at the
same index. Red, predicted until -2's job measures it: v5, .a and .ts, with its .stop at `lowers` the release build
prints what Node doesn't. v2b_function_expression_this doesn't move: it stops at the read, at the
function expression's this.

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
plain call is judged in place of main's refusal. It's the mutant's only witness: every other unknown-reach
row calls through .call and stops during lowering, before the reach rule is asked.

## (d) A restore let through to an object of another literal

In hostMemberValueRestore, old:

```go
			if site != sites[0] {
```

New:

```go
			if false && site != sites[0] {
```

Red: h4_layout_representation, .a and .ts, each `lowered, want ...`: x sits at index 0 in both literals,
so the layout rule alone admits the restore though one holds a number there and the other a string.
h4b_single_literal_restore stays green.

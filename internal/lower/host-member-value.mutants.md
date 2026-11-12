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

Red: unmeasured.

## (c) The reach rule dropped on the value read

In checkHostMemberValues, old:

```go
		if err := l.hostMemberReach(value.stop, reaches, index); err != nil {
```

New:

```go
		if err := l.hostMemberReach(value.stop, reaches, index); false && err != nil {
```

Red: unmeasured.

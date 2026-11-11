# Mutants for a host member written on a program object literal (#fj0vqtf, wave 1)

Each is one edit to internal/lower/host_member_write.go, run, then put back, with
`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -run '^TestHostMemberWriteAgreesOrStops$' -v`.

## (a) The reach rule dropped

In checkHostMemberWrites, old:

```go
		if err := l.hostMemberReach(write, store, reaches, graph.projectionIndex()); err != nil {
```

New:

```go
		if err := l.hostMemberReach(write, store, reaches, graph.projectionIndex()); false && err != nil {
```

Red: s3_class_instance, s4_exported_parameter, s5_arrow_parameter and s6_accessor_pair, .a and .ts, each
`lowered, want ...`: a class's instance, a receiver whose reach is unknown, and a literal with an accessor are
then stores with no rule left to stop them.

## (b) The holder rule judging the write site's type alone

In hostMemberLiterals, old:

```go
		holders := l.entriesRecordHandles(node)
```

New:

```go
		holders := []*checker.Type{l.checker.GetTypeAtLocation(write.target.AsPropertyAccessExpression().Expression)}
```

Red: s1_narrow_wide (Language's h1) and s2_narrow_wide_object, .a and .ts, each `lowered, want ...`: the write
through Wide fits Wide's read, and only the Narrow the literal is held as says it doesn't.

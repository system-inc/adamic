# Mutant for process.stdout._handle's views (#0asf98m)

The pins are `TestStdoutHandleViews` (internal/oracle/stdout_handle_view_test.go) over
internal/oracle/testdata/stdout_handle_view/: n1 and n2 stop at the _handle read in their .a and .ts twins;
n3 and n4 pin the earlier stops that close their roads. The controls are node_process_blocking.a and node_process_terminal.a (TestNodeProcess*), which still
lower and agree.

## The view check off

internal/lower/stdout_handle_view.go, `stdoutHandleView`. Old:

```go
	seen := []*checker.Type{l.checker.GetTypeAtLocation(node)}
```

New:

```go
	if node != nil {
		return ""
	}
	seen := []*checker.Type{l.checker.GetTypeAtLocation(node)}
```

Red, measured (Codex 01a1bf0b): `TestStdoutHandleViews/n1_fd.a`, `/n1_fd.ts`, `/n2_cast_chain.a` and
`/n2_cast_chain.ts` lower ("lowered, want ..."), and at the parent they print undefined natively where Node
prints the descriptor. n3 and n4 are stopped before the rule (the hiding rule, lowering's any stop), so the
mutant leaves them as they are.

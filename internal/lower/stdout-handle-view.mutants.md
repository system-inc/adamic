# Mutant for process.stdout._handle's views (#0asf98m)

The pins are `TestStdoutHandleViews` (internal/oracle/stdout_handle_view_test.go) over
internal/oracle/testdata/stdout_handle_view/: n1 to n4, each stopping at the _handle read in its .a and .ts
twin. The controls are node_process_blocking.a and node_process_terminal.a (TestNodeProcess*), which still
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

Red: `TestStdoutHandleViews/n1_fd.a` and `/n1_fd.ts` lower and print undefined natively where Node prints the
descriptor, and n2 to n4 lower or stop elsewhere, red on their stop text.

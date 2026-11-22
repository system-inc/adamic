# A group signalled after its leader's reap: mutant (#rcmjshy)

Capture kills its command's process group with `Kill(-pid)`. Once the leader is reaped and no member of its group
lives, the pid is free for an unrelated process to take and lead a group of its own, so a kill sent then can reach
processes childguard never started. `processGroup` signals only while the leader is unreaped: it sees the exit with
waitid's `WNOWAIT`, kills what the leader left, marks the group reaped under the lock ctx's kill takes, then reaps.

The mutant is one edit, applied alone on Linux, then reverted. Run
`go test ./internal/childguard/ -run '^TestCapture(Cancel)?SignalsTheGroupOnlyBeforeTheReap$' -count=1 -v`.

## The unconditional post-reap kill restored

File: internal/childguard/childguard.go, captureTo. After `stop()`, add the kill the code sent before this fix:

```go
	signal(cmd.Process.Pid)
```

Expected: `TestCaptureSignalsTheGroupOnlyBeforeTheReap` goes red at group_linux_test.go:51 with "signals, true where
after the reap: [false true], want one before it", and `TestCaptureCancelSignalsTheGroupOnlyBeforeTheReap` at
group_linux_test.go:69 with "[false false true], want two before it". The test's signal asks the kernel, not the
group's flag (waitid on a reaped child fails with `ECHILD`), and sends nothing once the leader is reaped, so the
mutant run itself signals no reused id.

# test262's 15 s child deadline counts CPU, not wall time

In the Adamic repository, `cmd/adamic-test262/run.go` runs each compiled native binary and each Node run of a test with `runCommand(15*time.Second, ...)` (run.go, the `native` and `node` calls near lines 328 and 341). That deadline is wall-clock time. On a loaded gate box a child that would finish in 2 s of CPU can wait more than 15 s for a core, so it reads as a timeout and twelve tests in the package go red with no change to the program (cache, edit and corpus tests reach it).

Make the 15 s mean what it is for, the program's own run time: a child is timed out when it has used more than 15 s of CPU, and a much longer wall-clock backstop (2 minutes) still catches a child that blocks without using CPU. One portable way, Linux and macOS both: start the child under `/bin/sh -c 'ulimit -t 15; exec "$0" "$@"'` so the kernel sends SIGXCPU past 15 CPU seconds, and classify a SIGXCPU death exactly as `TimedOut` is classified today. Any other way is fine if it keeps the same verdicts.

Done means:
- Your new test passes. Don't run the whole package: the fast gate runs it after your push and sends any red back with its first failure.
- A new test proves both halves: a child that spins forever is reported `TimedOut` (use a smaller CPU limit in the test so it takes a second or two, not 15), and a child that does 1 s of CPU work while the box is saturated (start `2 * runtime.NumCPU()` busy goroutines or processes for the test's duration) is not timed out even though a wall-clock limit of the same size would have killed it. Show that the second half fails on the old wall-clock code (run it once against the old code and paste the failure in the commit message).
- Nothing else changes: same execution struct, same output capture limits.

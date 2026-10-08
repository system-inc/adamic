# The fuzzer's 20 s run deadline counts CPU, not wall time

In the Adamic repository, `internal/fuzz/run.go`'s `execute` gives the Node oracle and the native binary 20 s each (`execute(directory, nil, 20*time.Second, ...)`, near lines 142 and 143) as a wall-clock `context.WithTimeout`. On a loaded gate box a run that needs 1 s of CPU can wait longer than that for a core, so it is reported as `TimedOut` and TestFuzzerSharesRuntimeLibrary and TestReduceKeepsTheSignature go red with no change to the program.

Make the limit count the child's CPU time: a child is timed out when it has used more than its limit in CPU seconds, and a wall-clock backstop of 2 minutes still catches a child that blocks without using CPU. `execute` already starts the child in its own process group and kills the group on cancel; keep that. One portable way, Linux and macOS both: start the child under `/bin/sh -c 'ulimit -t <seconds>; exec "$0" "$@"'` so the kernel sends SIGXCPU past the limit, and set `TimedOut` for a SIGXCPU death as well as for the backstop. Any other way is fine if it keeps the same verdicts. The 30 s compile calls (`adamic c` and `adamic js`) take the same treatment through the same function.

Done means:
- Your new test passes. Don't run the whole package: the fast gate runs it after your push and sends any red back with its first failure.
- A new test proves both halves: a child that spins forever is reported `TimedOut` (use a 1 or 2 s limit in the test), and a child that does 1 s of CPU work while the box is saturated (start `2 * runtime.NumCPU()` busy goroutines or processes for the test's duration) is not timed out under a limit a wall clock would have blown. Show that the second half fails on the old code (run it once against the old code and paste the failure in the commit message).
- `Run`'s fields and meaning are unchanged.

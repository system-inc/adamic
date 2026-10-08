# Step 38: honest checker benchmarks (#nxsh582)

Contract first. The finish line is TypeScript's checker compiled by Adamic beating
`typescript-go` on the same programs with identical baselines. A Node stand-in
for that checker is **pending**, never pass, regardless of its measurements.

## Command

`adamic-bench --manifest FILE --json FILE [--warmup 1] [--runs 5]
[--max-load 0.5] [--timeout 60s]`

Linux only. `--max-load` is an absolute one-minute system load average, not divided
by core count. Read it before any contestant starts and before every invocation;
above threshold (or unavailable) refuses to start. No override. Warmups must be
nonnegative, runs positive, timeout positive, load threshold finite/nonnegative.
JSON is newline-delimited, one record per warmup/timed invocation. The short table
goes to stdout. Exit 0 means the harness completed, **not** the native finish line
passed; exit 1 is mismatch or execution/measurement failure; exit 2 is invalid
configuration or load refusal. Partial runs never receive a summary.

## Declaration

Manifest paths resolve relative to the manifest. Commands are argv arrays, not
shell strings; executables must be on PATH or absolute. The first contestant is
the oracle. Names must be unique. Programs are source directories; each is copied
into a fresh temporary working directory for every invocation. Symlinks and
special files are rejected. `emit_dir` is a relative, initially absent directory
inside the program, containing **all** output artifacts.

```json
{
  "programs": [{"name":"small", "directory":"programs/small", "emit_dir":"out"}],
  "contestants": [
    {"name":"node-tsc", "command":["node","/absolute/tsc.js","-p","tsconfig.json"], "version_command":["node","/absolute/tsc.js","--version"]},
    {"name":"tsgo", "command":["/absolute/tsgo","-p","tsconfig.json"], "version_command":["/absolute/tsgo","--version"]},
    {"name":"adamic-native", "pending":true, "command":["node","/absolute/tsc.js","-p","tsconfig.json"], "version_command":["node","/absolute/tsc.js","--version"]}
  ]
}
```

The reserved `adamic-native` slot must currently declare `pending:true`. To replace
it, update this contract and supply evidence of an actual Adamic-compiled checker.
Version commands are required, run outside timing, and must succeed. Node users
should include the Node runtime version in their version command's output (a
wrapper is allowed). Pin compiler versions and library inputs in run evidence.

## Instrument and comparison

Instrument: Go `time.Now` monotonic elapsed time around `exec.Cmd.Run`, plus Linux
`wait4` resource usage (`ru_maxrss`, KiB). Wall includes process startup/teardown;
RSS is the kernel's maximum for the reaped command and waited descendants, not a
sum of concurrent process RSS. Commands must wait for their children and must not
daemonize. Timeout kills the process group and invalidates the run.

Oracle's first output is the baseline; every oracle repetition and every other
invocation (including warmups) must match its exit code, stdout, stderr, and sorted
relative emitted-file SHA-256 hashes. Only the fresh working directory's absolute
path is replaced with `<program>` in diagnostics; no whitespace, ordering, or
message normalization. Nonzero diagnostic exit codes are valid baselines. Spawn,
timeout and signal failures are errors, not diagnostic baselines. Inputs and
manifest have SHA-256 fingerprints. Changes outside emit_dir are rejected.

A mismatch invalidates **all** timings for that contestant/program: status
`differs`, wall/RSS null in every record and no time in the table. Pending identical
results can show instrument measurements but remain `pending`. Results compare to
the oracle as `win`, `loss`, or `tie` by median wall; both wins and losses appear.
No result here proves the finish line. Quantiles use linear interpolation at
`(N-1)*p`, timed runs only. Report median, p10, p90 for wall milliseconds and RSS
KiB. Run serially, rotating contestant order between repetitions.

## Record fields (schema 1)

Each record contains `schema`, `instrument`, `manifest_sha256`, `program`,
`input_sha256`, `contestant`, `command`, `version`, `pending`, `phase`
(`warmup`/`timed`), `iteration` (zero-based within phase), `status`
(`identical`/`pending`/`differs`/`error`), `exit_code`, `stdout`, `stderr`,
`emitted` (relative path → SHA-256), `wall_ms`, `peak_rss_kib` (nullable),
`started_utc`, `load_before`, `load_after`, `machine` (`cpu_model`, `governor`,
`cores`, `os`, `go_version`), and `error` (empty unless failed). Missing CPU or
governor metadata is explicitly `unavailable`; load or RSS unavailable is fatal.
Records are published after validation of the complete program, so no provisional
time escapes before a planted difference is discovered.

## Validation and evidence

Tests must plant diagnostic and emitted-file differences and verify null times;
a planted high load must refuse before executing any contestant. Real run evidence
must include the manifest, programs, JSONL, table, versions and machine metadata.
Small programs measure startup heavily and do not establish checker throughput.

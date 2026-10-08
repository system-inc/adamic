# Step 29 library startup and output checkpoint

Branch `codex/host29-startup-output`, based on library `71f91687` with
host surface `91215715` merged. The supplied refs have no `AGENTS.md`;
repository guidance was read from `CLAUDE.md`, `README.md`, `docs/0.1.md`
and `docs/memory.md`. This unit changes library lowering and its tests only.
Runtime's C primitive files are untouched.

The existing library already lowers `process.cwd`, `argv`, environment reads,
stdout writes, numeric exit and exitCode, performance and memoryUsage to native
calls. This checkpoint makes environment assignments and deletion explicitly
Refused under the step's live-read-only ruling. The Node declarations accept
string exit codes, but this base's native ABI represents numeric codes only;
those overloads now produce NotYet before emission instead of an invalid fit.

## Unchanged scout sources

The tests use the source fixtures retained by runtime scout `2cdc201b` without
rewriting their sys.ts bodies. Node is 24.19.0, on Linux.

| Fixture | Native and JavaScript | WASI | First compiler prerequisite |
| --- | --- | --- | --- |
| 14 getCurrentDirectory | Refused | No emission | `callback = undefined!` is a refused non-null assertion at 16:24 |
| 15 getExecutingFilePath | Agrees with Node | TargetRefused | Its scratch setup calls `fs.mkdtempSync`, unavailable on WASI |
| 16 getEnvironmentVariable | NotYet | No emission | String `process.env[name] \|\| ""` at 11:12; fixture environment writes would also be Refused |
| 17 write | NotYet | No emission | Object literal void method result at 17:5 is unrepresented |
| 18, 19, 20 exit | NotYet | No emission | Declared `activeSession: undefined` at 13:5 is unrepresented |
| 23 newLine | Agrees with Node | TargetRefused | Its scratch setup calls `fs.mkdtempSync`, unavailable on WASI |

Fixture 15 shadows `__filename` and `__dirname` with local strings. Its success
holds the sys.ts branch expression, not the still-pending executable globals.
No source-shape approximation is advertised as closing a compiler prerequisite.
WASI refusal is recorded separately so it cannot hide native/JavaScript results.

The binding-only `node_startup_environment.a` probe uses a nullish fallback to
exercise the live getter and distinguish missing from empty. It is explicitly
not fixture 16's unsupported truthy fallback. Externally supplied empty, zero
and Unicode values, and an absent key, agree on native, JavaScript and WASI.
These environment-dependent runs disable the oracle cache: its source Node
cache does not distinguish the externally supplied values.

## Mutants

Every scout source mutant runs cleanly on Node and disagrees only with the
original Node result: retain the memoize callback (14), test `sys.cjs` instead
of `sys.js` (15), return empty environment text (16), append `?` to writes (17),
change exit 0 to 1, 1 to 2, and 2 to 0 (18,19,20), and use CRLF (23).
The 15 and 23 mutants also compile and run without sanitizer errors on native
and JavaScript, and fail only the Node stdout comparison. Compiler-blocked
fixtures have no backend mutant success claim. The separate getter's empty-text
mutant compiles and runs cleanly on native, JavaScript and WASI and fails only
the Node stdout comparison.

## Runtime seam still pending

At the latest remote check, `refs/heads/runtime/host-exit-clocks` was absent.
Do not invent its C ABI or modify runtime's primitive implementation. Merge that
branch and read its section in `docs/scout-29-host.md` before completing:

- startup cwd/argv/environment bindings against its documented ownership and
  catchable Node error contract, including WASI capabilities;
- `process.execPath`, executable `__filename`/`__dirname`, the supported require
  mapping, and distribution of tsc's default libraries beside that executable;
- one stable `process.stdout` object, its absent private setBlocking probe,
  and stdout writing under the runtime output contract;
- exit and exitCode, flushing and termination without running finally blocks;
- performance.now/timeOrigin, Date.now and the supported memoryUsage fields.

The base already has calls named `adamic_node_cwd`, `adamic_node_argv`,
`adamic_process_environment`, `adamic_node_stdout_write`,
`adamic_process_exit`, `adamic_process_exit_code`,
`adamic_process_set_exit_code`, `adamic_node_performance_now`,
`adamic_node_time_origin` and `adamic_node_memory_usage`. Their existence does
not establish the new contract or all-backend support. Source-map hooks remain
NotYet. No branch is pushed while this runtime seam is pending.

## Validation

Run output stays outside the checkout. The finalized targeted command is:

```text
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_ORACLE_WASI=1 go test ./internal/lower ./internal/oracle -run '^TestNodeStartup' -count=1 -v -timeout 10m
```

It passes, with the explicit compiler/target skips above. Log:
`/tmp/host29-verification.log`. The existing newline comparison and emitted
native/JavaScript mutants also passed (`/tmp/host29-final.log`). Linux counts
are refreshed with `TestCountsAreRecorded -update-counts`, with output in
`/tmp/host29-counts-final.log`. Full repository tests are left to the fast gate.

The Linux getter count row is 2 allocations, 2 frees, 5 retains, 8 releases,
peak 2, regions 0. The existing directory-system row also changes because
it enumerates the fixture directory containing the new source: from
3220/3220/5812/4829/1938/0 to 3292/3292/5959/4937/1992/0. All other existing
rows remain unchanged. Targeted vet and diff/format checks pass. macOS was not
run. The runtime remote ref was checked again after validation and remained
absent, so this is an unpushed local checkpoint, not a completed runtime binding.

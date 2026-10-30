# Stdout columns lifecycle mutants

This evidence backs #jxq7ws3's stdout width cache and the pty lifecycle harness.
It changes no stage3 record or oracle outcome policy. The Node side always runs
through oracle/node.mjs; the comparison is the existing disagreement function.

Prepare the package's TestProduct_NativeRuntimes families in product order,
then TestProduct_StdoutColumnsProbes, and run:

```
ADAMIC_ORACLE_STDOUT_PTY=1 go test ./internal/oracle -count=1 \
  -run '^TestStdoutIsTTYMatchesNodeOnPty$' -v
```

The product compiles six independent mutants of stdout_tty_impl.h. Public
symbols are wrapped so incoming calls from async.c reach the mutant's own cache,
not the control archive. Every control is compared exactly to Node first.

| Mutant | Red control under TestStdoutIsTTYMatchesNodeOnPty | Observed catch |
| --- | --- | --- |
| Refresh on each read | ColumnsSynchronousResize | stderr differs |
| Never refresh | ColumnsAwaitResize | stderr differs |
| Do not clear the signal flag | ColumnsOneRefreshPerSignal | expected first ioctl plus exactly one refresh per signal |
| Leave broken stdout clear after failure | ColumnsRefreshFailure | exit codes differ |
| Skip refresh when the cache is unknown | ColumnsInitialFailureRecovery | stderr differs |
| Dispatch a host checkpoint before initial microtasks | ColumnsResolvedAwaitResize | stderr differs |

The mutant subtests pass only after observing the named red comparison. A
compiler error, missing initial control observation, sanitizer failure or changed
Node control fails the harness; none counts as a caught mutant. The one-refresh
case counts actual TIOCGWINSZ calls through a host injector, separately from
stdout and stderr. The first-ioctl fixtures return EIO until their parent allows
ioctl again, including on Node's duplicated tty descriptor. No Node method or
JavaScript property is replaced.

`resolved_await.a` is compiled through both backends and compared to its source
on Node. The premature-turn mutant calls the completion checkpoint before the
ordinary async runner, recreating the observed native `later 120` versus Node's
`later 80`. The fixed runner drains initial microtasks before entering live host
turns. Error fixtures put their readiness marker on a private descriptor, so
both compared output streams are empty at status 70.

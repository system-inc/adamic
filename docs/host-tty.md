# Host terminal reads

`process.stdout.isTTY` is `boolean | undefined`, from Adamic's served Node
host axioms. A plain property read is bound by the Node library symbol. Native
and WASI call `isatty(1)`: a terminal gives true, every other result gives
undefined. JavaScript reads Node's member directly. Writes, deletes, optional
chains, stdin and stderr remain NotYet with a repair. The stream is not a value.
The host read is a NodeFSFile operation, so forcing it into Wasm is impure.

`process.stdout.columns` is `number | undefined`. The native cache is
initialized on stdout's first use. Its first ioctl may leave it undefined;
a successful SIGWINCH refresh can establish it later. Reads and initialization
use the output lock. Synchronous reads never refresh. At an async loop turn,
the pending signal refreshes before completions; failure ends like the oracle's
stdout error listener, with status 70 and no error diagnostic. JavaScript reads
Node directly. Columns is refused for WASI preview1: it reports no terminal size.

The signal handler publishes a flag and a byte to a separate nonblocking,
close-on-exec pipe. Neither descriptor is closed or reused. Default host waits
poll this pipe beside their completion pipe. Custom hooks do not poll it: their
next completion refreshes width, later than Node, never earlier. This remains
a host-integration gap. Host liveness and the external stop-signal set do not
change. No new runtime object shape is introduced.

The pipe, file and `/dev/null` oracle units always run, comparing Node, native
under ASan and UBSan, and the JavaScript backend. The existing
`ADAMIC_ORACLE_WASI=1` oracle lane adds WASI to each comparison. Their fixture
writes its observation to stderr as well, so `/dev/null` still proves the value.

The pty overlay is `TestStdoutIsTTYMatchesNodeOnPty`, enabled with
`ADAMIC_ORACLE_STDOUT_PTY=1`. Platforms runs this lane by hand. It uses
`script -qec`, whose presence and a usable `/dev/ptmx` cloud/setup.sh does not
promise on Loom runners. The switch's named skip needs Loom classification;
this repository does not carry judge/testdata/census/census-extra.json. Once
enabled, missing tools or terminal support fail, rather than adding skips.
The isTTY comparison pins true or undefined. The columns lifecycle protocol
sets the same pty geometry for Node and native and compares the observed values.
It covers synchronous caching, a real pending host request's loop turn, one
refresh per signal, refresh failure, and all three initial-ioctl failure cases.
The language currently has resolved-promise microtasks but no pending host I/O
binding; `loop.c` therefore uses the public host bridge for the runtime turn,
held to the Node-side `loop.mjs` through `oracle/node.mjs`. The ordinary `.a`
fixture separately proves lowering and native/JS emission on every destination.
The failure injector operates on TIOCGWINSZ, including Node's duplicated tty fd;
it never patches JS properties. The manual pty lane is Linux-only evidence.

The three boundary mutants are reproducible with
`python3 internal/lower/testdata/stdout-istty-mutants.py`, after the oracle's
sanitized and WASI runtime products have been built ahead. They must turn red
`TestStdoutIsTTYMatchesNodeOnPipe`, `TestNodeStdoutIsTTYRefusesOtherStreams`,
and `TestNodeStdoutIsTTYRefusesMutation`, respectively.

The columns lifecycle controls and five independent mutants run below that same
pty guard. `TestProduct_StdoutColumnsProbes` prepares their input-keyed binaries
and ioctl injector ahead. The mutants refresh on reads, omit refresh, leave the
flag set, leave broken stdout clear on failure, and skip an unknown cache's
refresh. Their red subtests are SynchronousResize, AwaitResize,
OneRefreshPerSignal, RefreshFailure and InitialFailureRecovery, respectively.
The helper uses the house's `disagreement` comparison directly; no error-stack
normalization or oracle policy is added. No stage3 record changes.

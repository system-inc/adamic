# Host terminal reads

`process.stdout.isTTY` is `boolean | undefined`, from Adamic's served Node
host axioms. A plain property read is bound by the Node library symbol. Native
and WASI call `isatty(1)`: a terminal gives true, every other result gives
undefined. JavaScript reads Node's member directly. Writes, deletes, optional
chains, stdin and stderr remain NotYet with a repair. The stream is not a value.
The host read is a NodeFSFile operation, so forcing it into Wasm is impure.

`process.stdout.columns` remains NotYet. Node caches its width when stdout is
created and refreshes it through SIGWINCH on the event loop. A fresh ioctl and
an immutable cache both disagree with Node; this slice adds neither.

The pipe, file and `/dev/null` oracle units always run, on native under ASan and
UBSan, the JavaScript backend and the WASI oracle. Their fixture writes its
observation to stderr as well, so `/dev/null` still proves the value.

The pty overlay is `TestStdoutIsTTYMatchesNodeOnPty`, enabled with
`ADAMIC_ORACLE_STDOUT_PTY=1`. Platforms runs this lane by hand. It uses
`script -qec`, whose presence and a usable `/dev/ptmx` cloud/setup.sh does not
promise on Loom runners. The switch's named skip needs Loom classification;
this repository does not carry judge/testdata/census/census-extra.json. Once
enabled, missing tools or terminal support fail, rather than adding skips.
No fixed pty width is asserted. The only terminal value admitted here is true.

The three boundary mutants are reproducible with
`python3 internal/lower/testdata/stdout-istty-mutants.py`, after the oracle's
sanitized and WASI runtime products have been built ahead. They must turn red
`TestStdoutIsTTYMatchesNodeOnPipe`, `TestNodeStdoutIsTTYRefusesOtherStreams`,
and `TestNodeStdoutIsTTYRefusesMutation`, respectively.

# Process operations

The global prelude process supports numeric exit codes, stdout and stderr terminal observations,
and read-only environment values. Recognition uses the prelude declaration's symbol, so a program's
own object named process keeps its own behavior.

`process.exitCode` is `number | undefined`. It starts undefined. Assigning a finite integer stores
its signed 32-bit conversion, as Node 24 does, while the assignment expression still returns the
original right-hand value. Assigning undefined clears it. Normal completion releases the program's
values, flushes output and ends with the chosen status, or 0 when none was chosen. The operating
system observes the low eight bits on Unix. Uncaught throws still end with Adamic's panic and exit 70.

## Named oracle exception: no deliberate output loss

Decided by @system_adamic, October 2026: Adamic never loses output on purpose. On explicit
`process.exit`, native flushes synchronously and the JavaScript backend waits for both stdout
and stderr to finish writing before exiting with the chosen status. Raw Node can discard
queued writes to a backpressured pipe; that loss is documented Node behavior, not Adamic's
expected answer. The Linux oracle pins raw Node's 4,096-byte result for a 204,800-byte payload
and independently requires both Adamic backends to preserve all 204,800 bytes. Every other
covered exit path still compares byte for byte to Node. The exit-70 panic convention remains.
See [the draining report](process-drain-report.md) for the observations and executable mutant.

`process.exit(code?: number): never` stops source execution immediately, skipping finally blocks
and frame cleanup. Process termination waits for output to drain.
An omitted argument uses exitCode. An explicit undefined argument means 0 on Node 24. Fractional
and nonfinite numbers throw a catchable RangeError without changing the stored status. Numeric
strings are outside this prelude's admitted types. Direct calls are supported; storing process,
its streams, its environment object or its exit function as ordinary object values is NotYet.

Explicit exit stops source execution without running catch or finally, then drains output.
The JavaScript backend uses an internal stop signal; generated catches rethrow it and generated
finally blocks skip it, so waiting for the host streams cannot resume source execution.
File output and interleaved shared-descriptor order match Node. The earlier output observations
are preserved in [the historical output report](process-output-report.md). Counts at immediate
exit record the values still held, as panic counts do; LeakSanitizer's exit hook cannot inspect an
immediate `_exit`, including one with status 0.

`process.stdout.isTTY` and `process.stderr.isTTY` are true on a terminal and undefined on a pipe or
regular file, independently. `process.env[name]` and `process.env.NAME` return a copied string or
undefined for a missing name. Empty values remain empty strings. Names and values use Node's UTF-8
conversion, including replacement of lone surrogates and NUL-terminated names. Environment writes
are excluded by the readonly index signature. NO_COLOR and FORCE_COLOR are ordinary observations;
the compiler imposes no color policy. Pinned Go cohere checks NO_COLOR's presence, even when empty,
and does not read FORCE_COLOR. Its stream test is a character-device stat rather than isatty, which
also accepts devices such as /dev/null; this distinction remains relevant to its output port.

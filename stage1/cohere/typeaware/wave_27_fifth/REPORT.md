Built require-await, symbol-description and valid-typeof as native .a callbacks.
Claim pushed at f2a14e47f before implementation; landing revalidation follows.
185 controls/107 findings and both corpora match full Go bytes, normal/sanitized.
Three rule mutants exit 0 and are caught by Go bytes; retained-handle mutant caught.
React analysis claims remain parked; shared registration and emitted JS untested.

## Initial validation on main f8013f0b

`source /workspace/adamic-tools/env.sh` then the owned validate.py runner with
artifacts `/workspace/wave-27-scratch/fifth-final` and the current-main compiler
`/workspace/wave-27-scratch/f801-third/adamic`. The compiler sources were unchanged
through this port. Full command and all subprocess outputs are retained under
evidence-f801. Toolchain setup earlier passed ready=0s, warm-cache=116s,
total=116s; nproc=5. The touched checker package passed in 0.134s; focused
checker/typeaware vet exited zero with empty output.

209 extracted/supplemental candidates yielded 185 Go-parseable files. The 24
excluded malformed fragments are retained and are not counted as passes.
Normal and sanitized controls matched 64,291 bytes and 107 findings. The
287-file frozen repository corpus matched 18,485 bytes and zero findings;
the 77-file TypeScript compiler corpus matched 5,934 bytes and zero findings.
The positive controls prevent that clean corpus result from being vacuous.
All 351 numeric constants and the three listener metadata arrays match an
independent Go enum oracle. Native ASan/UBSan/LeakSanitizer runs have empty stderr.

Rule mutants invert the first-declaration ambient test (byte 39130), admit the
misspelling strng instead of string (byte 46614), and bypass the promise contract
(byte 33622). Each compiles, exits zero and prints no stderr; only the independent
Go diagnostic/suggestion byte comparison catches it. All four new query operations
reject a released handle with exact panic 70. A Go overlay retaining the handle
exits zero on a valid symbol query; the required-panic assertion catches it.

Three alternating whole-process samples have medians: repository native
619.395ms versus Go 284.208ms (2.179x); compiler native 3,312.937ms versus Go
380.613ms (8.704x). Measurements include parsing, checking, rendering and teardown,
not just rule time. This does not claim a speedup. The string-to-numeric adapter
and native parsing costs remain until the shared numeric parser/driver lands.

Two differential failures during development were corrected: the pinned checker
union flag differs from upstream TypeScript, and implements heritage syntax differs
between the two parsers. A new raw heritage question avoids shared parser edits.
The added concise nested-function witness exposed Go's initial child-walk behavior,
which is preserved exactly. Initial failed streams are scratch observations only;
the retained successful evidence reflects the final source.

Full repository tests, emitted-JavaScript comparison, nondefault options and
shared registry discovery were not run. No protected compiler source, shared
registration generator or shared test harness was edited. The four original
raw questions and this new question each have their own Go/Adamic source files.

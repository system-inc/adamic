Temporary: comes out when library's Node-types loader lands on main

This becomes permanent in adaptation 40 once library's loader is on main.
Until then, the checker cannot see @types/node 25.3.3's declaration of
ErrorConstructor.captureStackTrace without the pinned Node declaration loader.

Plan published before implementation: slice debug.ts only, replace both
`(Error as any).captureStackTrace` expressions with `Error.captureStackTrace`.
Keep the condition, call, arguments, Error construction, throw, and every other
statement. Stock TypeScript 6.0.3 must emit byte-identical JavaScript before
and after; also compare all 509,014 tokens to the full-tree Node oracle.
Rebuild with ADAMIC_NATIVE_SPLIT=0 and with ADAMIC_NATIVE_SPLIT=1/jobs=nproc,
then compare native tokens and collect emitted-C/build metrics if it builds.

The previously retired adaptation 80 for flag return inference is archived
under drivers/scanner/retired-adaptations; this number is reused as instructed.

Implemented: both casts removed; stock TypeScript 6.0.3 emits the same 805 bytes
(SHA256 2694db22c5088b6adb85a79a981cd282b58574334cc13a6749175663ce0d16af).
A changed capture argument fails the JavaScript byte-identity check. Node still
emits 509,014 identical tokens. The driver activates the pinned Node loader
with an erased type-only node:util import; --compiler-cwd selects its checkout.
Both native modes now stop at TS2740: adaptation 55's {} marker is not Function.
The unmodified call arguments are retained. Full baseline rerun is pending
after correcting a test-fixture copy error in the first attempt.

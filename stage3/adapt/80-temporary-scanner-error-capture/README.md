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

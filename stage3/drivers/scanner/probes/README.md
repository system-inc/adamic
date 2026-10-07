# Scanner blocker probes

`uninitialized-non-null.a` is the minimal literal-initializer blocker. Node's
oracle prints `assigned before read`. On non-null-check 6ae58a0 alone and the
refreshed integration scratch compiler it returns:

```
stage 0 can't lower a value of type never yet
```

This is not a nullish runtime failure: no executable is produced. The compiler
must recognize literal undefined!/null! as uninitialized declarations and check
reads. Removing `= undefined!` from this minimal control compiles and prints the
same line, demonstrating that ordinary assignment-before-read works. That
control does not implement the exact scanner initializer requested by the unit.
The scanner also assigns through nested setText, requiring its write to be known.

`stack-marker.cjs RAW_DEBUG_SLICE ADAPTED_DEBUG_SLICE` uses stock TypeScript
(`SLICE_TYPESCRIPT`) and checks failure messages and V8's explicit marker identity.
`SCANNER_MARKER_MUTANT=1` replaces the supplied marker with fail and must exit 1.

Copy `assert-failure.a` into a runner output directory (beside its adapted
symlink) and invoke scanner/node.mjs with the same SCANNER_TYPESCRIPT and
SCANNER_RUNTIME as run.py. It calls a real reached assertion incorrectly and
prints the exact failure message. Original and adapted messages match.

The token-end and declaration-omission mutants remain documented in BLOCKERS.md.
These probes do not claim that the scanner runs natively.

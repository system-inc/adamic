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

Current live readiness bc9f5d7 compiles uninitialized-non-null.a; native and
Node both print assigned before read. The old never diagnostic above is
historical. A scratch namespace/readiness merge artifact initially panicked;
it was isolated (the feature alone passed) and corrected before recording the
integrated result. No production compiler change is included in this unit.

error-constructor-value.a is the remaining compiler blocker: Node prints
captureStackTrace available; latest integrated stage 0 cannot lower reading
Error. Removing the actual capture call is caught by the stack-marker probe.

assert-unknown.a shows why adaptation 56 remains: proven assertion admission
succeeds, but the unknown parameter cannot lower. Node prints true.

scanner-state.a compares initial missing token values, callback state restoration,
JSX default versus false, and known versus unknown regexp keys. Copy it next to
a runner's adapted symlink and use node.mjs as for assert-failure.a. Independent
mutants shift scanRange's start, change the JSX default to false, and replace the
g table key with q. Each runs normally and is caught solely by output diff.

scanner-lazy-text.a preserves the original textInitial! and setText(text)
sequence. Integrated bc9f5d7 compiles it, then exits 70 on missing initial text:
read before assignment. Node prints [] and [source]. The feature alone stops
earlier at nested setText lowering. Therefore keep adaptation 57's rewrite.
scanner-captured-token-value.a builds and matches Node on the integration:
assigned before read 20, then assigned again. Together with the literal
undefined! control, this removes 58 from the selected profile. This proves
assigned captures; it does not license reading an absent required token value.

Discovery probes use the unpushed integration compiler, behind scratch stubs.
The optional-method probe is independent: native 0 versus Node 11. Trace
shows omitted start/length present natively. Required-method and explicit
undefined controls produce 11. Property initialization panics natively versus
Node 0; captured local initialization matches ([] and [source]). These are
observations on the integration build, not attribution to an individual tip.

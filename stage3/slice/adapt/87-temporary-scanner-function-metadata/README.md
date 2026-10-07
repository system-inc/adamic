Temporary: comes out when sound AnyFunction widening lands

Plan published before implementation. In the scanner slice's Debug.fail only,
annotate the captureStackTrace fallback as `(fail as Function)`. Retain upstream
AnyFunction marker parameters and every runtime statement and call argument.
The annotation is a checked metadata view, never invoked through Function.
Stock TypeScript must emit byte-identical JavaScript; a changed capture argument
must fail that check. This clears method-signature-style widening with compiler
census-small-families 8d34357's optional-function-value support.


Implemented and validated: both expanded slice Node dumps match the full-tree
reference (1,369,432 tokens, 466 errors). Full baseline: 106,366 passing; its
single API mismatch is exactly reconstructed from sanctioned owners, with all
60,930 other references unchanged. The next native stop is append's correlated
overload results, core.ts:34:1. Detailed controls/mutants and compiler pins are
in drivers/scanner/evidence/fallback-retry.json and BLOCKERS.md.

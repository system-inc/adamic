Temporary: comes out when sound AnyFunction widening lands

Plan published before implementation. In the scanner slice's Debug.fail only,
annotate the captureStackTrace fallback as `(fail as Function)`. Retain upstream
AnyFunction marker parameters and every runtime statement and call argument.
The annotation is a checked metadata view, never invoked through Function.
Stock TypeScript must emit byte-identical JavaScript; a changed capture argument
must fail that check. This clears method-signature-style widening with compiler
census-small-families 8d34357's optional-function-value support.

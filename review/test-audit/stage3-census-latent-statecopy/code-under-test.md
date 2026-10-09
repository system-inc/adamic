CODE UNDER TEST: statecopy generator, not lowering or the Go compiler.
ORACLE: self-written named-container and foreign-pointer diagnostics, nonzero subprocess result, and no output publication.

Complete package function inventory reached by the rows:
- main (entry): argument admission, input discovery, AST type inventory, struct traversal, output construction and publication.
- generator.copy: Ident and StarExpr rejection branches reached by the fixtures.
- printed: Go AST formatting reached by the foreign-pointer path.
Other branches in generator.copy are not executed by the two fixtures.

No tests, fixtures or oracle expectations are mutated. Six production changes are frozen here before outcome checks, spread across all three reached functions. P1 is a separate empty main-entry probe. The tests remain independent rows because they assert different rejection guards and distinct diagnostics; there is no shared fixture-check helper.

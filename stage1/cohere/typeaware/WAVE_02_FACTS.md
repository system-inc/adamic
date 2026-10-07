# Wave 02 checker questions

The native `.a` adapters are `awaited_shape.a` and `constraint_shape.a`. The
corresponding Go files are `bridge/tsgo/checker/awaited_shape.go` and
`constraint_shape.go`. The existing dispatcher has only the two new case/return
registrations. No checker implementation or compiler lowering was changed.

Both questions take `mode\nTYPE_ID`. TYPE_ID must be a canonical decimal,
nonzero identity previously issued by the same live program. Requests retain the
existing exact source-file/span/kind validation. Malformed, absent, out-of-range,
and released identities/handles are rejected rather than guessed.

* `awaited-shape` calls `Checker_getAwaitedType` on that identity.
* `constraint-shape` calls `Checker_getBaseConstraintOfType` on that identity.

Both return the existing version-1 type-graph frame: mode, strictNullChecks,
presence, zero-or-one root identities, zero comparison roots, and type records.
An absent result is explicitly absent, with no root. Identity and graph framing
reuse the existing program-owned graph encoder. These are raw checker facts;
all lint judgments, diagnostics, fixes and suggestions remain in Adamic.

`TestWave02CheckerQuestions` compares both returned identities against the direct
checker operations, tests malformed identities and inexact nodes, and pins the
Instantiable and Never flags used by the Adamic consumer. Each operation has a
compiling overlay mutant that returns the input type instead; the direct identity
comparison kills it. The native suite separately proves a query after release
panics with status 70 and kills a registry mutant that keeps the handle live.

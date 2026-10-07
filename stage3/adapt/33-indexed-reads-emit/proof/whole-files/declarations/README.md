# declarations.ts: 2 -> 1

Read the entire 1977-line file before editing. props[0] is a U-endpoint read:
isExpandoFunctionDeclaration requires a user-named expando Value export;
getPropertiesOfContainerFunction obtains that same parse-tree function symbol,
resolveAnonymousTypeMembers uses its exports, and getNamedMembers retains that
Value property. The local populated list is not mutated before its first read.
The existing ledger entry becomes an assertion; its existing parent assertion
stays unchanged. No read moves or defaults.

The remaining TS2345 at clause.types[0] (line 1678) is declined with a concrete
counterexample. getEffectiveBaseTypeNode only proves the first extends clause
has an element; map visits every extends clause. Stock Node parses the recorded
source with two clause lengths [1,0] and no parse diagnostics, then declaration
emit crashes reading kind. Nonemptiness of that second list cannot be asserted.
The exact remainder is in after.json; the observation is in counterexample.json.

All-code pinned latent census 2 -> 1. Stock JavaScript and adapter idempotence
pass. The props[0]! -> ?? 0 mutant fails both independent checks. Default oracle:
106367 passing, zero failing/pending, empty baseline diff,
213.824 seconds. Mechanical API projection accepts only 20's 189 and
40's 28 lines; every other reference is byte-identical. This file is not closed.
Reproduce with the visitor/sourcemap commands, the before snapshot
/tmp/emit33-close-source-before-declarations, and MUTANT_FILE=transformers/declarations.ts.

# Mutants for brand mints (#yth65gz, ruled on #b5w3ycg Nov 5)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| No mint | brand_mint.go, `brandMintOf`: return nil first | TestBrandMintsOfTscsSortedArraysAreAdmitted, every row Refused (a cast the runtime can't check) where NotYet at the value is wanted |
| A non-phantom member admitted | brand_mint.go: `!l.phantomBrandMember(property) ||` made `false ||` in the declared-member walk | TestBrandMintRecognitionIsExact, the real-member row passes the cast |
| The hop judged alone | brand_mint.go, `mintHop`: return false | TestBrandMintsOfTscsSortedArraysAreAdmitted, the four tsc shapes refused at the inner `as unknown` |
| No report line | cmd/adamic/checks.go: the brand mint loop not printed | TestExplainChecksListsBrandMints |
| A redeclared base member admitted (Language's must) | brand_mint.go: the walk skips a declared member the base also has (`continue`), as the first head did | TestBrandMintRecognitionIsExact, the n01, n07, n08 and merged-declaration rows pass the cast |
| Any literal exempt (the -3 head) | cast_proof.go, primitiveLiteral: return true first | TestBrandMintRecognitionIsExact, the spread row passes the cast |
| No literal exempt | cast_proof.go, primitiveLiteral: return false first | TestBrandMintOfAPrimitiveLiteral, the object row refused |
| Arrays skip the walk (the -4 head) | cast_proof.go, castAssignableAt's array case: its element walk dropped | TestBrandMintRecognitionIsExact, the named-element row passes the cast |
| Arrays walked by members | cast_proof.go, castAssignableAt's array case removed | TestBrandMintsOfTscsSortedArraysAreAdmitted, the emptyArray, deduplicated and toSorted rows refused |

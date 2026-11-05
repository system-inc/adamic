# Mutants for brand mints (#yth65gz, ruled on #b5w3ycg Nov 5)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| No mint | brand_mint.go, `brandMintOf`: return nil first | TestBrandMintsOfTscsSortedArraysAreAdmitted, every row Refused (a cast the runtime can't check) where NotYet at the value is wanted |
| A non-phantom member admitted | brand_mint.go: `!l.phantomBrandMember(property) ||` made `false ||` in the declared-member walk | TestBrandMintRecognitionIsExact, the real-member row passes the cast |
| The hop judged alone | brand_mint.go, `mintHop`: return false | TestBrandMintsOfTscsSortedArraysAreAdmitted, the four tsc shapes refused at the inner `as unknown` |
| No report line | cmd/adamic/checks.go: the brand mint loop not printed | TestExplainChecksListsBrandMints |
| A redeclared base member admitted (Language's must) | brand_mint.go: the walk skips a declared member the base also has (`continue`), as the first head did | TestBrandMintRecognitionIsExact, the n01, n07, n08 and merged-declaration rows pass the cast |
| A literal judged by the optional-field walk | cast_proof.go, `freshLiteral`: return false | TestBrandMintsOfTscsSortedArraysAreAdmitted, createSortedArray's `[]` Refused; TestBrandMintRecognitionIsExact, the .a twin reads "a cast the runtime can't check" |

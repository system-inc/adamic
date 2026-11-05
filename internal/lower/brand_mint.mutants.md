# Mutants for brand mints (#yth65gz, ruled on #b5w3ycg Nov 5)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| No mint | brand_mint.go, `brandMintOf`: return nil first | TestBrandMintsOfTscsSortedArraysAreAdmitted, every row Refused (a cast the runtime can't check) where NotYet at the value is wanted |
| A non-phantom member admitted | brand_mint.go: `!l.phantomBrandMember(property) ||` made `false ||` in the declared-member walk | TestBrandMintRecognitionIsExact, the real-member row passes the cast |
| No report line | cmd/adamic/checks.go: the brand mint loop not printed | TestExplainChecksListsBrandMints |
| A redeclared base member admitted (Language's must) | brand_mint.go: the walk skips a declared member the base also has (`continue`), as the first head did | TestBrandMintRecognitionIsExact, the n01, n07, n08 and merged-declaration rows pass the cast |
| Arrays walked by members | cast_proof.go, castAssignableAt: the array term dropped from the skip | TestBrandMintsOfTscsSortedArraysAreAdmitted, the emptyArray, deduplicated and toSorted rows refused |
| Object literals walked by members | cast_proof.go, castAssignableAt: the object-literal term dropped from the skip | TestBrandMintOfALiteralIsExact, the object row refused |
| The operand not its origin | cast_proof.go, castAssignableAt: optionalRelationFailure given nil for the operand | TestBrandMintOfALiteralIsExact, the object row refused |
| No hidden-field judge on the value | cast_proof.go, castAssignableAt: the optionalRelationFailure return made true | TestBrandMintRecognitionIsExact, the named-element, named-array and spread rows pass the cast |

mintHop is gone: its mutant survived on Golem (job of aea74b09), because an `as unknown` on its own is an
upcast castProof admits as on main, so nothing needed erasing.

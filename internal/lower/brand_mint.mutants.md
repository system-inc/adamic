# Mutants for brand mints (#yth65gz, ruled on #b5w3ycg Nov 5)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| No mint | brand_mint.go, `brandMintOf`: return nil first | TestBrandMintsOfTscsSortedArrays, TestBrandMintIsTheOperandsOwnValue (refused: a cast the runtime can't check); TestBrandMintsAgreeWithNode |
| A non-phantom member admitted | brand_mint.go: the `if !l.phantomBrandMember(member) { return nil }` test skipped | TestBrandMintRecognitionIsExact, the real-member row lowers |
| The hop judged alone | brand_mint.go, `mintHop`: return false | the four tsc shapes refused at the inner `as unknown` |
| No report line | cmd/adamic/checks.go: the brand mint loop not printed | TestExplainChecksListsBrandMints |
| A redeclared base member admitted (Language's must) | brand_mint.go: the declared-member walk replaced by `if l.checker.GetPropertyOfType(base, member.Name) != nil { continue }` over the target's properties | TestBrandMintRecognitionIsExact, the n01, n07 and n08 rows lower |

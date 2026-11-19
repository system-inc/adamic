# Receiver freshness mutant (#69vkn0j)

Using a scratch Go overlay, replace the body of `freshObjectLiteral` in `invariance.go` with `return true`. This disables the receiver check at every caller. Run `go test -overlay=<scratch>/overlay.json ./internal/lower -run '^TestLiteralReceiverFreshness' -count=1 -v`.

| Row | Original, .a and .ts | Mutant, .a and .ts |
| --- | --- | --- |
| f1 | Refused at 2:57 | Lowers, test fails |
| f2-guarded | Refused at 2:57 | Lowers, test fails |
| f3 | Refused at 2:57 | Lowers, test fails |
| Arrow member | Lowers | Lowers |
| Exact receiver slots | Lowers | Lowers |
| Literal without methods | Lowers | Lowers |

The original suite exits 0 and the mutant exits 1. The stored-receiver test has exactly two failing rows, f1 and guarded f2. The full suite also catches the change in f3, because f3's method declares an explicit `this` parameter, even though it only reads the tag. Holding the f3 refusal fixed while disabling the check would change what the mutant measures.

The requested f2 source is rejected during checking because Adamic's console accepts strings. After converting the log argument, the bare assertion also fails to lower in a `.a` source. The guarded variant tests the same length access after a proven array-element read and reaches the declaration relation in both formats.

## Contextual ThisType mutant (#n7mgy3s)

Using the same overlay, replace the body of `literalOwnThis` in `invariance.go` with `return true`. The new arm then always reports a member's `this` as the literal's own, while the declared-this arm stays as it is. Run `go test -overlay=<scratch>/overlay.json ./internal/lower -run '^TestLiteral(Receiver|ThisType)Freshness' -count=1 -v`.

| Row | Original, .a and .ts | Mutant, .a and .ts |
| --- | --- | --- |
| g1 | Refused at 3:48 | Lowers, test fails |
| g2 | Refused at 3:48 | Lowers, test fails |
| Read-only getter under ThisType | Refused at 2:48 | Lowers, test fails |
| Getter whose this is the literal's own | Lowers | Lowers |
| Arrow member under ThisType | Lowers | Lowers |
| Getter under ThisType without this | Lowers | Lowers |
| f1, f2-guarded, f3 | Refused at 2:57 | Refused at 2:57 |

Exactly six subtests go red, g1, g2 and the read-only getter in each of .a and .ts, which is the hole as it stood on main: g1 and g2 are the evidence fixtures (stage3/meter/evidence/this-parameter-getter), accepted there. The f rows hold, since their refusal comes from the declared-this arm, which the mutant leaves alone.

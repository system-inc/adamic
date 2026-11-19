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
| Arrow member | Lowers | Lowers |
| Getter under ThisType without this | Lowers | Lowers |
| f1, f2-guarded, f3 | Refused at 2:57 | Refused at 2:57 |

Exactly six subtests go red, g1, g2 and the read-only getter in each of .a and .ts, which is the hole as it stood on main: g1 and g2 are the evidence fixtures (stage3/meter/evidence/this-parameter-getter), accepted there. The f rows hold, since their refusal comes from the declared-this arm, which the mutant leaves alone.

## Outer expression mutant (#rj9wg3b)

Using the same overlay, restore `ast.SkipParentheses(member.AsPropertyAssignment().Initializer)` in `freshObjectLiteral` in place of `ast.SkipOuterExpressions(..., ast.OEKAll)`, and change nothing else. Run `go test -overlay=<scratch>/overlay.json ./internal/lower -run '^TestLiteral(WrappedReceiver|ThisType)Freshness' -count=1 -v`.

| Row | Original | Mutant |
| --- | --- | --- |
| c02-satisfies, .a and .ts | Refused at 2:57 | Dynamic-receiver NotYet, test fails |
| c03-as, .ts | Refused at 2:57 | Dynamic-receiver NotYet, test fails |
| c04-non-null, .ts | Refused at 2:57 | Dynamic-receiver NotYet, test fails |
| c09-bare, .a and .ts | Refused at 2:57 | Refused at 2:57 |
| c01-parentheses, .ts | Refused at 2:57 | Refused at 2:57 |
| c08-property, .ts | Refused at 3:48 | Refused at 3:48 |
| c10-setter, .a and .ts | Refused at 3:48 | Refused at 3:48 |
| c00 honest twin, .ts | Lowers | Lowers |

Exactly four subtests go red: `TestLiteralWrappedReceiverFreshnessRejectsStoredReceiver/c02-satisfies.a`, `c02-satisfies.ts`, `c03-as.ts` and `c04-non-null.ts`. With only parentheses skipped, the wrapped function expression is never looked at, the literal stays fresh, and lowering stops later at "stage 0 can't lower a function expression with a this parameter (dynamic receivers are not implemented)", which hides the hole until dynamic receivers lower. The bare and parenthesized rows hold because parentheses are still skipped; c08 and c10 hold because their refusal comes from the contextual ThisType arm, which reads a bare function expression and a setter.

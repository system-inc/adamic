# Binding and design boundaries

## Identifier assignments: 143 unique root sites, zero newly covered

These are not the 190 binary `=` sites. The old binary failure occurred after both operands lowered; `assigning to an Identifier` means `local(target)` found no registered binding. Assigning to declared locals, parameters and captures already works, and the assignment-value order fixture holds those cases to Node.

Replay builder.ts:2169:17 still reproduces this stop. The selected unit previously rolls back the `result` declaration at builder.ts:2133:13 because its checker type is `any`, and subsequently cannot read that binding either. It also fails Map/Set branded Path representations and a structural method call. Replay checker.ts:4521:13 no longer reproduces: its selected statement reaches checker.ts:15893:44, `a value of type __String`, first. The attached JSON preserves ordered findings. These observations establish dependencies for the two examples, not a classification of all 143 sites. No missing binding is fabricated, and no net coverage is claimed.

## Preserved design refusals

Nine NonNullExpression assignment targets, one DeleteExpression statement and one YieldExpression statement are explicitly refused before lowering in `internal/lower/refusals.go`: non-null assertions must be replaced with narrowing or `?? panic`, object shapes are fixed (use a Map for removal), and suspended generator frames require ownership/cancellation rules. The census is measured on a checker-rejected entry-root program and can attempt these refused bodies. Its replays reproduce their old NotYet labels, but that does not authorize changing the accepted language. These eleven sites remain for a language ruling; zero newly covered.

## Boxed null

Mixed logical selection with null still stops at `logical selection requiring a boxed null distinct from undefined`. Both null references and undefined currently use NULL; their boxed identities cannot be distinguished. A separate tagged null box would be a representation choice affecting typeof, strict equality, narrowing and truthiness. No runtime C helper was added and no coverage is claimed for this boundary.

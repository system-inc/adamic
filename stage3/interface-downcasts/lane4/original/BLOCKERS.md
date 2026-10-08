# Remaining original __String admission blockers

The original table has 27 / 540 certified and 3 / 3 remaining. All five nullable
brand fields / nine reads are certified. The fixtures import complete official
050880ce declarations; namespace/private declarations are derived from original
AST, never from cohere. All source fixtures load without checker diagnostics.

| Pair | Original receiver | Field | Candidate reads | Observed refusal |
| --- | --- | --- | ---: | --- |
| 103520:name | ActiveLabel | name | 1 | unsupported never contract |
| 70885:name | { name: __String; oldSymbol: Symbol; } | name | 1 | unsupported never contract |
| 9477:escapedText | LeftHandSideExpression & Identifier | escapedText | 1 | unsupported recursive intersection payload contract at carrier.child |

The first two valid Node controls print word-built. Lowering refuses the helper's
name read before either backend runs. The third valid control prints word-built
on Node, but lowering refuses the carrier child read before reaching escapedText.
Wrong, null, undefined and missing cases retain the same named admission refusal.
These are gap pins, not runtime wrong-value pins and not certified pairs.

Code inspection points to checkLazyViewReads' unsupportedFields map, keyed only
by field name over every descriptor, for the first two failures. A supported
__String read shares the name key with unreachable never descriptors. Scope the
fallback with the existing whole-program allocation graph; retain checks and
refusals for Unknown flows and wider helpers receiving viewed values. Removing
the fallback without preserving these obligations would be unsound.

The third failure is the recursive-intersection descriptor admission guard.
Its complete original ancestors and fields must survive a fix. Do not replace
this receiver with the reduced Identifier fixture from the earlier lane ledger.

Newest integration 322bd65d was inspected and still contains both behaviors.
Its dictionary wiring does not resolve these failures. No individual lane was
merged and no shared guard was disabled. TestCheckedViewBrandsOriginalBlockedPairs
retains eighteen Node controls with exact named compile-refusal suffixes; it must
fail when a blocker closes so its rows can move into the runtime certification
suite with Node, both backend pins and a release-mode mutant.

Working targets: October 9, 2026 at 17:00 MDT for the complete branded family,
conditional on resolving these three blockers; October 13 at 17:00 MDT for mixed
primitives. The other 32 branded pairs / 549 candidate reads are delivered now.
Exact production allocation reachability and whole-tsc compilation are unclaimed.

October 11 lane 7 continuation: original intersected Identifier now lowers with
its complete recursive contract. Its former compile-gap row has moved from
TestCheckedViewBrandsOriginalBlockedPairs to lane 7's
TestCheckedViewIntersectionOriginalIdentifierGap. Original valid controls now
supply complete Node and required Symbol fields and pass all three backends.
The escapedText pair remains uncredited: bypassing the helper read is still
caught by the earlier carrier.child bounded read. A direct intersection cast
remains refused; a wider alias-write isolation attempt is refused as an
unsupported representation conversion at escapedText. See lane7's
REMAINING-REPORT.md for commands and the surviving mutant. The other two
compile-gap rows still pass their original named-refusal pins.

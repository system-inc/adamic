# Temporary adaptation scope

Full-tree adaptations remain under `stage3/adapt/`: 50 (implicit returns),
51 (deferred fallthrough plan, no-op), 60 (factory localSymbol), 61 (factory
 typeExpression), 62 (tracing write), 63 (tracing legend), 64 (parser range read).
These accept the complete upstream tree; 60–64 are the parser worker's edits.

Slice-only adaptations moved here, with names unchanged: 52–59 and 80–86.
52–57, 59, 80–82 and 85 are the active scanner profile. 58, 83 and 84 are
retired; 86 is a plan without an adapter. Each implemented scanner adapter
requires the declaration-slice manifest. No parser adaptation was moved.

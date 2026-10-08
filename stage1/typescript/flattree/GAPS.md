# Flat-tree native blockers

This package contains Go/Node reference tooling, not a native Adamic flat tree.
`gaps/typed_arrays.ts` prints `0` on Node. Stage 0 returns `lower.NotYet`, What
`new an Identifier`, at line 1 column 15. TestNativeBaselineAndTypedArrayGap holds
both observations. No ordinary-array representation is substituted.

Binary mapping, pooled string views and zero-retain borrowing have now been ruled
by @system_adamic. Native work waits for runtime area #4gkdjsz to land. Read
SCOUT.md for the received rulings and exact shared files.

The parser also misparses clean `let a=({b:c=>c});`. TestSharedParserBoundary
records Go's clean parse and the port's false diagnostics. Babel's real source
exposes the same boundary. Two additional recovery probes omit Go diagnostics.
These remain explicitly tested differences, never counted as oracle agreement.
No parser, compiler, runtime, bridge or shared lint harness file is changed.

The follow-up reduces the same missing speculative close-paren guard to the
14-byte clean input `let a=([b=>c])`. Go expected trees, port answers and context
traces are in testdata/parser-parity. The parser remains unchanged.

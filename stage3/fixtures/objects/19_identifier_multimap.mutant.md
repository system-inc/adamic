Mutant evidence for objects fixture 19's stage-0 record in status.json, which this change keeps as it is.

What moved: #9dmh4em refuses a method seen as a writable field of a wider function type (adamic/invariant-mutable). This fixture declared SymbolTracker's trackSymbol as an optional property, `trackSymbol?: (symbol: Symbol, node: Node | undefined, meaning: SymbolFlags) => boolean`, where tsc's types.ts:10048 declares an optional method, `trackSymbol?(symbol: Symbol, enclosingDeclaration: Node | undefined, meaning: SymbolFlags): boolean`. Seen through the property, SymbolTrackerImpl's trackSymbol method could be written (undefined, or another function), so the head refused it at 181:56, ahead of the row's NotYet at 79:82 (Codex 01a1c6b1's delta, at 7725a0bd). The fixture now declares it as tsc does, a method seen as a method, which isn't written, so its first stop and status.json row stay unchanged. Node's output is unchanged.

Mutant: the fixture's line 58 restored to the property declaration, on top of this change; status.json stays.
Command: go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixturesObjects$' -v
Result with the mutant: MEASURED

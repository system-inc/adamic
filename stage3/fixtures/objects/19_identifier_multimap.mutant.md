Mutant evidence for objects fixture 19's stage-0 record in status.json.

What moved: #9dmh4em refuses a method seen as a writable field of a wider function type (adamic/invariant-mutable). This fixture declared SymbolTracker's trackSymbol as an optional property, `trackSymbol?: (symbol: Symbol, node: Node | undefined, meaning: SymbolFlags) => boolean`, where tsc's types.ts:10048 declares an optional method, `trackSymbol?(symbol: Symbol, enclosingDeclaration: Node | undefined, meaning: SymbolFlags): boolean`. Through the property, SymbolTrackerImpl's trackSymbol method could be written, so #9dmh4em refuses it at 181:56. The fixture now declares it as tsc does. Its first stop then moves from 79:82's NotYet to 87:33, `!!this.inner?.trackSymbol` (tsc's SymbolTrackerImpl constructor), a presence test of an optional method through ?., which the method-value rule refuses as unbound-method. The row records that, measured. Node's output is unchanged.

Mutant: the fixture's line 58 restored to the property declaration, on top of this change; status.json stays.
Command: go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixturesObjects$' -v
Result with the mutant (Codex, at 747fde51, against the 79:82 row that change carried): exit 1, red in objects on exactly 19_identifier_multimap.a, at #9dmh4em's own refusal:

    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/objects/19_identifier_multimap.a:181:56: Adamic 0.1 refuses a value of type SymbolTrackerImpl seen as SymbolTracker | undefined, whose method trackSymbol becomes a writable field: a write of ((symbol: Symbol, node: Node | undefined, meaning: number) => boolean) | undefined would replace a method of type (symbol: Symbol, enclosingDeclaration: Node | undefined, meaning: number) => boolean; make trackSymbol readonly in the type it's seen as (readonly trackSymbol: ...), or declare it there as a method, which isn't written (adamic/invariant-mutable) Stdout:}; recorded {...79:82...}

With the fixture as tsc writes it, the same run gave the 87:33 unbound-method stop this row now records.

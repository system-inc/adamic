Mutant evidence for objects fixture 19's stage-0 record in status.json.

What moved: #9dmh4em refuses a method seen as a writable field of a wider function type (adamic/invariant-mutable). This fixture declared SymbolTracker's trackSymbol as an optional property, `trackSymbol?: (symbol: Symbol, node: Node | undefined, meaning: SymbolFlags) => boolean`, where tsc's types.ts:10048 declares an optional method, `trackSymbol?(symbol: Symbol, enclosingDeclaration: Node | undefined, meaning: SymbolFlags): boolean`. Through the property, SymbolTrackerImpl's trackSymbol method could be written, so #9dmh4em refuses it at 181:56. The fixture now declares it as tsc does. Its first stop then moves from 79:82's NotYet to 87:33, `!!this.inner?.trackSymbol` (tsc's SymbolTrackerImpl constructor), a presence test of an optional method through ?., which the method-value rule refuses as unbound-method. The row records that, measured. Node's output is unchanged.

Mutant: the fixture's line 58 restored to the property declaration, on top of this change; status.json stays.
Command: go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixturesObjects$' -v
Result with the mutant (Codex, at 747fde51, against the 79:82 row that change carried): exit 1, red in objects on exactly 19_identifier_multimap.a, at #9dmh4em's own refusal:

    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/objects/19_identifier_multimap.a:181:56: Adamic 0.1 refuses a value of type SymbolTrackerImpl seen as SymbolTracker | undefined, whose method trackSymbol becomes a writable field: a write of ((symbol: Symbol, node: Node | undefined, meaning: number) => boolean) | undefined would replace a method of type (symbol: Symbol, enclosingDeclaration: Node | undefined, meaning: number) => boolean; make trackSymbol readonly in the type it's seen as (readonly trackSymbol: ...), or declare it there as a method, which isn't written (adamic/invariant-mutable) Stdout:}; recorded {...79:82...}

With the fixture as tsc writes it, the same run gave the 87:33 unbound-method stop this row now records.

## #p2n17c4 shape 1: the first stop moves from 87:33 to 181:56

What moved: shape 1 (lower/optional_member_presence.go) lets the refusal pass through a presence test of an
optional member, `!!this.inner?.trackSymbol` at 87:33 among them, and judges it after lowering
(checkOptionalMemberPresence). So 87:33 is no longer a refusal, and the fixture's first stop is the next
refusal, 181:56's adamic/invariant-mutable (SymbolTrackerImpl seen as SymbolTracker | undefined, its
reportInaccessibleThisError becoming a writable field), which runs before lowering's deferred checks.

87:33 is not lowered unjudged. The fixture cut to 87:33 alone (the interfaces, SymbolTrackerImpl with only its
constructor and trackSymbol, and three trackers: a literal with trackSymbol, one without, undefined) stops
at 87:33 as shape 1's NotYet, "a presence test of optional member trackSymbol whose receiver may hold a
value the program didn't build as a plain object literal (no reaching allocation established; untracked
source ir.Defined)", on `adamic c` .a and .ts and on `build --sanitize`, at 8be26a7e (Loom job J-kwwz6vv).
Node prints true:false:false there.

Measured: `go test -count=1 -v -run '^TestFixturesObjects$/^objects$/^19_identifier_multimap\.a$' ./stage3/fixtures`
at 8be26a7e (Loom job J-mqmg9pg) reports `got {Outcome:Refused What:...181:56...invariant-mutable...}` against
the 87:33 row, which this change replaces with the measured text.

Mutant: shape 1's `!l.optionalMemberPresenceTest(node)` removed from methodValueStop's exemption
(internal/lower/refusals.go), with this row as it now is. The same command goes red on exactly this fixture,
got 87:33's unbound-method refusal against the recorded 181:56.

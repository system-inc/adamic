# Mutants for branded arrays held as their base (#gb0fqaa)

Each is one edit, run on a house box, then put back. The presence design is Diagnostics' (Nov 5, on #gb0fqaa).

| Mutant | Edit | Red |
|---|---|---|
| No unbranding | unbranded.go, `unbranded`: return t first | TestBrandMintsOfTscsSortedArraysAreAdmitted, TestBrandMintOfALiteralIsExact, TestUnbrandedArraysAgreeWithNode, TestUnbrandedMintLowersToItsOperand: each stops NotYet at the branded value |
| A raw array question | expression.go, `representation`: `l.checker.IsArrayType(proven)` put back in the Array case | TestArrayQuestionsGoThroughUnbranded |
| A real member admitted | brand_mint.go, `phantomExtension`: `!l.phantomBrandMember(property) ||` made `false ||` | TestUnbrandedOnlyPhantomExtensions, the real-member row lowers; TestBrandMintRecognitionIsExact, its real-member row passes the cast |
| No `in` road | phantom_array_presence.go: the `in` clause made `false &&` | TestUnbrandedPresenceIsRefused, the three `in` rows (b09's union heard as inProperty's union refusal) |
| No Object.hasOwn road | phantom_array_presence.go: the hasOwn case made `false &&` | TestUnbrandedPresenceIsRefused, the Object.hasOwn row |
| No hasOwnProperty road | phantom_array_presence.go: the hasOwnProperty case made `false &&` | TestUnbrandedPresenceIsRefused, the hasOwnProperty row |
| No whole-value road | phantom_array_presence.go: the keys, values, entries, getOwnPropertyNames and assign case made `false &&` | TestUnbrandedPresenceIsRefused, those six rows |
| No JSON.stringify road | phantom_array_presence.go: the stringify case made `false &&` | TestUnbrandedPresenceIsRefused, the JSON.stringify row |
| No object spread road | phantom_array_presence.go: the spread case returns nil | TestUnbrandedPresenceIsRefused, the object spread row |
| No write road | phantom_array_presence.go: the assignment case returns nil | TestUnbrandedPresenceIsRefused, the three write rows |
| Every key refused | phantom_array_presence.go, `phantomPresenceKeyed`: a literal key refused whatever it names | TestUnbrandedPresenceControls, the `"length" in` and hasOwn `"length"` rows |
| Every literal write refused | phantom_array_presence.go, `phantomPresenceWrite`: a literal key refused whatever it names | TestUnbrandedPresenceControls, the write at a number |
| An array spread refused | phantom_array_presence.go: SpreadElement judged as an object spread is | TestUnbrandedPresenceControls, the array spread row |

Not held by a mutant: the oracle fixture of tsc's four shapes on both backends. Adding one means a recorded count row and
the fixture families on a box; the JavaScript backend holds the shapes to Node (TestUnbrandedArraysAgreeWithNode), and
internal/oracle/testdata/notyet_element_access/sorted.a, now lowering, holds a SortedReadonlyArray parameter to Node on
both backends and the leak check.

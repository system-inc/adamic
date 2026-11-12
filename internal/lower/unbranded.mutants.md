# Mutants for branded arrays held as their base (#gb0fqaa)

Each is one edit, run on a house box, then put back. The presence design is Diagnostics' (Nov 5, on #gb0fqaa).

| Mutant | Edit | Red |
|---|---|---|
| No unbranding | unbranded.go, `unbranded`: return t first | TestBrandMintsOfTscsSortedArraysAreAdmitted, TestBrandMintOfALiteralIsExact, TestUnbrandedArraysAgreeWithNode, TestUnbrandedMintLowersToItsOperand: each stops NotYet at the branded value |
| No void guard | unbranded.go, `unbranded`: the voidBrandMember test dropped | TestUnbrandedVoidBrandStops, the length row lowers whole |
| A cast that unbrands | cast_proof.go, `arrayLikeType`: the phantomExtension guard removed (survived every program pin on Home for 1663124c, so pinned directly) | TestUnbrandedCastSeesTheBrand, SortedReadonlyArray reads as an array |
| A raw array question | expression.go, `representation`: `l.checker.IsArrayType(proven)` put back in the Array case | TestArrayQuestionsGoThroughUnbranded |
| A real member admitted | brand_mint.go, `phantomExtension`: `!l.phantomBrandMember(property) ||` made `false ||` | TestUnbrandedOnlyPhantomExtensions, the real-member row lowers; TestBrandMintRecognitionIsExact, its real-member row passes the cast |
| No `in` road | phantom_array_presence.go: the `in` clause made `false &&` | TestUnbrandedPresenceIsRefused, the three `in` rows (b09's union heard as inProperty's union refusal) |
| No Object.hasOwn road | phantom_array_presence.go: the hasOwn case made `false &&` | TestUnbrandedPresenceIsRefused, the Object.hasOwn row |
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
| No own-member rule (Language's must) | phantom_array_presence.go: the `l.holdsArrayBrand(access.Expression) && !l.arrayOwnMember(callee)` case made `false &&` | TestUnbrandedPresenceIsRefused, the hasOwnProperty, propertyIsEnumerable and hasOwnProperty-of-length rows |
| Type parameters unjudged | phantom_array_presence.go, `phantomBrandMatching`: the type parameter case appends the parameter instead of its constraint | TestUnbrandedPresenceIsRefused, the l13 and l14 rows, each heard by a later stop instead of rule 450 |

## Object presence (#3hsrk2t, Nov 12)

Mutant run and restored: the classifier accepts array bases only, dropping both object extensions and branded supertypes. The q06.ts pin loses rule 450 and a separate mutant probe confirms that the whole program lowers. q06.a keeps its older overload-result refusal. The other object observation pins also lose the phantom-presence refusal. The production file was restored through an uncommitted Go overlay.

## Object reads and keyed prototype methods (#3hsrk2t follow-up)

Mutant run and restored through Go overlays: making `phantomObjectRead` return nil loses every read pin. The destructuring probe lowers again; Node prints `undefined` and exits 0, while native ASan/UBSan and JavaScript both exit 70 with `adamic: panic: field read failed: __tagged: mark (field __tagged) is not initialized; expected undefined, found missing`.

Changing the object prototype method back to the whole-value rule turns `TestPhantomObjectPrototypeKeys` red on the real `kind` key. Removing the optional-member and literal-storage exemptions turns `TestPhantomObjectAgreeingReads` red on both controls. All three mutants were run and restored; array presence pins remain unchanged.

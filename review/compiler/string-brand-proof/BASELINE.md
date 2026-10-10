# Main and the pinned alias

At main 55347dce1537c0bd05d35eab0dc5d663c4823803, `phantomParts`,
`phantomBase` and `viewRepresentation` already recognize primitive intersections
whose brand fields have only phantom void types. V2 read contracts use these
proofs. Ordinary `representation` instead sends a string intersection to
`objectIntersection`; that object proof does not authorize a primitive result.
Thus a standalone `string & { __brand: void }` has a view representation but
lacks the ordinary string ABI proof. The three-arm `__String` union likewise
cannot find representations for its primitive intersections, and adding
undefined does not repair that failure. Function results and ordinary values
stop at their respective NotYet sites. Direct string-to-brand assertions also
retain the existing unchecked-cast refusal on main.

The alias in upstream types.ts at 050880ce59e30b356b686bd3144efe24f875ebc8 is
exactly the text saved in types-pin.txt. InternalSymbolName is a const enum of
string literals. The reductions use an ordinary enum with actual upstream
string values to avoid changing the separate const-enum rule. No upstream
implementation was copied into Adamic.

The baseline CLI was built with Go overlays restoring the main versions of
expression.go, refusals.go, cast.go and cast_proof.go. Its optional and value
witness logs independently reproduce the two requested first-error reasons.
The historical full-mode census remains 28 value stops and 16 return stops;
that result is historical evidence, not a delivery rerun.

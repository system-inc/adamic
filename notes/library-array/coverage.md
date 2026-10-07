# Library Array coverage

Base: origin/codex/library-array at 34da3d84c7f3b1860405a8b0bfd545e7ba9b2a43.
Coverage branch: coverage/library-array. Seven added oracle programs; zero output disagreements.
The two refused programs in this directory are compilation-limit reproductions, not disagreements.
No compiler implementation change is retained.

## Case inventory and previous oracle coverage

Paths in the Existing column are in internal/oracle/testdata. Names in the Added column are
library_array_coverage_NAME.a in the same directory. This inventories the behavior changed
by the branch rather than claiming that all pre-existing Array implementations are new.

| Code case | Existing program | Added program or limit |
|---|---|---|
| Intrinsic immediate indexOf.call and lastIndexOf.call | library_array_search.a, object/string/number/boolean receivers | arrays, conversions, effects, shapes |
| Parenthesized method followed immediately by .call | none | arrays |
| Parenthesized Array.isArray call and typeof observation | metadata only unparenthesized | identity |
| Direct array searches indexOf/lastIndexOf/includes with proven object slots | library_array_search.a | arrays; distinct equal-looking object does not match |
| Generic call with an array receiver | none | arrays, conversions |
| Generic array number, string, object, number-or-undefined slots | none | arrays covers all four |
| Generic array empty and readonly receiver | none | arrays |
| Search NaN, signed zero, negatives, fractional and outside starts | library_array_search.a, direct arrays and generic object | arrays adds generic arrays |
| Default last start versus explicit undefined | library_array_search.a, direct array and object | arrays adds generic array; conversions adds optional-number start |
| Missing generic search argument defaults to undefined | none | TypeScript requires 2-3 arguments, so cannot reach this lowering path |
| Generic array search null/mismatched representation | refusal unit tests, no positive oracle | branch refuses distinct null/undefined reference slots or mismatched representation |
| Generic fixed object shape, literal/const/unreassigned let | library_array_search.a | effects and shapes |
| Sparse numeric fields and key equal to length | library_array_search.a | conversions adds fractional-length truncation |
| Numeric fields canonicalized from hex/exponent spellings | library_array_search.a | shapes observes numeric-key object through isArray |
| Nonnumeric field sorting and ignored noncanonical numeric keys | library_array_search.a has NaN field | shapes has 01, 1.0, -1 and excluded 2^53-1 |
| Largest eligible key, length clamped to 2^53-1 | library_array_search.a has 2^32 boundary and Infinity length | shapes finds key 2^53-2 |
| Zero/negative/NaN/Infinity length | library_array_search.a | shapes adds negative fraction and empty-string length |
| Absent, null, undefined, false and fractional length | none | conversions |
| String/boolean length conversion | library_array_search.a | shapes adds empty-string length |
| Number/undefined optional start conversion | none | conversions, array and object receivers |
| String/boolean/null start conversion helper paths | none | TypeScript signature refuses these; refused_bounds.a is a minimal example |
| Nullable/undefined union length fields, optional numeric fields | refusal unit tests cover nullable fields | branch refuses fields whose presence/representation is not fixed |
| Strict primitive equality, different primitive types, NaN | library_array_search.a | shapes adds boolean fields |
| Strict null versus undefined, including present undefined versus absent key | library_array_search.a | shapes, conversions |
| String content equality with owned strings | library_array_search.a | arrays |
| Object identity rather than structural equality | library_array_search.a | arrays |
| Nested array identity in generic object fields | none | shapes |
| String receiver, UTF-16 walking and reverse relative start | library_array_search.a | conversions adds empty string, Infinity reverse start and type mismatch |
| Number/boolean primitive receiver without indexed properties | library_array_search.a | already covered |
| All receiver/search/from arguments evaluated before reading length | library_array_search.a changes length in search argument | effects changes it in both search and from arguments |
| Generic shape view hidden by annotation/reassignment, object ToPrimitive, union searches, empty object fields, hidden field/search representation | library_array_generic tests, not positive oracle programs | intentionally refused by branch |
| Detached, shadowed, optional or non-.call intrinsic search forms; missing receiver/extra args | refusal tests cover detached/shadowed forms | not supported as generic intrinsic calls; cannot add an agreeing positive fixture |
| Array.isArray.name/.length/typeof | library_array_metadata.a | identity |
| Array.isArray empty literal, normal array, primitive/null/undefined, plain object | library_array_metadata.a | identity adds string/object/optional-number arrays and heterogeneous readonly tuple |
| Array.isArray Object/Array/String/Number/JSON/Math/Map/Set and Array.prototype | library_array_metadata.a covers Array/JSON/Math/prototype | identity adds Object/String/Number/Map/Set |
| Array.isArray operand evaluation and ownership | library_array_metadata.a, array-producing function | identity adds owned string and optional-number function operands |
| Array.isArray with shadowed Array binding | none | shapes |
| Array.isArray mixed/optional array or object view, nonliteral object result, spread-created object | optional array and hidden object view refusal unit tests | intentionally refused; refused_is_array_view.a reproduces function-returned object |
| Array.isArray function values and optional boolean/string scalar operands | none | identity |
| Array.isArray unreassigned let bound to plain object | none | identity |
| Array.isArray wrong arity/spread operand | none | one explicit operand is required; no positive program can exercise the refusal as a successful call |
| copyWithin undefined third argument and optional-number end | library_array_copy_within.a | copy_types, effects, identity, walk_chains |
| copyWithin optional end absent/present/NaN/negative/fraction/outside/infinities | existing optional end only undefined/3 | copy_types |
| copyWithin direct end undefined after operand mutates length | library_array_copy_within.a appends in start | effects evaluates receiver, target, start and optional end in order |
| copyWithin number and string slots, overlaps/identity/empty | library_array_copy_within.a | identity chains string copies |
| copyWithin object and number-or-undefined slots, retained element lifetime | none | copy_types, including overlapping copies and keeping objects after removing all slots |
| copyWithin empty object/string/number-or-undefined arrays | empty number array only | copy_types |
| copyWithin nonnumeric target/start or wrong arity | none | refused before emission; no positive oracle program |
| Search slot guard: homogeneous object slot versus primitive hidden in object type | positive object search in library_array_search.a; negative library_array_refused/library_array_heterogeneous.a | arrays exercises homogeneous objects; heterogeneous representation remains refused |
| Slot guard sees literal/spread source across modules and stops after first unsafe item | heterogeneous refusal unit test | unsafe slots cannot produce an agreeing positive program; tagged elements are not implemented |

## Requested callback and chain stress

These callback methods predate the branch, but their interaction with the changed copyWithin
end handling is useful coverage. walk_chains calls copyWithin with undefined inside findLast,
findLastIndex and flatMap callbacks, changes elements, appends while walking, prints visit order,
and chains flatMap/copyWithin/toReversed/join and copyWithin/toReversed/copyWithin/indexOf.
The object/string callbacks use strings allocated at runtime, not only immortal literals.
copy_types chains copyWithin on its returned object and optional-number arrays; identity
chains it on a string array before Array.isArray. Existing library_array_find_last.a already
covers shrinking a string array, early stopping, no match, an empty array, replacing an element
and pushing during reverse walking. Existing library_array_flat_map.a covers shrinking as well.

## Limits encountered

- TypeScript rejects a missing search value and string, boolean or null fromIndex even though
  generic lowering contains conversion/default paths. No unsafe casts or ts-ignore were used.
- A function returning pure undefined is not lowered; identity instead uses number | undefined.
- Array.isArray on a function-returned object or a spread-created object is refused as an object
  view that can hide an array. A plain object literal and an unreassigned plain-literal binding work.
- Assigning an object numeric field with generic[1] is not lowered. effects changes the ordinary
  length property instead. Array element mutation inside callbacks is covered.
- Null/undefined generic receivers require catchable TypeError support; mixed/optional array
  identities and unproven generic object/field/search representations are explicitly refused.
- Searching trillions of absent generic object keys is impractical on Node. The largest-index
  case starts near its present key or has an immediately terminating bound. An initial unbounded
  attempt was stopped and replaced; it is not counted as an output disagreement.

## Mutation proof

In a detached scratch worktree at the branch base, register only the new identity fixture and
change internal/lower/library_array_is_array.go's final BooleanConstant from `Value: isArray`
to `Value: !isArray`. Run that fixture uncached, then restore exactly that line and run it again.
The original coverage checkout is never mutated, so concurrent repository checks use real code.
Actual output and exit statuses are in mutant.log and restored.log beside this report.

The final identity program also covers named and arrow function values, optional boolean/string
operands both absent and present, and an unreassigned let plain-object binding. These last scalar
cases were added after the full gate had completed its oracle package; the final focused oracle
and counts/standalone checks were rerun on the completed program.

Mutation result: test exit 1, caught only by stdout; native and backend both exit 0 with
empty stderr. Restored control: test exit 0. The implementation diff is empty after restoration.

## Final validation

- Seven final fixtures: uncached oracle passed, including Node, JavaScript backend, sanitized
  native, release native and leak checks.
- Each fixture separately built with go run ./cmd/adamic build and executed: exit 0,
  empty stderr, stdout equal to Node. Final identity additions were separately rebuilt.
- Final complete counts update passed (41.003s), adding only seven new rows.
- gofmt -l cmd internal and go vet ./... produced no findings.
- Full uncached repository gate exited 0. Its oracle package passed in 234.507s; final
  scalar identity additions were also validated by the focused oracle and final counts update.
- Mutant killed by stdout with exit 0 and empty stderr; exact restoration passed.
- Setup: go ready 0s, clang ready 0s, node ready 0s, submodules ready 0s, build cache warm 74s;
  done in 74s on 5 processors, cpu.max 400000 100000, 17.6 GB.

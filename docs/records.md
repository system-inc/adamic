# String records

Status: stage-3 lowering, October 7, 2026. The native and JavaScript backends
implement pure mutable string records through the runtime from `754e666`.
The admitted operations are held to Node, with explicit checked stops at the
prototype boundary. This decision supersedes the index-signature policy in
[0.1](0.1.md) for the forms below; fixed-shape objects retain their own rules.

## Evidence and scope

The census at `origin/codex/tsc-census`, `stage3/census/REPORT.md`, section
"String records and post-creation properties", reports 11 index signatures in
7 original compiler files, 5 explicit `Record<string, T>` references in 4 files,
and 68 element-access sites with a string-index receiver. The last number counts
uses, including aliases, not independent dictionary declarations. The pinned
source is TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`.

`stage3/census/repro/r24/main.a` declares `interface Values { [key: string]:
number; }`, initializes `{}`, and prints `String(values)`. Its census observation
is `Refused: an index signature`. The record declaration and allocation now lower, but this exact repro remains
`NotYet: String conversion of an object, array, map or function (ToPrimitive is
not lowered)`. `records_refuse/r24.a` pins that remaining refusal. No original corpus entry
reached lowering in the census; its checker diagnostics precede these operations.

`stage3/census/data/string_lookups.json` contains concrete examples:

- `core.ts:1279`: `hasOwnProperty.call(map, key) ? map[key] : undefined`, on
  `MapLike<T>`, distinguishes own entries from prototype properties.
- `core.ts:1470`: `result[key] ??= []`, on `Record<string, T[]>`, needs one
  evaluation of the receiver and key and keeps the array written back.
- `commandLineParser.ts:4142,4144,4159`: lookup, assignment and deletion on
  `MapLike<WatchDirectoryFlags>`.
- `parser.ts:10734,10741,10794`: writes to pure string-index objects, including
  a value union and a possibly missing array read.
- `commandLineParser.ts` and `utilities.ts`: many `CompilerOptions` lookups,
  which do not establish that a mixed named-property shape is a pure record.

Inference: dictionary lowering addresses a real source form, but it does not
close all 68 sites. Mixed shapes, unchecked presence assumptions, host values,
casts and unrelated syntax need their own decisions and proofs.

## Admitted type forms

A mutable record type has exactly one unrestricted string index signature and
no named properties, methods, call signatures, construct signatures, numeric
index signature or symbol index signature. Recognition uses the resolved checker
type, including inherited members, not the name of an alias or interface.

```typescript
type Values<T> = { [key: string]: T };
type ByName<T> = Record<string, T>;
interface MapLike<T> { [key: string]: T; }
```

These three spellings have the same representation after `T` is instantiated.
`T` must itself have a proven supported representation. Finite key unions such
as `Record<"left" | "right", T>` are fixed shapes, not unrestricted records.
Numeric, symbol, template-pattern, multiple and readonly index signatures remain
`NotYet` in this unit. Readonly views need a separate variance and mutation design.

An index signature beside named properties remains `NotYet`, with a diagnostic
explaining that the dictionary does not yet preserve named-property contracts.
For example, `{ required: number; [key: string]: number }` promises `required`
is present, whereas a dynamic delete could remove it. A narrower named field
also has a stronger value type than an arbitrary index write. Optional fields,
readonly fields and inherited named members do not remove this problem. Do not
erase the fields, silently route them to another store, or accept a mixed shape
through an alias. Conversions between fixed objects and records require a proven
copy or a representation-aware relation; an existing object cannot simply be
reinterpreted as a dictionary.

## Finite partial records

A library `Partial<Record<K, V>>` with a nonempty finite union of string or
number literal keys now uses the same dictionary storage as `MapLike<V>`.
Program aliases, including generic aliases, resolve through the checker to the
library utilities. Every field must be mutable, optional and have the same
non-missing value type `V`; call signatures and additional index signatures are
excluded. Required finite `Record<K, V>` objects retain fixed storage.

The representation is chosen at creation, including a contextually typed `{}`.
Only actual literal entries are defined; the declared optional key set does not
preallocate own properties. A missing key reads undefined and is absent from
`Object.keys`. An explicitly stored undefined is present in keys when `V`
admits it. The finite view and unrestricted view share identity and storage, so
writes and deletes through either alias are visible through the other. Existing
record operations provide Node key order, own enumeration, iteration, spread,
JSON, counted ownership and cycle analysis in both backends.

The checker’s `GetNonMissingTypeOfSymbol` distinguishes the optional absence
marker from an explicit undefined value in `V`. This keeps a partial number
record in number-valued slots rather than changing its MapLike view to optional
number slots. Mutable value invariance remains enforced: partial Dog records
cannot acquire an Animal-valued alias, nor can number-valued records acquire an
alias that can write undefined. Existing fixed objects cannot be reinterpreted
as finite partial dictionaries, and dictionaries cannot acquire a fixed object
view. Empty key sets, symbol keys, readonly shapes and unsupported value slots
remain outside this extension.

The unchanged front24 `native-partial-record-view.a` from parser proof
`5d777de3` is `records_partial_parser.a`. It prints `0` on Node and both
backends. `records_partial_views.a` proves shared writes/deletes and integer-key
order; `records_partial_undefined.a` proves absent versus own undefined,
reference values and a generic alias. The absent-entry IR mutant defines a
missing declared field as own undefined: both backends print `1` where Node
prints `0`, with clean native sanitizer execution. Refusal probes retain
fixed/dictionary storage separation and mutable value invariance.

## Representation and IR

The distinct counted `Record` representation with string keys and a proven
homogeneous value representation is separate from fixed storage. It is neither a
fixed-shape `Object` nor a
JavaScript `Map`. Record expressions carry their instantiated element type;
write sites carry source/type information for invariance and cycle analysis.
The semantic operations below use `RecordCall` with `Method`, ordered
`Arguments`, `Element`, `Returns`, `OwnOnly` and a source `Site`.
`OwnOnly` selects get_own for a proven unobserved inherited result. Value
provenance, borrow inference and cycle reachability remain those of an ordinary get. Creation uses
`RecordLiteral`; lazy `??=` uses `RecordCoalesce`.

| Source | IR operation and meaning |
| --- | --- |
| `{}` or a contextually typed record literal | `RecordLiteral`: allocate a record; define own data entries in source evaluation order |
| `r[k]` | `RecordRead`: own value or missing `undefined`, with the ruled missing-member check below |
| an unobserved read snapshot | `hasOwn` followed by `getOwn` only on a hit; preserve receiver/key evaluation |
| `r[k] = value` | `RecordSet`: assignment semantics, one receiver/key/value evaluation, stores `T` and returns the assigned value where used |
| `delete r[k]` | `RecordDelete`: delete an own entry; JavaScript's delete expression returns true even when absent |
| `k in r` | `RecordHas`: own presence with the same missing-member check, not value truthiness |
| own-property test | `RecordHasOwn`: own presence, including a present undefined value |
| `Object.keys(r)` | `RecordKeys`: a new owned `string[]` of own enumerable keys |
| `Object.values(r)` | `RecordValues`: a new owned `T[]` of present entries, including undefined when `T` permits it |
| `Object.entries(r)` | `RecordEntries`: a new owned `[string, T][]`, preserving each pair and key order |
| `for (const k in r)` | `RecordForIn`: own enumerable string keys for the supported ordinary prototype; mutation semantics must be held to Node |
| `{ ...r, key: value }` into a record | `RecordSpread`: copy own enumerable entries before evaluating following fields, then define them; obey the existing single leading spread restriction |
| `JSON.stringify(r)` | record JSON schema/traversal: own enumerable entries in key order, serialized by the existing JSON rules |

Under `noUncheckedIndexedAccess`, a general `r[k]` is `T | undefined`, not `T`.
Missing numbers and booleans need their optional representations; references use
their missing representation. Presence is independent of the value: an own entry
holding undefined appears in keys, values, entries and `in`. A presence proof must
remain valid across any call or write that could delete or replace the entry.
`??=` evaluates the receiver and key once and its fallback only when missing or
undefined. Tagged optional `??=` values and other compound writes remain `NotYet`. Dot access on a record uses the same
record-key operation rather than a fixed object slot.

JSON omits object entries whose values are undefined or functions; nested values,
non-finite numbers and escaping retain the existing serializer's Node semantics.
A record schema must establish all possible value behavior, including `toJSON`:
an own callable `toJSON` can change serialization. Unsupported callable hooks,
replacers or value representations stay `NotYet`, never silently omitted under
a schema that claims otherwise. A record itself is not iterable by `for...of`.
Runtime iteration is an implementation tool, not a new source-level protocol.

## Key order and prototype boundary

Own array-index keys come first in ascending numeric order: canonical decimal
strings from `"0"` through `"4294967294"`. Other strings follow in first insertion
order. `"01"`, `"-0"`, `"1.0"` and `"4294967295"` are ordinary strings.
Overwriting keeps an existing key's place. Deleting and re-adding an ordinary
string moves it to the end; a re-added array-index key resumes numeric order.
Keys compare as JavaScript strings, including lone surrogates. Numeric keys,
where supported, undergo JavaScript property-key conversion exactly once.

A plain Node `{}` inherits Object.prototype. The ruled Adamic record holds own
keys only. A literal read or `in` key naming any of Node's twelve Object.prototype
members is refused statically and names that member, even when an own entry may
exist. This includes dot reads and string/template literal keys. A dynamic read
calls `adamic_record_get`; dynamic `in` calls `adamic_record_has`. Both return an
own result or ordinary absence, and panic on a missing prototype-member name:

```
adamic: panic: record member 'toString' is missing; records hold own keys only
```

The runtime checks only on a miss. Own prototype-name entries, including own
undefined, work through dynamic keys. `Object.hasOwn`, enumeration, for-in, JSON
and spread use own-only operations and do not invoke this guard. The JavaScript
backend uses a plain object and the same own-first check and exact diagnostic.
Prototype mutation, symbols and inherited values remain outside the representation.
There is no inherited-value tag.

`new`, ordered `keys` and `iterate` return owned references. `get`, `get_own` and
iterator results borrow; retain kept values before advancing or mutating, and
release iterators with `adamic_release`. Writes consume retained key/value counts.
Assignment and definition differ: literal/spread entries use `define`, including
computed `__proto__`. The colon prototype initializer is `NotYet`. `set` stops
on every `__proto__` assignment with the runtime's documented setter limitation,
even when an own data property exists. General primitive conversion is not added.

The JavaScript backend may use a plain object or a Map plus explicit ordering
and prototype behavior; raw Map insertion order is insufficient.

## Inherited-key observations and read use

The per-site evidence is pinned to `7e400a0`,
`stage3/fixtures/records/INHERITED_KEYS.md` on the fixtures branch. Its global
census has 1,041 source operations, including 145 string-capable sites. The
ownership review covers all 36 user-input and 43 unknown sites. These are source
counts, not observed whole-compiler executions. The table does not establish
that every possible valid input avoids an inherited miss.

A pure record read whose result is discarded now preserves evaluation of the
receiver and key, checks `has_own`, and fetches with `get_own` only on an own hit.
This includes a read used directly as a statement and a non-exported const
snapshot with no uses. A const snapshot can also be read before an immediately
following `if (Object.hasOwn(record, key))` when every use of the snapshot is
inside its true branch. Matching uses checker symbol identity, not spelling.
The receiver must be an identifier and the key an identifier or literal; no intervening statement or second
binding is admitted. This proves ownership at the snapshot time and preserves
a reference even if the branch subsequently deletes its entry. An intervening
call, a different key, or an observation outside the branch retains the loud
lookup. A matching own check can also dominate the read itself when the read is in the
first statement of its true branch, reached through a return, a single binding
or the first argument of an identifier/console call. These evaluation paths
contain no intervening call or write. Normal narrowed-value checks remain on
observed own reads. Other control-flow proofs and aliased own-test intrinsics
remain gaps.
Literal prototype names are permitted only when this proof or the primitive
comparison proof below establishes that no inherited value is observed.

Strict equality or inequality between a record read and a present primitive
(number, string or boolean) also uses `has_own` then `get_own`. Every default
Object.prototype value is a function or the prototype object; none can strictly
equal that primitive. A missing own value therefore selects false for `===` or
true for `!==` without fetching an inherited value. Both operands still run in
source order. A left record read is snapshotted before a right operand can delete
or overwrite its entry. Comparisons with undefined, other records, functions or
objects do not acquire this exemption. There is no blanket substitution of
undefined for an observable inherited result.

| Original location | Resolution | Evidence and fixture |
| --- | --- | --- |
| utilities.ts:8159:28, `src[e]` | reads then rejects, scalar bucket | `records_compare_missing_scalar.a` specializes the helper to number records; own and missing constructor, toString, hasOwnProperty and __proto__ print exactly Node's results |
| utilities.ts:8154:45, `src[e]` | input-dependent rejection or real observable inherited value | Original buckets 17 and 19 distinguish nonempty objects (false) from empty arrays (true); the any-valued recursive helper remains refused rather than rewritten to an own miss |
| core.ts:2143:18, `a[key]` | real loud-stop boundary | Neither receiver-presence test proves key ownership; an arbitrary comparer can observe the inherited value. `records_compare_properties_left.a` isolates a missing left and an own right |
| core.ts:2143:26, `b[key]` | real loud-stop boundary | The same argument applies to the right. `records_compare_properties_right.a` isolates an own left and missing right |
| sys.ts:1566:24, `process.env[name]` | real inherited-value boundary; host storage still unsupported | Stock Node reports no own toString and a function on read. `records_environment_boundary.a` checks the analogous plain-record loud stop; it does not claim process.env has a dictionary representation |

No compiler caller of the exported core compareProperties helper was found in
the pinned table; checker.ts has a separate local helper. The environment keys
used by compiler callers exclude the prototype member spellings. Thus these
boundary probes are helper/API observations, not CLI counterexamples. The table
records no demonstrated user-input read-first miss; absence of a demonstration
is not a whole-program proof.

The original comparison bucket fixtures 17, 18 and 19 are retained unchanged in
`internal/oracle/testdata/records_buckets`. The dedicated oracle runs their
unchanged text on stock Node and checks all four keys. These intentional
any/refusal fixtures are excluded from ordinary cohere source linting in
CohereSettings.json, like the existing 0.1 refusal fixtures. For Adamic checking, only
the boolean console arguments are wrapped with String in a scratch source;
the complete helper body remains unchanged and is refused for any. The native
scalar slice is an explicitly typed adaptation, not an untouched upstream build.
All three supported read shapes have guard-removal mutants: restoring a loud
lookup yields exit 70 with the missing-member message where Node prints and
finishes. Sanitized and release native runs, JavaScript, and counted ownership
runs cover the new successful fixtures; the observable boundary cases are
checked stops and do not reach normal-exit leak checks.

## Ownership and soundness

The record owns counts on its stored keys and reference-bearing values, as a Map
does. `set` and `define` consume the key and value: lowering hands over an owned
count, retaining a borrowed input first. Replacing or deleting an entry lets go
of its old owned contents exactly once. The new count must be secured before the
old value is released, including self-assignment through an alias. Freeing the
record releases every remaining entry without recursive-stack growth.

`get_own`, `get` and iterator outputs borrow. A read kept beyond the
immediate operation takes its own count before any subsequent evaluation can
mutate the table. Keys/values/entries arrays and spread copies own their contents,
so they survive deletion and record destruction. Iterators must keep the record
alive and define when borrowed outputs cease to be valid; table reallocation
cannot leave a saved pointer dangling. Loop exit, break, return and exception
cleanup release iterator ownership exactly once.

Record values are mutable slots in the cycle finder, just like Map values.
Reaching follows the record's values, including records nested through objects,
arrays, unions or closures. Every set, define, literal and spread write must be
recorded and judged by the fresh-write proof; an unknown operation stays
conservative. A strong value that can reach its holder cannot bypass the cycle
finder merely because its field name is dynamic. Weak values need their existing
handle representation and cannot be treated as strong references.

Records are invariant mutable containers. `Record<string, Dog>` is not a
`Record<string, Animal>`: the wider alias could store an Animal without `bark`,
then the narrower alias would read it as a Dog. Apply this rule wherever a value
enters a typed slot, including arguments, returns, fields, captures, unions,
assertions and inferred widening. A separately allocated copy can be judged by
its contents; aliases of the same mutable record cannot be widened.

## Acceptance evidence required for lowering

Hold each operation to original source on Node, the JavaScript backend, and native
under ASan, UBSan and the leak check. Use `.a` fixtures with runtime-built keys
and values for ownership probes. Cover numeric-looking key boundaries, overwrite,
delete/re-add, missing reads, present undefined, prototype names, both kinds of
`__proto__` write, evaluation order, spread snapshots, JSON and mutation during
for-in. Use the census snippets above with their adaptations stated explicitly;
do not claim the untouched tsc corpus compiles.

Run a mutant per operation that changes its actual result or ownership, and say
which oracle observation or sanitizer catches it. Independently prove refusal
checks can fail by allowing a mixed shape, mutable widening or cycle-closing
write and running its failing probe. Counts-table changes must be recorded.
Lowering evidence is in `internal/oracle/records_test.go` and `.a` fixtures:
operations, ownership, optional scalars, for-in mutation, adapted census patterns
three checked prototype stops, and stale scalar/reference narrowings. Native sanitized and release builds agree
with Node for supported operations and with the JavaScript backend for ruled
stops. The unadapted r24 String conversion is deliberately a refusal fixture.
`records_census.a` specializes core.ts getProperty/getOwnKeys/groupBy and parser.ts
argMap writes; `Object.hasOwn` replaces the unsupported aliased
`hasOwnProperty.call` intrinsic. The generic mapped groupBy return is omitted.
These are source-pattern fixtures, not an untouched corpus compilation.

The 68 census string-index sites have **zero prototype-member literal key types**,
observed by matching each recorded `key_type` against the runtime's twelve names.
The ledger does not measure dynamic key values or `in` sites. Therefore zero
runtime hits in tsc is an expectation from hasProperty/getProperty, not an
observed whole-program result. No whole-program tsc run is claimed.

Operation mutants alter read, write, delete, membership, own presence, key order, values,
entries, spread, for-in and JSON; all compile and finish sanitizer-clean, and
Node stdout comparison catches each. A JavaScript mutant reads the real inherited
function instead of stopping; the checked exit catches it. Removing releases
preserves output and is caught by LeakSanitizer. Eager coalescing fallback is
caught by Node stdout; removing a stale scalar narrowing check is caught by the
checked stop. Temporary production mutants
for mixed named signatures, readonly signatures, numeric signatures, prototype
literal refusal, storage views, invariance cycle-closing writes, and shallow-spread value widening are caught by
the refusal probes. Their production files are restored after each run.

Unsupported forms include mixed and readonly signatures, fixed/record alias
conversions at any depth, optional record indexing, unsupported value slots (including Weak values and optional booleans),
spreads from fixed objects, more than one leading spread, generic instantiated
record functions outside existing generic support, and JSON values needing
complete object metadata or callable hooks. Cycles use the existing fresh-write
proof and cycle finder, as Map values do. Counts are recorded in counts.md.
Session commands and final gate observations are reported with the commit.
The full repository gate was stopped after over seven minutes in unrelated
bridge port tests; affected packages and an uncached records oracle are the
worker gate. Refusal fixtures live below `records_refuse/`, outside flow's glob
of supported programs. Nullable record types remain `NotYet`; optional record
literals allocate dictionary storage and JSON resolves their defined schema.


## Lowering verification observations

Linux amd64, Go 1.27.1, clang 20.1.8, Node v24.19.0. Setup printed Go 0s,
clang 0s, Node 0s, submodules 0s, build cache 70s, total 70s; `nproc` was 5
with a four-CPU cgroup quota. All test output was redirected to session logs.

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower ./internal/ir ./internal/fresh ./internal/flow ./internal/javascript ./internal/native -count=1 -timeout 10m
# After correcting the Weak test import and moving the refusal fixture:
go test ./internal/lower ./internal/flow -count=1 -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 10m -run 'TestNativeAgreesWithNode/internal/oracle/testdata/records_|TestRecord' -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts
go test ./internal/oracle -run '^TestRecordOperationMutants$/has_own' -count=1 -v
go vet ./...
gofmt -l cmd internal
git diff --check
```

The native package passed in 127.760s and fresh in 45.356s. The corrected
lower/flow run passed in 20.517s / 69.581s. The final uncached records run passed
in 6.832s, with native hits 0 / misses 28 and Node hits 0 / misses 51. It covers
11 admitted or checked programs plus r24's explicit NotYet, sanitizer/release
agreement, leaks on normal completion, and operation mutants. The final count
update passed in 24.046s. Normal record fixture allocations equal frees; deliberate
panics stop while holding live values and are counted at that stop.

There are 24 lowering mutants: eleven native operation mutants caught by Node
stdout, two JavaScript check-removal mutants caught by checked stops, one eager
coalescing mutant caught by Node stdout, one release mutant caught by
LeakSanitizer, and nine temporary production static mutants caught by refusal
assertions. Static mutation logs recorded exit 1 for the intentionally failing
probe; the production files were restored each time. The runtime's own nine
mutants are documented separately below. macOS, other targets, untouched tsc
compilation and dynamic prototype-name frequency in tsc were not tested.

## Finite partial verification, October 8, 2026

Merged main `f4efdd23` into the feature branch with merge commit `dfcbb5d8`.
The expression conflict keeps both main’s enum dispatch and record dispatch;
the counts conflict retains main’s rows and feature-only fixtures. The updated
checker’s named file-path type required explicit string conversion at the two
regexp-library tests. Re-greening found enum namespace objects must bypass the
record index-signature gate; their existing enum representation is retained.

The initial setup failed on those two checker API compile errors. After the
merge compatibility fixes, setup succeeded: Go 0.085s, Node 0.086s, markdown
0.258s, submodules 0.258s, clang 0.818s, build 82.003s, cache warm 82.713s,
total 82.871s; `nproc` 5. Tools are Go 1.27.1, Node v24.19.0 and clang 20.1.8.
The broad fetch attempted historical nested-submodule fetching and was stopped;
setup fetched the pinned submodule revisions successfully instead.

All test output was redirected to `/tmp/partial-record-*.log`. The final
uncached Node oracle covers 73 fixtures, including finite partial records,
existing records, detached own tests, optional/fixed objects and enums. Native
sanitizer/release comparisons and normal-completion leak checks agree with Node.
The absent-entry mutant is killed separately in native and JavaScript by stdout.
The complete counts update covers 515 rows. Three rows are new:

| Fixture | Allocations/frees | Retains/releases | Peak |
| --- | --- | --- | --- |
| records_partial_parser | 5/5 | 0/5 | 5 |
| records_partial_views | 20/20 | 54/44 | 6 |
| records_partial_undefined | 15/15 | 13/23 | 8 |

The other changed rows reflect main’s borrowing improvements applied to the
feature-only record fixtures. Their allocations, frees, peak and region counts
are unchanged; the deliberate environment stop remains a partial execution.
Each retain/release change is recorded here:

| Fixture | Retains before/after | Releases before/after |
| --- | --- | --- |
| detached_own_records | 15/3 | 16/4 |
| detached_own_objects | 12/4 | 16/8 |
| records_discarded | 6/4 | 12/10 |
| records_guarded_snapshot | 21/15 | 33/27 |
| records_compare_missing_scalar | 92/75 | 104/87 |
| records_environment_boundary | 3/1 | 2/2 |
| records_operations | 137/136 | 99/98 |
| records_for_in | 30/28 | 26/24 |
| records_census | 42/38 | 49/45 |

Final commands, all exit 0:

```sh
go test ./internal/lower ./internal/ir ./internal/javascript ./internal/flow ./internal/fresh -count=1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -v -run 'TestPartialRecord|TestNativeAgreesWithNode/internal/oracle/testdata/(records_|detached_own_|optional_|object_|enum)'
go test ./internal/oracle -count=1 -run '^TestCountsAreRecorded$' -args -update-counts
go vet ./internal/lower ./internal/ir ./internal/javascript ./internal/oracle
gofmt -l internal/lower/expression.go internal/lower/records.go internal/lower/records_partial.go internal/lower/records_partial_test.go internal/oracle/records_partial_test.go
git diff --check
```

Lower passed in 94.583s, IR in 29.381s, flow in 140.540s and fresh in
102.812s; JavaScript has no package tests. The uncached oracle passed in
55.811s (native hits 0/misses 193, Node hits 0/misses 152). Counts passed in
85.592s; vet, formatting and whitespace logs are empty.

The full repository gate, untouched tsc parser and non-Linux targets were not
run for this unit. Empty finite key sets and conversion of existing fixed
objects remain unsupported.

## Runtime implementation

# Runtime records

Records are the runtime half of string index signatures and `Record<string, T>`.
The compiler owns lowering, including distinguishing a record view from a fixed
object view. This interface implements own enumerable string data properties.
Records never read inherited values. Symbols and prototype mutation are out of scope.

## Representation and fixed objects

`adamic_record` is currently an alias for `adamic_object`. Its private fixed tuple shape
has one counted slot containing a string-keyed `adamic_map`. The wrapper preserves
record identity independently of table growth. It also lets the existing heap's
iterative freeing release the table and its children, without adding a heap kind
or recursive destructor. The private `0` slot is storage, not an observable property. The iterator's
private slots are `0` (record), `1` (keys) and `key` (cursor), at offsets already
covered by the compiler's global field-layout proof. This avoids introducing new
unchecked field offsets or editing compiler-owned layout logic. Use record operations, never fixed-object reflection, on this wrapper.

Map already provides byte hashing of Adamic's canonical UTF-8/WTF-8 strings,
string equality, open-addressed buckets and entries in insertion order. Numeric
Map hashes, including its optional-number hash, are not used: record keys are
strings. Deletion leaves tombstones; Map's rebuild compacts them while preserving
live insertion order. Reusing Map avoids a second implementation of counting,
overwrite and rehashing. It adds one fixed object allocation per record.

A small literal with a statically fixed set of fields can stay an ordinary
fixed-shape object, with slot caches, when all its uses support that representation.
A record needing computed insertion, deletion or the record interface is allocated
as a table from creation, even when empty or small. There is no size threshold and
no automatic runtime promotion. The compiler must choose one representation for
all aliases of the same object; copying a shared fixed object into a table would
change observable identity and is not an admissible promotion. A general conversion
or mixed fixed-object/record view needs a compiler agreement and is not built here.

## Own-key order

ECMA-262's array indices are canonical decimal strings from `0` through
`4294967294`, inclusive. They precede every other string and are ordered numerically.
`-0`, `-1`, `01`, `1.5`, `1e0`, the empty string and `4294967295` are ordinary string
keys, ordered by first insertion. An overwrite preserves position. Deleting and
reinserting an ordinary key puts it last; an array index returns to its numeric
position. Unicode and long keys use the same rules, with no C-string termination
assumption in lookup or classification.

`keys` scans the live entries, sorts only the array-index keys, then appends the
ordinary keys in insertion order. Classification bounds the length before decimal
arithmetic, so long decimal strings cannot overflow. Lookup, insertion and deletion
reuse Map's expected constant-time hash operations. Enumeration costs O(u + i log i)
and O(n + i) temporary storage, where u includes tombstones, i is live array indices
and n is live keys. It does not sort ordinary strings.

## Counting and C interface

The declarations are in `internal/native/runtime/adamic.h`:

```c
typedef adamic_object adamic_record;
typedef adamic_object adamic_record_iterator;

adamic_record *adamic_record_new(bool reference_values);
adamic_value *adamic_record_get_own(
    const adamic_record *, const adamic_string *key);
adamic_value *adamic_record_get(
    const adamic_record *, const adamic_string *key);
void adamic_record_define(
    adamic_record *, adamic_string *key, adamic_value value);
void adamic_record_set(
    adamic_record *, adamic_string *key, adamic_value value);
bool adamic_record_delete(adamic_record *, const adamic_string *key);
bool adamic_record_has_own(const adamic_record *, const adamic_string *key);
bool adamic_record_has(const adamic_record *, const adamic_string *key);
size_t adamic_record_size(const adamic_record *);
adamic_array *adamic_record_keys(const adamic_record *);
adamic_record_iterator *adamic_record_iterate(adamic_record *);
bool adamic_record_iterator_next(
    adamic_record_iterator *, adamic_string **key, adamic_value *value);
```

`new`, `keys` and `iterate` return one reference the caller owns. Keys are non-null
strings and always counted. Values are homogeneous `adamic_value` slots; the
compiler selects the scalar representation or sets `reference_values` for counted
values, just as with Map. The compiler must continue proving strong cycles absent
or breaking them with Weak; this table does not collect cycles.

Both writes consume their supplied key and value counts. If the caller still needs
either reference, it retains it before the call. On overwrite, Map releases the
newly supplied key, preserves the stored key, releases the old reference value,
and takes the new value. On deletion it releases the stored key and value. On
final release the wrapper releases the Map, which releases its live keys and
reference values through the heap's existing iterative cleanup.

`get_own` returns a borrowed slot or NULL for absence. A present slot containing
undefined is still present; the slot's reference may itself be NULL. Slot pointers
must not survive writes, deletion or growth. `has_own` is presence rather than
truthiness. `delete` returns true even for absence, as JavaScript's delete does,
rather than Map.delete's removed-entry boolean. `size` counts only own live keys.

An iterator is a fixed object with counted record and key-snapshot slots and a
scalar cursor. It holds the record and its ordered keys until released. Each next
step consults the live own table: deleted keys are skipped, overwritten values are
current, keys added after creation are not visited. A snapshot key deleted and
reinserted before its turn resolves to the current own value. Returned keys and
values are borrowed; callers retain references they keep across mutation. Ending a
loop early requires releasing the iterator. Independent and nested iterators each
own their own snapshot. This is an own-key snapshot contract, not Map's live-addition
iterator contract; the V8 for...in mutation probe holds the ordinary case to Node.

## Object.prototype and __proto__

Node's literal objects have Object.prototype, but Adamic's ruled record interface
holds own keys only. `get` returns the own slot, borrowed, or NULL for an ordinary
miss. `has` returns true for an own key and false for an ordinary miss. When a
missing dynamic key names an Object.prototype member, both stop loudly with exit
70 instead of reading through to a prototype. An own data property with such a
name reads normally, including one whose value is undefined. After deleting it,
`get_own` still returns NULL and `has_own` still returns false, while `get` and `has`
stop. Compile-time refusal of literal keys belongs to the compiler.

The member list is taken from `Object.getOwnPropertyNames(Object.prototype)` on
Node v24.19.0: constructor, __defineGetter__, __defineSetter__, hasOwnProperty,
__lookupGetter__, __lookupSetter__, isPrototypeOf, propertyIsEnumerable, toString,
valueOf, __proto__ and toLocaleString. The check dispatches on byte length before
comparing bytes, with at most four candidate comparisons at any length. It uses
neither strlen nor allocation and runs only after an own miss. `has` shares the
same lookup and guard as `get`, so neither scans member names on the hit path.
The diagnostic names the member, for example:

```
adamic: panic: record member 'toString' is missing; records hold own keys only
```

The message is assembled in a fixed stack buffer; recognized members are at most
20 ASCII bytes. No counts are taken for lookup or diagnostics. Enumeration,
iteration and spread continue using own keys and `get_own` and never invoke this
missing-member guard.

`define` models an own data property for literals and spread. It accepts
`__proto__`, which enumerates, overwrites and deletes like any ordinary key.
The compiler already refuses the special prototype-setting literal syntax; a
computed data property and a spread are distinct from that syntax.

`set` models assignment. Every `__proto__` assignment, even if an own shadow is
present, stops with exit 70 and the explicit message:

```
adamic: panic: NotYet: record assignment to __proto__ requires the Object.prototype setter; use an own data property
```

This is an explicit limitation authorized for this interface, not an emulation of
the setter. Node can ignore a primitive assignment, change a prototype for an object
or null, or overwrite an existing own data property. Adamic does none of those
silently. All other keys are stored as by `define`.

There is no inherited-value result or prototype-aware tag. An inherited function
or accessor result must never be read as a homogeneous T slot.

## Consumers and scope

The ordered keys plus `get_own` are the interface for Object.values, Object.entries,
JSON.stringify and spread. Spread into a record uses `define`, retains transferred
references, and therefore creates a data property for `__proto__`. A fixed-object
target needs a compiler-proven shape and appropriate slots; arbitrary keys cannot
be silently dropped. Reflection, JSON schemas and record/fixed-object interoperation
are not wired into the compiler or generic fixed-object APIs in this unit.

The harness compares Object.keys, Object.values, serialized number-valued objects,
for...in key/value pairs and record spread against Node. Its JSON adapter constructs
an ordered schema for the existing JSON writer; it proves the consumer interface,
not end-to-end record lowering or generic dynamic-record JSON support.

## Verification

`internal/native/record_test.go` builds the C fixture in
`internal/native/testdata/records/harness.c` through the same Build path as heap tests.
`oracle.js` uses Node's own operations as the independent oracle. Every successful
fixture runs with Count, ASan, UBSan and LeakSanitizer enabled, compares stdout byte
for byte, and requires allocations equal frees. The deliberate missing-member
stops and NotYet panic are checked for exact message and exit; like other panics,
they do not reach normal-exit leak checking.

Fixtures include empty records, the index boundary and noncanonical numbers,
delete/reinsert and overwrite, every prototype name, empty and 4096-byte keys,
non-ASCII keys, 100,000 numeric keys inserted in reverse order then deleted and
reinserted, one million ordinary keys with hit/miss lookup and deletion of half,
snapshot deletion and overwrite, dynamic read hits and ordinary misses, own
prototype-member reads (including undefined), references retained past record release, and
counted key/value overwrite and compaction churn, present undefined, scalar booleans,
and iterative freeing of a 100,000-record chain. Mutation during iteration is
covered for deletion, addition and overwrite, not every engine-permitted for...in
mutation combination. Symbols and exotic property descriptors are out of scope.

Isolated production-runtime mutants prove the comparisons and memory checks fail:
integer keys left in insertion order, UINT32_MAX treated as an array index, and a
deleted key still yielded are caught by Node comparisons without sanitizer failures.
Removing Map's release of the overwrite key is caught by LeakSanitizer; freeing a
stored key is caught by ASan; a null own-slot access is caught by UBSan. Mutated
libraries are built in private test caches, without changing the working tree.

The missing-member fixtures derive their list from Node at test time and check
both `get` and `has` for every name, so a drift in Node's list cannot silently
omit a guard. The diagnostic text and exit 70 must match, followed only by the
count report. Near-matches in every recognized length group are ordinary misses.
A mutant restoring inherited membership (`has` returns true for a missing
prototype name) and a mutant silently returning NULL from `get` are caught by the
exact stop checks, with no sanitizer failure and with balanced counts. Moving
the guard onto the hit path is caught by an own `toString` read that must succeed.

`ADAMIC_RECORD_BENCH=1 go test ./internal/native -run '^TestRecordBenchmark$' -count=1 -v`
runs five alternating fresh native/Node process pairs at 1,000, 10,000, 100,000 and
1,000,000 keys. Both build k-prefixed keys, look up every hit and a separate set of
misses through the dynamic `get` API, delete even-numbered keys, then enumerate
remaining keys and sum values.
Formatting lookup keys is included on both sides. Iteration includes native key
snapshot creation and V8's for...in enumeration. Native uses -O2 with the ordinary
allocator and no counting or sanitizers. Timings exclude process startup and final
record release; checksum, visited count and size must agree. Each operation's best
of five is reported independently. These are observations on a shared worker,
not a portable speed guarantee.

### Initial table observations for f4a0ecc, October 7, 2026

Linux amd64, clang 20.1.8, Node v24.19.0; `nproc` 5 with a cgroup quota of
4 CPUs. Setup completed in 97s (Go 0s, clang 0s, Node 1s, submodules 1s,
build cache warm 97s). These initial measurements used `get_own`.
Best-of-five milliseconds, native / Node:

| Keys | Insert | Hit | Miss | Delete half | Iterate |
|---:|---:|---:|---:|---:|---:|
| 1,000 | 0.102 / 0.376 | 0.066 / 0.088 | 0.062 / 0.363 | 0.036 / 0.057 | 0.017 / 0.065 |
| 10,000 | 1.235 / 3.535 | 0.688 / 1.540 | 0.624 / 2.725 | 0.387 / 0.394 | 0.171 / 0.711 |
| 100,000 | 13.246 / 42.140 | 8.796 / 15.053 | 9.517 / 34.706 | 5.192 / 12.114 | 2.566 / 8.318 |
| 1,000,000 | 262.425 / 680.951 | 219.440 / 411.281 | 292.286 / 637.302 | 133.981 / 250.839 | 85.971 / 169.375 |

The million-key run produced the same checksum, 749999500000, with 1,000,000
misses and 500,000 keys remaining and visited. The counted million-key correctness
fixture made and freed 3,500,004 heap values. Final gate details and mutant results
are reported with the commit; raw session logs are `/tmp/runtime-records-*.log`.

Final checks, all exit 0, with output redirected to session log files:

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/native -count=1 -timeout 10m
go test ./internal/native -run '^TestRecordsAgainstNode$' -count=1 -timeout 10m -v
go test ./internal/native -run '^TestRuntimeFieldLayoutsAreIncluded$' -count=1 -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|objects|library_object_order|library_object_own|library_map|map_)' -count=1 -timeout 10m -v
ADAMIC_RECORD_BENCH=1 go test ./internal/native -run '^TestRecordBenchmark$' -count=1 -timeout 10m -v
go vet ./internal/native
gofmt -l internal/native/record_test.go
git diff --check
```

The complete native package passed in 88.159s, including all six mutants. The
record-only counted sanitizer comparison passed in 11.639s; its ownership fixture
made and freed 350,024 values, including the deep chain. The layout check included
25 runtime layouts. The uncached oracle passed 18 selected fixtures in 11.471s
(native cache hits 0, misses 53; Node hits 0, misses 36). The final benchmark passed
in 18.880s. Vet, formatting and whitespace logs were empty. The full repository
gate and end-to-end index-signature lowering were not run. Linux was tested;
macOS and other targets were not.

The first full native run correctly rejected the new private field name `table`
as absent from the global layout proof. Reusing the existing tuple offsets fixed
that integration failure without weakening the check or editing compiler files.
The first oracle filter matched no subtests; the full-path filter above was then
run and its actual fixture results checked.

### Own-only read observations, October 7, 2026

Member names were read from Node v24.19.0 itself with:

```sh
node -e 'console.log(process.version); for (const name of Object.getOwnPropertyNames(Object.prototype)) console.log(name.length, name)'
```

The new dynamic-read fixture made and freed 27 heap values. All successful
fixtures remained counted, ASan/UBSan/LeakSanitizer clean and byte-for-byte equal
to Node for their own operations. Both dynamic operations stopped with the exact
member-specific diagnostic and exit 70 for every Node-derived name (24 cases).
Those intentional stops do not reach normal-exit leak checking. The own-only
helpers, ordered keys, iteration and spread remain unchanged in the runtime.

Final commands, all exit 0, with output redirected to `/tmp/runtime-records-read-*.log`:

```sh
go test ./internal/native -run '^TestRecordsAgainstNode$|^TestRecordReadMutants$' -count=1 -timeout 10m -v
go test ./internal/native -count=1 -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|objects|library_object_order|library_object_own|library_map|map_)' -count=1 -timeout 10m -v
ADAMIC_RECORD_BENCH=1 go test ./internal/native -run '^TestRecordBenchmark$' -count=1 -timeout 10m -v
go vet ./internal/native
gofmt -l internal/native/record_test.go
git diff --check
```

The targeted counted fixture and new mutant run passed in 14.490s; the native
package passed in 88.954s, including all nine mutants. The uncached regression
oracle passed 18 fixtures in 8.316s (native cache hits 0, misses 53; Node hits 0,
misses 36). Static-check logs were empty. The benchmark now calls `get`, including
its miss-only member guard, and passed in 18.655s. Best-of-five milliseconds,
native / Node, on the same toolchain and worker as above:

| Keys | Insert | Hit | Miss | Delete half | Iterate |
|---:|---:|---:|---:|---:|---:|
| 1,000 | 0.119 / 0.385 | 0.067 / 0.092 | 0.063 / 0.348 | 0.036 / 0.056 | 0.017 / 0.064 |
| 10,000 | 1.273 / 3.504 | 0.716 / 1.496 | 0.666 / 2.538 | 0.382 / 0.409 | 0.174 / 0.719 |
| 100,000 | 14.880 / 45.369 | 9.959 / 18.137 | 12.457 / 40.397 | 5.615 / 17.118 | 2.969 / 8.656 |
| 1,000,000 | 263.827 / 632.294 | 234.505 / 387.938 | 298.775 / 603.479 | 130.902 / 247.342 | 80.718 / 157.407 |

The three new mutants restored inherited `in` membership (caught by the exact
stop check), returned NULL silently for a missing member read (the same check),
and applied the guard to an own `toString` hit (caught by the successful own-hit
fixture). The six original ordering and memory mutants still passed their
detection checks. No compiler files were edited; literal-key refusal and
end-to-end lowering are the compiler's work. The full repository gate and
non-Linux targets were not run for this change.

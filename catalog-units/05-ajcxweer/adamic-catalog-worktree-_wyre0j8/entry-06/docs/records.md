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

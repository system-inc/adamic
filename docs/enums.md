# Enums in Adamic 0.2

The current decision is [Numeric enums are open](#numeric-enums-are-open).
It supersedes the numeric closed-domain and flag-domain decisions recorded below.
The earlier decisions and their verification records are retained as history.
The [literal-tag follow-up](#literal-tag-narrowing-follow-up-october-7) records
the October 7 narrowing implementation and its remaining checked-view
boundaries.

Decision for Kirk, October 6, 2026: admit constant-valued numeric and string
enums, const enums, and enum member types. Keep enum values a closed union of
that declaration's members. Disable `erasableSyntaxOnly` in both
`internal/load/load.go` and `tsconfig.json`; explicit lowering refusals continue
to exclude namespaces and parameter properties. This supersedes the enum entry
in the approved [0.1 refusal table](0.1.md), without opening the other syntax in
that entry.

The [survey's syntax census](https://github.com/system-inc/adamic/blob/codex/tsc-strictness-survey/docs/tsc-strictness-next-layer.md)
counts 164 enum declarations: 139 const and 25 ordinary, with 157 entirely
numeric and seven entirely string-valued. It found no nonconstant or mixed
families. These are declaration counts. Supporting these value families does
not establish that the TypeScript compiler now compiles: its number assignments,
bitmask combinations, namespaces and other contracts still need adaptation.

## Behavior

* Numeric members support implicit numbering, references, aliases, arithmetic
  constant expressions and bitwise flags, using the checker's constant evaluator.
* Ordinary enums have a runtime object. Numeric members add forward and reverse
  properties; string members add only forward properties. The last numeric alias
  wins the reverse lookup. Key enumeration follows JavaScript's integer-key order.
* `Enum[value]` evaluates the receiver and key once, preserves key effects, and
  returns `undefined` for a missing numeric reverse key. Dynamic forward lookup
  accepts proven member names. Numeric key spelling matches JavaScript, including
  fractions, negative zero and exponent notation.
* `Enum.Member` remains an enum member type in parameters, returns, fields,
  arrays, aliases and generic instantiations. Its runtime representation is the
  corresponding number or string.
* Const enum members inline, including imported members. No const enum object
  or binding is allocated. Runtime uses of the const enum object are rejected
  by TypeScript's checker. Use an ordinary enum for an observable runtime object.
* A switch on a numeric or string enum must cover its distinct member values or
  have a default. Numeric aliases need only one case. Narrowed member types need
  only their possible values; void switches are checked too.

Adamic compiles a closed, typed program. It does not implement
`preserveConstEnums` for external, untyped JavaScript consumers. Upstream
TypeScript's use of that option therefore remains a porting obligation rather
than an assumed equivalence. The source oracle uses Node 24's independent
`stripTypeScriptTypes` transform mode, which understands enum syntax; it never
uses Adamic's lowering to construct the reference output.

## Why the types remain true

TypeScript's union-enum representation alone is insufficient: its checker still
accepts a `number` variable as a numeric enum, and can relate separate enums with
matching names and values. The negative tests deliberately load those programs
successfully before requiring Adamic to refuse them.

Adamic requires both declaration identity and member compatibility. An arbitrary
number, a bare numeric literal, an arithmetic result, or a member from another
enum cannot be stored as this enum. The diagnostic explains the closed union and
asks for a declared member or explicit validation returning a member:

```ts
enum Permission { None, Read = 1, Write = 2, Both = Read | Write }

function permission(value: number): Permission | undefined {
    if (value === Permission.None) { return Permission.None; }
    if (value === Permission.Read) { return Permission.Read; }
    if (value === Permission.Write) { return Permission.Write; }
    if (value === Permission.Both) { return Permission.Both; }
    return undefined;
}
```

Use `Permission.Both` for a stored enum mask. An arbitrary bitwise combination
remains a number until validated. Increment, decrement and compound arithmetic
cannot write an enum slot either.

The same relation is applied inside containers, properties and function views,
with both directions required for writable slots. Thus `E[]` cannot become
`number[]` and receive an arbitrary number through that alias. Method overrides,
`Object.assign`, readonly-to-writable views and invariant generic class arguments
cannot bypass the member rule. Readonly views of enum values as numbers remain
safe. Enums retain declaration identity across module imports.

A `typeof E` view must name the actual enum object or an alias of it. Structural
copies can omit numeric reverse properties or hide heterogeneous fields that
would invalidate reflection. This requirement also holds inside containers.
Use an explicit interface when an ordinary object is intended.

Ordinary enum objects contain only primitive member values and immortal string
constants. They use the existing object representation, ownership and cleanup.
Lookup helpers use ordinary typed IR, so flow, borrowing, regions and Perceus
reuse see their arguments and reads; no analysis is disabled for an enum program.

## Explicit limits

Computed nonconstant members, non-finite member values, ambient definitions,
merged declarations, and enums inside functions or blocks are `NotYet`.
Mixed numeric/string objects work, but mixed-representation switches remain
`NotYet`. The unit does not compile the complete TypeScript compiler.

Reads before a regular enum's runtime initialization are `NotYet`. JavaScript
hoists its enum variable, so such a read can fail differently from a lexical
TDZ. The preflight follows functions, defaults and immutable aliases and checks
callbacks conservatively; an indirect call or class construction before enum
initialization may also be refused. Declare enums before executable module code.

Member names containing NUL or `__proto__` are refused. Numeric members named
`NaN`, `Infinity` or `-Infinity` are refused too: TypeScript permits these names,
but numeric indexing could then return a number where it promises a string.
Rename the member. Numeric enum reverse indexes are readonly in the checker.
Numeric enum `Object.values`/`entries` have heterogeneous runtime values and
remain refused by the existing homogeneous reflection contract.

## Validation and mutants

The source fixtures are `enums.a`, `enums_const.a`, and
`enums_modules/main.a` under `internal/oracle/testdata`. They cover ordinary and
const numeric/string members, aliases, constant expressions, member types,
imports, narrowing, forward and reverse lookup, side effects, optional member
reads and key order.
Each runs against source Node, generated JavaScript, native release,
ASan/UBSan and the leak check.

Toolchain setup succeeded: Go 0s, clang 1s, Node 1s, submodules 1s,
build-cache warm 79s, total 79s; `nproc` reported 5, with a four-CPU cgroup quota.

The six semantic mutants below compile and exit 0 without a sanitizer finding.
Only independent Node stdout comparison catches them. The cleanup mutant keeps
Node stdout unchanged and is caught by LeakSanitizer.

| Mutant | Catch |
| --- | --- |
| Wrong numeric forward value | Node stdout |
| Wrong string forward value | Node stdout |
| First numeric alias wins the reverse map | Node stdout |
| Add a reverse property for a string member | Node key enumeration |
| Wrong inlined const member | Node stdout |
| Skip the numeric lookup key's effects | Node stdout |
| Skip the enum object's final release | LeakSanitizer |
| Permit number-to-enum assignments | `TestEnumsRefuseOpenNumericValues/variable` |
| Ignore enum identity across modules | `TestEnumIdentityAcrossModules` |
| Permit arithmetic updates | `TestEnumsRefuseOpenNumericValues/increment` |
| Skip enum switch coverage | `TestEnumSwitchExhaustiveness` |
| Skip enum initialization analysis | `TestEnumLimitsStayLoud` |
| Permit parameter properties after opening the syntax option | `TestEnumLimitsStayLoud` |
| Permit merged declarations | `TestEnumLimitsStayLoud` |
| Materialize an ambient enum | `TestEnumLimitsStayLoud`, explicit constant initializer |
| Permit prototype-changing names | `TestEnumLimitsStayLoud` |
| Permit numeric members named NaN/Infinity | `TestEnumLimitsStayLoud` |
| Allocate an erased const enum | `TestConstEnumErasesRuntimeObject` |
| Ignore enum domains in generic class arguments | `TestEnumsRefuseOpenNumericValues/generic_class_invariant` |
| Ignore enum domains in Object.assign | `TestEnumsRefuseOpenNumericValues/object_assign` |
| Permit non-finite values | `TestEnumLimitsStayLoud` |
| Accept a fresh structural enum object | `TestEnumsRefuseOpenNumericValues/structural_numeric_enum_object` |
| Accept structural enum objects inside arrays | `TestEnumsRefuseOpenNumericValues/structural_enum_object_array` |

The first ambient mutant was masked by the unresolved-value check on an
uninitialized ambient member. Adding an explicitly initialized ambient member
made that mutant fail. No build failure is counted as a killed mutant.

Reproduce the permanent semantic and ownership mutants with:

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/oracle -run 'TestEnumSemanticMutants|TestEnumCleanupMutant' -count=1 -v > /tmp/enums-mutants.log 2>&1
go test ./internal/lower -run 'TestEnum|TestConstEnum' -count=1 -v > /tmp/enums-types.log 2>&1
```

Compiler guard mutants were run by temporarily removing each listed guard,
running its named regression test, and restoring the source after each run.

An additional mutant disabling the literal-key routing guard was not killed:
the checker did not resolve an enum-member symbol for that computed key, so
lookup still evaluated it. That masked mutant is not included in the 23 killed
mutants above.


## Integration changes

New lowering and oracle files contain the implementation, negative probes,
fixtures and permanent semantic/cleanup mutants. The existing-file edits are:

| File | Reason |
| --- | --- |
| `internal/load/load.go` | Admit enum syntax by disabling the checker option. |
| `internal/load/load_test.go` | Remove the superseded enum syntax rejection expectation. |
| `tsconfig.json` | Match the compiler's language option. |
| `oracle/node.mjs` | Let independent Node transform enum syntax. |
| `internal/lower/refusals.go` | Replace blanket enum rejection with precise checks. |
| `internal/lower/lower.go` | Declare runtime bindings and invoke enum lowering/preflight. |
| `internal/lower/expression.go` | Route enum reads to their lowering. |
| `internal/lower/object.go` | Accept enum switch labels and check exhaustiveness. |
| `internal/lower/invariance.go` | Preserve closed enum domains through views and writes. |
| `internal/lower/class_inheritance.go` | Preserve enum domains in invariant generic class arguments. |
| `internal/lower/library_object.go` | Check reflection shapes and Object.assign writes. |
| `internal/oracle/counts.md` | Record the three new fixtures. |

No native emitter, runtime or analysis visitor changes are needed: enums use
existing typed objects, primitive values and calls.


## Gate and counts

The full gate passed with `go test -count=1 -timeout 30m ./...`, including
stage 1 cohere and TypeScript parser/scanner packages. The slowest package was
`internal/unicodeproperties` at 704.463s. Final shape and optional-read fixes
were followed by fresh tests of all affected packages and the oracle package.
`go vet ./...`, `gofmt -l cmd internal` and `git diff --check` are clean.
All test output was written to log files.

The counts table grows from 262 to 265 rows. Every pre-existing row is unchanged;
allocations equal frees for each new fixture:

| Fixture | Allocations | Frees | Retains | Releases | Peak | In regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| enums.a | 101 | 101 | 79 | 147 | 16 | 0 |
| enums_const.a | 11 | 11 | 6 | 18 | 3 | 0 |
| enums_modules/main.a | 24 | 24 | 20 | 31 | 12 | 0 |

Reproduce the final checks:

```sh
go vet ./... > /tmp/enums-vet.log 2>&1
go test -count=1 -timeout 30m ./... > /tmp/enums-gate.log 2>&1
go test -count=1 -timeout 30m ./internal/lower ./internal/load ./internal/native ./internal/flow ./internal/fresh ./internal/oracle > /tmp/enums-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 > /tmp/enums-counts.log 2>&1
gofmt -l cmd internal > /tmp/enums-format.log
git diff --check > /tmp/enums-diff.log
```


Recorded final command output:

```text
ok  internal/lower   30.065s
ok  internal/load     1.958s
ok  internal/native 161.801s
ok  internal/flow   100.816s
ok  internal/fresh   46.614s
ok  internal/oracle 156.175s
```

The separate final permanent-mutant run passed in 2.895s. Counts regeneration
passed in 22.854s; the final whole-oracle run also verified the recorded table.
Vet, formatting and whitespace logs were empty. The compiler guard mutation
runs recorded `CAUGHT` for all 16 guard mutants, after restoring each edit.
Together with the seven runtime mutants, this gives 23 killed mutants.

## Flag enums

Decision by @system_adamic: opt in by initializer shape, never by the evaluated
values. Every member must have an explicit initializer: `0`, `1 << n` with a
numeric literal `n` from 0 to 30, or an `|` expression whose leaves are members
of this same enum. Parentheses do not change the shape. Implicit `0, 1, 2`,
bare nonzero literals and arithmetic initializers keep an enum closed and
switch-exhaustive under the preceding rules. The source-fixture follow-up below
also admits direct same-enum aliases, whose values cannot add bits.

```ts
enum Flags { None = 0, A = 1 << 0, B = 1 << 1, High = 1 << 30 }
const combined: Flags = Flags.A | Flags.B;
const masked: Flags = combined & ~Flags.A;
const toggled: Flags = masked ^ Flags.High;
```

A flag enum's domain is the non-negative int32 values whose bits are a subset
of the union of its declared member bits. Values retain declaration identity.
`x & y` has a domain proof when either operand has that enum's proof; `x | y`
and `x ^ y` require both. This admits `flags & ~Flags.A` without admitting
`~Flags.A` alone. A bare number gains no enum identity just because its bits
fit. Complement, arithmetic, shifts, increment and decrement leave the domain.
Their results belong in a `number`, and cannot be stored into a flag slot.
Bitwise compound assignments follow the same proofs. Compound arithmetic and
compound shifts are refused. Immutable aliases preserve an expression's proof;
an inferred mutable number variable does not.

The same writes are checked in fields, arrays, Map values, arguments and returns,
including imports and const enums. Writable views retain the preceding invariant
container checks. An enum member type still denotes that member, so a combination
cannot be stored as `Flags.A`. Excluding other members does not prove that a flag
is the last remaining member: it could be a combination. Such a narrowed full
flag slot is conservatively refused when stored into a member-specific slot;
return the explicit member after validation instead.

Every switch on a flag enum requires a `default`, even if it lists every declared
member. Combinations and zero are values too. Reverse lookup of an undeclared
combination returns `undefined`, with type `string | undefined`.
A closed non-flag enum switch that covers all its possible values may use
`default: { const unreachable: never = value; ... }`. Its proven unreachable
default body is omitted locally during switch lowering; flag defaults are kept.

These programs are refused, with a rule and a fix in the diagnostic:

| Program | Reason |
| --- | --- |
| `const value: Flags = ~Flags.A` | Complement has no flag domain proof. |
| `const value: Flags = Flags.A + 1` | Arithmetic produces a number. |
| `flags++` | Increment writes an unproven number into a flag slot. |
| `flags <<= 1` | A compound shift leaves the domain. |
| `enum Bad { A = 1 << 31 }` | The sign bit exceeds the non-negative int32 bound. |
| `enum Color { Red, Green, Blue }; const c: Color = Color.Red | Color.Green` | Implicit initializers do not opt into flags. |
| A flag switch without `default` | Declared cases do not cover the flag domain. |
| `const value: Flags = Flags.A | Other.B` | OR requires two operands from this enum. |
| `const value: Other = Flags.A | Other.B` | Declaration identity also excludes the other target. |

The new fixtures are `enums_flags.a`, `enums_flags_modules/main.a` and
`enums_flags_never_default.a`. They cover combinations, masking in both operand
orders, bit tests, fields, arrays, Map values, function parameters and returns,
ordinary and const enums across modules, the high bit at position 30, reverse
lookup of combinations, default switches and the never-default idiom. Node is
the independent source oracle for stdout and exit code; both generated backends,
native release, ASan/UBSan and LeakSanitizer are checked.

One guard mutant per rule was run against a named lowering test and restored.
All 16 below were caught by assertion failures; no Go or clang build failure
counts as a caught mutant. The mutation runner initially misparsed the trailing
`$` in the never-default test filter. After fixing that parser, every mutant was
rerun and all 16 were reported caught in the fresh run.

| Mutant | Named catch |
| --- | --- |
| Classify by numeric values instead of initializer shape | `TestFlagEnumsRefused/implicit` |
| Require both AND operands to be flags | `TestFlagEnumsDomain/and_left` |
| Admit OR with only one flag operand | `TestFlagEnumsRefused/or_number` |
| Admit XOR with only one flag operand | `TestFlagEnumsRefused/xor_number` |
| Preserve the operand's domain through complement | `TestFlagEnumsRefused/complement` |
| Give addition a flag proof | `TestFlagEnumsRefused/arithmetic` |
| Give shifts a flag proof | `TestFlagEnumsRefused/shift_result` |
| Permit increment of a flag slot | `TestFlagEnumsRefused/increment` |
| Permit compound shift of a flag slot | `TestFlagEnumsRefused/shift_update` |
| Drop the flag switch default requirement | `TestFlagEnumsRefused/switch_default` |
| Drop the non-negative int32 initializer bound | `TestFlagEnumsRefused/sign_bit` |
| Ignore operand enum identity | `TestFlagEnumsRefused/cross_enum_left` and `cross_enum_right` |
| Trust inferred mutable number aliases | `TestFlagEnumsRefused/mutable_alias` |
| Trust exclusion narrowing into a member slot | `TestFlagEnumsRefused/narrowed_member` |
| Permit bitwise compound updates with an open number | `TestFlagEnumsRefused/compound_or_number` |
| Lower a proven unreachable never-default body | `TestEnumNeverDefault` |

Setup reported Go, clang, Node and submodules ready at 1s, build cache warm at
122s, and total 122s. `nproc` reported 5; the cgroup quota was four CPUs.
The new fixture counts are:

| Fixture | Allocations | Frees | Retains | Releases | Peak | In regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| enums_flags.a | 44 | 44 | 23 | 64 | 11 | 0 |
| enums_flags_never_default.a | 2 | 2 | 6 | 6 | 2 | 0 |
| enums_flags_modules/main.a | 9 | 9 | 6 | 12 | 5 | 0 |

Final restored-source validation passed with the commands below. Lowering took
39.028s, the uncached three-fixture oracle took 2.492s, and counts verification
took 17.832s. Vet, formatting and whitespace logs were empty. This unit ran the
full touched lowering package and a filtered oracle, not the full repository
gate or a compilation of TypeScript's compiler.

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower -count=1 -timeout 30m > /tmp/flag-enums-final-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/enums_flags' -count=1 -v -timeout 30m > /tmp/flag-enums-final-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m > /tmp/flag-enums-final-counts.log 2>&1
go vet ./... > /tmp/flag-enums-final-vet.log 2>&1
gofmt -l cmd internal > /tmp/flag-enums-final-format.log
git diff --check > /tmp/flag-enums-final-diff.log
```


## Integration with current main

Merged `origin/main` at `50045bd` with a merge commit. The only conflict was
`internal/lower/lower.go`: the enum branch modified code that main extracted.
The resolution preserves main's ownership map: initialization preflight stays
in `lower.go`, module enum registration moves to `modules.go`, and enum statement
dispatch moves to `statements.go`. The native split is unchanged from main.

Validation passed: `go vet ./...`; `gofmt -l cmd internal` (empty output);
`go test ./internal/lower ./internal/load -count=1 -timeout 30m` (52.204s and
2.586s); the six enum Node fixtures with `ADAMIC_GATE_UNCACHED=1`, `-count=1`
(65.508s, zero cache hits); and `TestCountsAreRecorded -count=1` (101.364s).
Logs are `/tmp/flag-enums-merge-{vet,format,packages,oracle,counts}.log`.
Setup reported Go ready at 0s, clang, Node and submodules at 1s, build cache warm
and total at 152s. `nproc` is 5 and the CPU quota is four CPUs.

Counts verification passed without regeneration. Main supplied 12 new rows and
changed the 20 existing rows below; the merge introduces no further count drift.
Only retains and releases changed. All fixture paths are under
`internal/oracle/testdata/` except the explicitly named load fixture.

| Fixture | Retains before | Retains after | Releases before | Releases after |
| --- | ---: | ---: | ---: | ---: |
| class_oct6_subclass_holder.a | 49 | 55 | 110 | 116 |
| internal/load/testdata/0.1/compile/09_tree.ts | 82 | 85 | 123 | 126 |
| casts.a | 16 | 17 | 24 | 25 |
| visits.a | 124 | 125 | 208 | 209 |
| narrowed_reads.a | 0 | 3 | 5 | 7 |
| narrowed_methods.a | 1 | 3 | 5 | 6 |
| narrowed_fields.a | 2 | 4 | 5 | 6 |
| fills.a | 27 | 28 | 50 | 51 |
| fresh_writes.a | 407 | 409 | 523 | 525 |
| fresh_calls.a | 266 | 268 | 321 | 323 |
| weak_narrowed.a | 50 | 53 | 69 | 72 |
| undefined_keys.a | 171 | 173 | 236 | 238 |
| undefined_strings.a | 20 | 21 | 42 | 43 |
| undefined_references.a | 36 | 37 | 51 | 52 |
| reuse_narrowed.a | 4 | 6 | 7 | 8 |
| reuse_lent_global.a | 7 | 8 | 14 | 15 |
| regexp.a | 355 | 369 | 408 | 422 |
| regexp_null_narrowed.a | 7 | 8 | 5 | 5 |
| regexp_exec.a | 81 | 82 | 159 | 160 |
| regexp_unicode.a | 99 | 101 | 86 | 88 |


## TypeScript source fixtures

Merged `codex/stage3-fixtures-enums` at `671e1fd` with merge commit `df351a6`.
The thirteen upstream fixtures remain unchanged. The audited before count was
6 Compiles, 5 Refused and 2 Checker. The observed after count is 7 Compiles,
4 Refused and 2 Checker. All seven native observations match source Node on
stdout, stderr and exit code. Exact observations are in
[implementation-results.json](../stage3/fixtures/enums/implementation-results.json).

Same-enum member aliases now preserve flag classification. Every declaration
still needs explicit qualifying initializers and every member value remains a
non-negative int32. A const for-of binding over a fresh inline array of proven
flags also preserves its domain when TypeScript infers number[]. An aliased
mutable array, a mutable loop binding, an unproven element and a different enum
cannot acquire that proof. Shorthand fields resolve their value symbol, rather
than the field symbol, and use the same domain checks as explicit fields.
These changes make `04_parse_tree_mask.a` compile without changing its source.

Name enumeration over a runtime enum and its typeof aliases now lowers through
`for...in`. Its fixed shape includes numeric reverse keys, sorted as JavaScript
integer keys, followed by forward member names in declaration order. String
enums have forward names only. `enums_names.a` holds these rules to Node.
This support does not admit an arbitrary any object or inherited unknown keys.

The six remaining boundaries and their required adaptations are documented in
[the fixture report](../stage3/fixtures/enums/README.md#implementation-follow-up).
`05` and `09` need a signed-mask contract or a source adaptation. `06` needs
an invariant flag field rather than a generic mutable cast. `07` needs a
readonly Map view and an explicit sparse-table representation. `08` and `11`
need typed reflection and checked reads instead of any and unchecked indexing.

The merge had four conflicts. `lower.go` retains both enum initialization
preflight and accessor-name registration. `class_inheritance.go` retains both
enum identity/invariance and recursive nominal checks. `refusals.go` retains
both enum-specific and definite-assignment refusals. `counts.md` retains both
branches' rows; counts are regenerated for the combined fixture set.

The full uncached `TestNativeAgreesWithNode` suite and the six boundary tests
passed in 108.586s (native hits 0, Node hits 0). All seven real compiling fixtures
and the name-enumeration fixture run source Node, generated JavaScript, native
release, ASan/UBSan and LeakSanitizer. Full lower/load tests passed in 37.094s
and 2.151s. The permanent name-ordering mutant passed in 0.485s: it compiled and
exited 0 with empty stderr, and only Node stdout comparison caught its error.

Six temporary compiler mutants were restored after each run and caught by
assertion failures, with no build failures counted:

| Mutant | Named catch |
| --- | --- |
| Disable direct alias classification | TestFlagEnumMemberAliases |
| Accept a foreign enum alias | TestFlagEnumAliasBoundaries |
| Skip inline iterable element proofs | TestFlagEnumAliasBoundaries |
| Preserve a mutable loop binding's proof | TestFlagEnumAliasBoundaries |
| Skip shorthand value/domain proof | TestFlagEnumInlineIteration |
| Disable enum-object for-in origin proof | TestEnumNameEnumeration |

Logs are `/tmp/stage3-enums-{oracle-restored,packages,mutants,permanent-mutant-restored}.log`.
Setup's first warm build overlapped the unresolved merge and failed on conflict
markers. After resolution, setup passed: Go, clang, Node and submodules ready
at 0s; cache warm and total at 125s; nproc 5, quota four CPUs. The complete Go
repository test gate and the complete TypeScript compiler remain untested.

Counts regeneration passed in 16.205s and adds eight rows, the seven compilable
source fixtures and `enums_names.a`. No existing numeric counts changed.
`class_inheritance_conditional.a` moved in table order only. Vet, gofmt and
whitespace outputs are empty.

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower ./internal/load -count=1 -timeout 30m > /tmp/stage3-enums-packages.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStage3EnumBoundaries|TestNativeAgreesWithNode' -count=1 -v -timeout 30m > /tmp/stage3-enums-oracle-restored.log 2>&1
go test ./internal/oracle -run 'TestEnumNameEnumerationMutant|TestStage3EnumBoundaries' -count=1 -v -timeout 30m > /tmp/stage3-enums-permanent-mutant-restored.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/stage3-enums-counts-update-restored.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m > /tmp/stage3-enums-counts-verify.log 2>&1
go vet ./... > /tmp/stage3-enums-vet-restored.log 2>&1
gofmt -l cmd internal > /tmp/stage3-enums-format-restored.log
git diff --check > /tmp/stage3-enums-diff-restored.log
```

## Numeric enums are open

Ruling by @system_adamic, October 7, 2026: a numeric enum is a number with named
constants. Any value typed as `number` can be assigned to it. The numeric closed
union, flag subset domain, non-negative bound and numeric declaration identity
restrictions are removed. String enums remain closed. A heterogeneous enum is
not an entirely numeric enum and retains the preceding representation limits.
TypeScript's checker still decides whether the source itself is well typed.

Arithmetic, ranges, casts from numbers, increments and compound updates work.
Flag operators retain JavaScript's ToInt32 behavior. OR, XOR, AND, complement
and shifts can produce signed numbers, and their results can be stored in a
numeric enum slot. `1 << 31` is admitted. Combining members from two numeric
enums produces a number that can be stored in either enum when the checker
admits the assignment. These accepted programs supersede the earlier refusal
examples:

```ts
enum SyntaxKind { First, Last }
enum Flags { None = 0, A = 1 << 0, B = 1 << 1, Sign = 1 << 31 }
function kind(value: number): SyntaxKind { return value as SyntaxKind; }
let value: SyntaxKind = kind(42);
value++;
value += 7;
value <<= 1;
const inRange = value >= SyntaxKind.First && value <= SyntaxKind.Last;
const combined: Flags = Flags.A | Flags.B;
const masked: Flags = combined & ~Flags.A;
const complement: Flags = ~Flags.A;
```

Reverse numeric lookup remains `string | undefined`: an unnamed number or
combination has no reverse property. Enum objects themselves retain exact shape
and immutable named constants. Structural copies cannot acquire `typeof E`.
`Object.assign(E, { A: value })` is refused with `adamic/enum-object`; copy into
an ordinary object before writing it. This preserves reverse mapping and the
member-origin proof described below. Ordinary enum name enumeration still works
through `Object.keys` and `for...in`; Debug's enumeration is not treated as a
numeric reverse lookup.

### Never is checked

The checker can narrow a numeric enum to `never` after excluding its declared
members, even though another number can exist at runtime. Each such read gets a
non-returning check in ordinary typed IR. Native and generated JavaScript both
exit 70 with, for example:

```text
adamic: panic: unreachable value 42 for numeric enum SyntaxKind
```

The unmatched edge of a switch covering the checker's possible members gets the
same check, including an implicit default. A switch listing only some members
can use an ordinary default; string switches retain their closed coverage rule.
The checked switch stores its scrutinee once, so a call or lookup with effects
is not repeated by the diagnostic. The never-default idiom lowers. If chains,
field reads, optional enum values, increments, compound reads and numeric never
returns into a string slot use the same check. Non-enum `never` handling is not
opened by this change.

A numeric default is erased only when coverage is complete and its value came
from an actual member, directly or through an immutable `const` alias. A type
annotation, parameter, mutable slot, numeric literal, cast, arithmetic result
or flag combination is not that proof. Proofs are conservative: checking a
parameter against a member does not currently erase later checks automatically.

```ts
enum E { A, B }
function describe(value: E): void {
    switch (value) {
        case E.A: return;
        case E.B: return;
        default: { const unreachable: never = value; console.log('unreachable'); }
    }
}
function other(): number { return 42; }
describe(other()); // Accepted source, checked stop 70 in both backends.
describe(E.A);    // Accepted, agrees with Node.

const member: E = E.A;
switch (member) {
    case E.A: break;
    default: { const unreachable: never = member; console.log(`${unreachable}`); }
} // The default is erased with a member-origin proof.
```

Programs still refused include an arbitrary string cast to a string enum,
a string member from another declaration, an uncheckable `as never`, a write
into the runtime enum object and an unproven write through `Mutable<T>`.
String enum switches missing members still need a default or the missing cases.
Computed/non-finite declarations, ambient or merged definitions, and unsupported
mixed representation switches retain their named limits. This ruling does not
turn off checker errors or admit `any`.

### Original TypeScript fixtures

The original thirteen sources are unchanged. `open-results.json` records fresh
source Node, generated JavaScript and ASan/UBSan observations; every program
that finishes also passed LeakSanitizer. Before is `246ecc0`, after the two
integration merges and before this ruling. The original fixture audit was
6 matches, 5 refusals and 2 checker failures; the first implementation reached
7 matches, 4 refusals and 2 checker failures.

| Fixture | Before ruling | After ruling |
| --- | --- | --- |
| 01 token range | Node match | Node match |
| 02 node range | Node match | Node match |
| 03 JSDoc range | Node match | Node match |
| 04 parse tree mask | Node match | Node match |
| 05 local/export signed SymbolFlags | Refused | Node match |
| 06 generic setNodeFlags | Refused | Refused: Mutable<T> cannot prove its field's write contract. |
| 07 regex reverse array map | Refused | Compiles; pinned stop 70 on index 1 of an empty array. |
| 08 Debug formatter | Checker | Checker: TS2532; unchecked array reads and any remain. |
| 09 signed exclusion mask | Refused | Node match |
| 10 string enum | Node match | Node match |
| 11 SyntaxKind formatter | Checker | Checker: TS2532; unchecked array reads and any remain. |
| 12 diagnostic reverse lookup | Node match | Node match |
| 13 regex Map keys | Node match | Node match |

That is 7 Node matches, 0 checked stops, 4 refusals and 2 checker failures before;
9 Node matches, 1 checked stop, 1 refusal and 2 checker failures after. Ten
sources now compile, but the sparse-array program is not a Node-equivalent
completion. Its existing array safety boundary remains in force.

05 and 09 also exposed their original optional-object ternary condition after
the enum restriction was removed. The small `control.go` hook lowers an optional
object to a presence test, evaluating it once. Numbers and strings still require
explicit boolean comparisons. Both original signed-mask fixtures now match Node
byte for byte in native release, sanitized native and generated JavaScript.

### Ruling verification and mutants

Nine new `enums_open*.a` fixtures cover arithmetic, ranges, number assertions,
signed and combined flags, fields, arrays, Map values, member-only switches,
explicit and implicit never branches, if chains, optional values and updates.
The seven firing fixtures have exact exit, stdout and value/enum stderr assertions
in both backends. The implicit-default fixture proves a side-effecting scrutinee
is evaluated once. Successful fixtures retain the full Node/release/sanitizer/leak
comparison.

All eleven temporary compiler mutants were restored. Every catch is an assertion
or observed runtime disagreement, and no build failure is counted:

| Mutant | Named catch | Observation |
| --- | --- | --- |
| Remove never panic, return the value | TestNumericEnumNeverPinned | Both backends exit 0 and print unreachable/after; required exit is 70. |
| Erase default without a member-origin proof | TestNumericEnumNeverPinned | Both backends exit 0 and print after; required exit is 70. |
| Close numeric assignability again | TestNumericEnumsAreOpen/arithmetic | Lowering refuses the arithmetic return with the old closed-union message. |
| Invert optional-object presence | TestNativeAgreesWithNode/stage3/fixtures/enums/05 | Native and JavaScript stop where source Node exits 0. |
| Permit runtime enum mutation | TestEnumSlotViews/object_assign | Expected refusal disappears. |
| Repeat switch scrutinee effects | TestNumericEnumNeverPathsPinned/implicit | Lookup is printed three times instead of once. |
| Omit numeric never increment check | TestNumericEnumNeverPathsPinned/update | Both backends exit 0 instead of 70. |

The existing six semantic mutants, name-order mutant and cleanup mutant remain
held by their original checks. The cleanup mutant now disables conservative dead
stack/register roots in LeakSanitizer, retaining global roots; this prevents a
stale temporary from hiding a deliberately leaked enum object.

Counts add twelve rows: nine new fixtures, signed fixtures 05 and 09, and the
checked sparse-array fixture 07. Every pre-existing row's numbers are unchanged.
Stops intentionally retain what they held when panic ended the process; programs
that finish have equal allocations/frees and clean leak checks.

Full lower/load passed in 13.017s and 1.088s, counts regeneration in 28.850s.
Vet, formatting and whitespace logs are empty. Final uncached oracle and counts
results, both integration SHAs, all conflict resolutions and inherited counts
rows are recorded in [the integration report](verification/enums-open/report.md).
No main or area branch was merged into or pushed. Only the two requested codex
feature branches were pushed; the ruling commits belong to codex/flag-enums.

### Numeric literal promises

The whole numeric enum is open. A member-specific annotation such as `Kind.A`
in a multi-member enum still promises that literal value. It cannot accept an
unproven number or rely on excluding other members from an open enum. Likewise,
assigning an open singleton enum to a plain numeric literal requires a proof.
These programs are refused with `adamic/enum-literal`; widen the destination to
the whole enum or number, or compare and return the actual literal constant.

```typescript
enum Kind { A, B }
function numeric(): number { return 1; }
const tag: Kind.A = numeric(); // Refused.
function remaining(k: Kind): Kind.A {
  if (k === Kind.B) return Kind.A;
  return k; // Refused: another number may remain.
}
```

A singleton whole enum remains open. It cannot prove a member-specific literal
promise merely because the checker represents both with the same type. Written
member annotations are recovered for construction, casts, structural object
views and mutable containers. Construction and ordinary views require a proof;
explicit casts insert a member-value check. An ambiguous union selected by an
open tag receives checked payload reads. Primitive storage is checked before
reading, followed by literal-value checks when needed; nominal class views use
the existing class check. Incompatible structured or optional payload views
remain `NotYet` rather than reading the wrong storage.

`TestNumericEnumLiteralPromises` covers those promises and excluded-member
narrowing. `TestNumericEnumsAreOpen` proves singleton numeric assignments and
member-tagged object unions compile. The historical singleton refusal mutant
is superseded by `TestEnumTagPayloadMutant` and pinned checked-view witnesses.

Array element reads also recover their declared enum element type before flow
narrowing. `enums_open_never_index.a` pins an out-of-domain array value at the
never binding in both backends. Removing that recovery makes
`TestNumericEnumNeverPathsPinned/index` fail at lowering, without a Go or C build
failure. `TestNumericEnumsAreOpen/coalesce` preserves optional enum Map values
through nullish coalescing; absence is checked separately from numeric values.

## Literal tag narrowing follow-up, October 7

The ruling by @system_adamic permits numeric member tags to narrow object unions
by value, including `===`, `!==`, switches, and marker aliases such as `FirstX`.
An unrelated open numeric enum field in a variant is no reason to reject that
refinement. `enums_tag_narrowing.a` holds these operations to independent Node,
including literal construction, an exact class and an existing checked downcast.

An explicit numeric enum default remains reachable after all declared values
are excluded. It runs normally until an actual `never` read inserts the existing
runtime check. The fixture `enums_tag_never.a` prints `default` before stopping
with exit 70 and exactly:

```text
adamic: panic: unreachable value 42 for numeric enum SyntaxKind
```

Singleton checked views now cover primitive payloads, literal payload values,
member-specific casts, nominal class views and property receivers. Native shape
identity includes actual field types; a checked read verifies its layout before
interpreting an untagged slot. JavaScript checks the value independently. Object
remainders retain the full stored union for checker-admitted observations and
typed aliases, including after all values of aliased members are excluded. A
`never` assertion on that object stops with the tag value in its message.

Unmatched implicit numeric switches fall through. A terminal switch in a
function requiring a result instead has a checked missing-result site, whose
message is `numeric enum switch fell through a function requiring a result`.
It does not prove numeric exhaustiveness.

The complete ruling still has limits: incompatible structured or optional
payload views remain `NotYet`, and a direct remainder property read rejected by
the source checker as a property of `never` never reaches lowering. A typed
alias to the full union permits that read. The final census retains 90 of the
101 discriminant sites under the explicit checked-view `NotYet` boundary; a
zero in the old refusal row does not mean those sites all compile.

The [final unit report](verification/enum-tag-final.md) records the negative
control on main, the discriminant-only checkpoint, final meter evidence,
fixtures, checks, mutants and changed files. The [initial report](verification/enum-tag-narrowing.md)
is retained as history.

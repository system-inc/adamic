# Enums in Adamic 0.2

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
bare nonzero literals, direct member aliases and arithmetic initializers keep
an enum closed and switch-exhaustive under the preceding rules.

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

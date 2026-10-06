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

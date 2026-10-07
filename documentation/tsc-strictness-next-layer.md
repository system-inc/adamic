# Meter returns, argument contracts and non-erasable syntax

For Kirk and Ahra, continuing the [strictness survey](tsc-strictness.md) and
[optional-property adaptation](tsc-strictness-adaptation.md). No language,
checker options, lowering rules or runtime rules were changed. The only new
adapter makes already-undefined function completions explicit, in memory.

## TS7030: implement only the declared contract

The previous adapted measurement contained 251 TS7030 findings. **109 functions
have an explicit return annotation whose resolved type contains undefined**;
**142 have inferred return types**. The new rule removes exactly the 109, leaving
all 142 inferred cases. Resolved aliases count: for example,
`CompilerOptionsValue` and `SearchResult<T>` include undefined despite not spelling
it in the function header. An explicit `T | undefined` also permits undefined for
every specialization; a bare `T` does not.

```ts
function lookup(flag: boolean): string | undefined {
    if (flag) return "found";
    // Before: implicit completion with undefined.
    // Adapted, immediately before this body's closing brace:
    return void 0;
}
```

`void 0` is an explicit undefined return that remains correct when a parameter
or local shadows the identifier `undefined`. Appending it at the function body's
end preserves all preceding statements. Bare `return;` statements belonging to
that function become `return void 0;`: they keep return completion, automatic
semicolon insertion and `finally` behavior. Nested functions own their returns
and are considered separately. Edits are insertions at parsed byte offsets,
including UTF-16 diagnostic columns, with a newline before the appended return
so a trailing line comment cannot swallow it.

The rule requires a TS7030 diagnostic, an explicitly annotated ordinary function
with a block body, and ownership of its source file as a meter root. It excludes
inferred returns, any/unknown/void alone, unconstrained generic contracts,
async functions, generators, expression bodies and imported-only files. These
exclusions are deliberate scope limits, not claims that every excluded function
is unsound. Its tooling-only resolver never reaches lowering; `LoadOverlay`
rechecks the edited program with Adamic's authoritative checker gate.

The JSON adaptation entry reports `functions_changed: 109` and
`diagnostics_removed: 109`. Disk sources are unchanged. This is not a default
value, a checked unwrap, or permission to append an undefined return to a
function declaring only `number`.

## TS2345: 100 findings from the 693 population

Population: the 2,070-diagnostic result at `be67356`, after imports and optional
properties, before the return rule. Sort TS2345 findings by relative filename,
line and column; select zero-based ranks `floor((k + 0.5) * 693 / 100)` for
`k = 0..99`. All rows are inside the pinned TypeScript 6.0.3 `src/compiler`.
[Complete ledger](tsc-strictness/argument-sample.json): A001..A100, full diagnostic
chains, source lines, numbered context, ranks, patterns and individual decisions.
This is systematic sampling, not a random estimate of the entire compiler.

M means a meaning-preserving candidate, subject to declaration ownership and
consumer rechecking. C means an unproved invariant or legitimate absence that
needs review: if the value is required, use an explicit loud check at its
original evaluation point. If absence is legitimate, retain or design its
handling instead. U means the contract is demonstrably unsound for admitted
inputs; it does not establish a reachable upstream compiler bug.

| Pattern | Count | Decision | Examples |
| --- | ---: | --- | --- |
| A-loop | 35 | C | A015, A027, A043, A089 |
| A-parallel | 24 | C | A005, A030, A051, A076 |
| A-nonempty | 19 | C | A012, A041, A052, A096 |
| A-generic-array | 9 | U | A058, A061, A066, A067 |
| A-optional-value | 3 | C | A003, A025, A046 |
| A-result | 3 | C | A002, A048, A070 |
| A-regex-capture | 2 | C | A082, A100 |
| A-own-key | 1 | C | A064 |
| A-optional-view | 1 | M | A074 |
| A-narrowing | 1 | M | A079 |
| A-coercion | 1 | M | A098 |
| A-overload | 1 | M | A099 |
| **Total** | **100** | **4 M, 87 C, 9 U** | |

Bounds and nonempty tests do not prove dense storage. A012 checks that
`parameters.length === 1` before reading `[0]`; A096 checks an argument count of
three before reading `[1]`. Ordinary arrays can have holes. A041 has
`Debug.assert(typeParameters[i] !== undefined)` before another read. An assertion
whose behavior depends on the debugging level is not an unconditional production
check, and rereading an indexed value needs stability evidence. For parallel
arrays, a bound on one array does not prove an element in the other. Capturing
and checking the original read is the conservative route, pending a construction
proof that can move an individual case into M.

The 9 U cases are generic array helpers in `core.ts`. A058's `forEach<T>` accepts
`readonly T[]` but supplies `array[i]` to a callback requiring `T`. A typed
`new Array<number>(1)` has length one and supplies undefined. A callback can also
delete a later element from an initially dense array. **Node demonstrates both**.
The source-local contract is unsound over these admitted arrays. That does not
show that TypeScript's actual compiler callers supply sparse arrays. A dense
representation and stability contract, checked reads, or a consciously different
hole policy is needed. Replacing indexed loops with `for...of` does not prove
elements exist: that iteration also visits holes as undefined. Skipping holes
would change callback counts and effects.

The four M candidates have distinct proofs:

* **A074, `moduleNameResolver.ts:1450`: optional view alignment.** The source's
  `isReadonly?: boolean | undefined` is passed to a view with `isReadonly?: boolean`.
  Consistently representing present undefined in the owned view can preserve all
  runtime expressions and key presence. Resolve ownership and specialized/base
  relationships and recheck consumers. The existing adapter only seeds from
  TS2412/2375/2379; it does not yet seed from TS2345.
* **A079, `program.ts:4158`: direct narrowing.** `subst` is a const local and
  `typeOfSubst` is `typeof subst`. Change the branch condition from
  `typeOfSubst === "string"` to `typeof subst === "string"`, retaining the alias
  where used elsewhere. This recomputes a pure operation on the same local and
  exposes the existing check to the checker. Do not generalize to a getter or
  mutable captured variable.
* **A098, `utilities.ts:7765`: existing builtin coercion.**
  `base64Digits.indexOf(input[i + 3])` already applies ToString to a string or
  undefined search argument. Making it `String(input[i + 3])` preserves the
  coercion, including invalid/truncated inputs, under the standard builtin
  bindings. A `charAt` replacement changes out-of-range undefined to an empty
  string and changes the result. This proof is specific to this builtin and
  these value types; it is not permission to stringify arbitrary arguments.
* **A099, `utilities.ts:11201`: overload selection.** Add a checked annotation to
  the captured `stringReplace`, selecting its string/string overload:
  `(this: string, search: string, replacement: string) => string`. Keep the captured
  `String.prototype.replace`, `.call`, receiver, arguments and evaluation order.
  No cast is needed. The last overload selected by generic `.call` causes this
  rejection. Stock TypeScript accepts the checked narrower callable view and
  Node confirms unchanged first-star and replacement-string expansion behavior.

Only the TS7030 rule is implemented here. The four argument candidates are
survey recommendations. Loud checks change failure behavior on invalid states;
they are not meaning-preserving rewrites over all JavaScript inputs. Neither
`!`, a cast, nor broadening a required callee parameter proves presence.

## TS1294: complete syntax census

All 180 findings were classified through their AST declaration, not a text search.
[Complete census](tsc-strictness/syntax-census.json) records every location, name,
source line, kind, const modifier and resolved enum value family.

| Syntax | Findings | Detail |
| --- | ---: | --- |
| Enum declarations | 164 | 139 const, 25 ordinary |
| Runtime namespaces | 10 | Includes nested `Debug.log` and `Parser.JSDocParser` |
| Parameter properties | 6 | Six readonly callback parameters in one state-machine constructor |
| **Total** | **180** | No other non-erasable syntax kind in this population |

The enum unit is a declaration; parameter properties are counted per parameter.
Type-only namespaces that do not cause TS1294 are outside this diagnostic census.
Stock TypeScript resolves **157 enums entirely to numeric constants** and **7 to
string constants**. There are no mixed, nonconstant or unresolved member-value
families in these 164 declarations. Const enums split into 133 numeric and 6
string; ordinary enums into 24 numeric and 1 string. These are declaration
counts, not execution frequency or implementation-time measurements.

Supporting enums would require checked enum types and member resolution in
lowering, constant evaluation (including references, auto-increment, bitwise
flags and computed constant expressions), and exact representation in both
backends. Ordinary numeric enums expose a runtime object with forward names and
numeric reverse mappings; string enums expose forward mappings only. Duplicate
numeric values, object key order, exported runtime values and declaration merging
must keep TypeScript/Node behavior. Numeric enums are not automatically a closed
finite set: variable numbers and bitmask combinations have TypeScript rules that
need an explicit design decision under Adamic's true-types contract.

The 139 const enums make static member support a promising first subset, but
**do not assume all const enums can simply disappear**. Upstream's
`src/tsconfig-base.json` enables `preserveConstEnums: true`; that output can retain
runtime enum objects while inlining references. Adamic's exact enum policy,
runtime observability, imports and backend oracle must be specified. A static
constant-member subset with explicit refusal for unimplemented runtime use can
be small; it does not by itself support the full TypeScript compiler. Of the
syntax proposals, enums address the largest measured wall, **164/180**.

Namespaces require scoped names and exports, merge resolution, initialization
order and live runtime storage. A type-only namespace can disappear, but these
10 rejected declarations carry runtime values. Their emitted namespace object,
reopening, nested objects, function/enum/class merging, and cyclic initialization
must preserve behavior and obey ownership/cycle proofs. Flattening them into
modules is not an unconditional source rewrite; namespace identity and effects
can be observable.

Parameter properties require constructor parameters to become instance fields
with correct initialization and accessibility/readonly information. Six findings
belong to one `BinaryExpressionStateMachine` constructor in `factory/utilities.ts`
(lines 1415..1420). Lowering must preserve parameter evaluation, base-constructor
ordering, field initialization order, property descriptor semantics and behavior
when a constructor returns another object. A blanket `this.x = x` insertion is
not a complete implementation. This small observed case may permit a tightly
specified subset, but it addresses fewer blockers than enums.

Recommendation: keep the strictness checks, retain only proven mechanical
adaptations, and put static enum/member support first among these syntax design
questions. Keep namespace support and parameter-property ordering as separate
language decisions rather than silently enabling non-erasable syntax in the meter.

## Re-measurement and validation

Pinned input: TypeScript 6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`, 77 roots under `src/compiler`.
[Final report](tsc-strictness/meter-returns.json).

| Measure | Previous adapted result | With explicit returns |
| --- | ---: | ---: |
| Checker diagnostics | 2,070 | 1,961 |
| TS7030 | 251 | 142 |
| TS2345 | 693 | 693 |
| Entries reaching lowering | 0 | 0 |

All other checker-code counts are unchanged. Leading reasons are now TS2345
693, TS18048 354, TS2532 224, TS1294 180, TS7030 142, TS2322 102, TS7029 83 and
TS2591 54. Type-only imports still remove 3,718 diagnostics; optional declarations
still change 415 declarations and remove 693 exact-optional diagnostics.
An independent diagnostic export contains exactly 1,961 diagnostics and the same
three adaptation entries. Independently tried roots whose own files are clean
still fail their dependency closures, so no first Refused or NotYet result exists
for this input.

Commands used, with test output saved directly to logs:

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic-meter --adapt --json /tmp/adamic-tsc-strictness/typescript/src/compiler > /tmp/adamic-tsc-strictness/meter-returns.json 2> /tmp/adamic-tsc-strictness/meter-returns.stderr
go test ./cmd/adamic-meter ./internal/load -count=1 > /tmp/adamic-tsc-strictness/returns-tests-final.log 2>&1
go vet ./... > /tmp/adamic-tsc-strictness/returns-vet.log 2>&1
go test -count=1 -timeout 30m ./... > /tmp/adamic-tsc-strictness/returns-gate.log 2>&1
TSC_SURVEY_TYPESCRIPT=/tmp/adamic-tsc-strictness/npm/node_modules/typescript/lib/typescript.js node docs/tsc-strictness/layer-probes.cjs > /tmp/adamic-tsc-strictness/layer-probes.log 2>&1
TSC_SURVEY_TYPESCRIPT=/tmp/adamic-tsc-strictness/npm/node_modules/typescript/lib/typescript.js node docs/tsc-strictness/layer-audit.cjs /tmp/adamic-tsc-strictness/diagnostics-optional-final.json /tmp/adamic-tsc-strictness/typescript/src/compiler > /tmp/adamic-tsc-strictness/layer-audit.log 2>&1
```

Vet, formatting, touched-package tests and the complete worker gate pass. The
full gate includes native/Node oracle comparisons and both TypeScript scanner
and parser slices; the unrelated exhaustive Unicode package finishes in 536.828s.
This is the ordinary worker gate with oracle caching permitted, not the uncached
integration gate. After strengthening only the return fixture's finally override
and UTF-16 case, the final touched-package run also passes: meter 2.519s, loader
1.075s. The 100-row/180-finding audit and all five stock-TypeScript/Node probes
pass. No new native fixture or counts-table row is needed for this CLI-only edit.

The Node fixtures compare original and adapted source, with checker acceptance
required before the comparison. They cover explicit aliases and generic unions,
shadowed undefined, fallthrough effects, bare returns, nested functions, methods,
finally effects and finally overriding a return, trailing comments, UTF-16 columns,
`.ts`/`.a`/real `.a.ts`, root ownership, unchanged disk bytes and idempotence.
Negative fixtures retain inferred, required, arbitrary generic, any/unknown,
async and generator contracts. The stock TypeScript/Node probes independently
verify the four M pattern examples and sparse-array/callback-deletion counterexamples.
Stock TypeScript reports TS2379 for the reduced optional-view argument while the
pinned Go checker reports TS2345 for A074; the incompatibility is the same.

Actual unsafe adapter mutant: replace the insertion position immediately before
the function's closing brace with immediately after its opening brace, adjusting
the matching brace validation. It still typechecks and removes all four fixture
TS7030 diagnostics, but returns before side effects and existing value returns.
`TestAdaptUndefinedReturnsPreservesNodeAndDisk` catches it on Node stdout in all
three source extensions, exit 1. The mutant prints only seven undefined results;
the original includes fallthrough/finally/arrow effects and values 5 and 7.
It is caught by runtime equivalence, not a parser/type error or compiler warning.
It was run through a Go file overlay, leaving the implementation unchanged.

For audit reproduction, export the previous result using the exporter at
`be67356` with `ADAMIC_SURVEY_OPTIONAL=1` in a scratch worktree; follow the
[original exporter instructions](tsc-strictness.md#reproduction-and-validation).
Run the new layer-audit script against that export and the pinned compiler root.
The exporter on this commit uses all three rewrites when that flag is set; its
final result has shifted source lines and is deliberately not the baseline sample.
The original 200-row survey and historical reports remain intact.

Setup succeeded: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 16s, total
16s. `nproc` is 5; CPU quota is 4. Validation outcomes and remaining scope are
recorded here rather than treating fewer diagnostics as proof the compiler port
can yet compose.

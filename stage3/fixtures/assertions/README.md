# Assertions from TypeScript 6.0.3

This bucket contains 20 programs, the complete 4,101-site assertion ledger, and
all 1,123 non-null sites. Source is TypeScript v6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`. Adamic was observed on current main
`ef3d907ecdc4c771b016f7d9c52372def057a340`. The census reference is
`429c1177f0130f785c19cf590d1860513b2ddbfc`; the pipeline README was read from
`8728405135d329efc12c837a7a6c293234abbe1c`.

## The complete ledger

Stock npm TypeScript 6.0.3's compiler API checked the generated upstream project
with its own inherited tsconfig and pinned Node declarations: **zero diagnostics**.
The classification uses checked operand types, including control-flow narrowing,
and `isTypeAssignableTo` in both directions. It does not use the spelling of a
cast's type as a substitute for its actual type. Every original compiler file's
SHA256 is checked against the census. Generated diagnostics are excluded.

| Exclusive category | Sites |
| --- | ---: |
| Upcast | 208 |
| `as const` | 16 |
| Downcast from a discriminated union | 136 |
| Downcast between structural interfaces without a narrowed runtime tag | 1,178 |
| Nodes in `as unknown as` chains | 10 |
| Other | 2,553 |
| Total | 4,101 |

All 4,101 are `as` expressions, in 61 files; there are no angle-bracket assertions
in this population. [ledger.json](ledger.json) has every file, line, column,
start/end offset, expression, source/target type, both assignability results,
category, candidate discriminant and its literal values. [locations.tsv](locations.tsv)
is the compact location index. [ledger-summary.json](ledger-summary.json) has counts.

The precedence is `as const`, both nodes of an exact `as unknown as` chain,
upcast, tagged union downcast, untagged interface downcast, other. The ten chain
nodes are **five chains**, not ten independent double casts. Upcast means stock
assignability, including identity assertions; it does not certify Adamic's
additional variance/readonly rules. In particular, a mutable view may be
stock-assignable and still be rightly refused by Adamic.

A tagged union downcast requires reverse assignability, a required property with
a strictly narrower finite literal type, and a partition: every source member
whose tag overlaps the accepted values must be assignable to the target. This
avoids accepting two differently shaped members with the same tag. A tag may be
a literal union rather than one constant. **134 sites discriminate by `kind`;
2 by `type`** (both in `tsbuildPublic.ts:1733`). This is type-level checkability,
conditional on the source union's types already being true.

For the untagged interface category, every source and target constituent must
resolve to an interface declaration, reverse assignability must hold, and no
required property may narrow from a broader type to a finite set of literals.
A broad `flags: TypeFlags` field does not discriminate a sub-interface whose
`flags` has that same broad type. Typical targets are `UnionType` (155),
`TypeReference` (84), `IntersectionType` (81), `ConditionalType` (57),
`IndexedAccessType` (55), and `MappedType` (43). Masks in the surrounding code
may supply additional invariants; this category does not pretend the interfaces
encode those invariants.

Other is fully located and subdivided: **1,842** tagged narrowings outside a
partitioned union; **264** other narrowings; **343** non-assignable assertions;
**104** with a top-level `any` operand or target. Of the 1,842 candidate-tag sites, **1,825** use `kind`,
10 `flags`, 2 `length`, 2 `version`, and one each `isTypeOnly`, `type`, `operator`.
These are candidates, not permission to treat every field as a sound type tag.
Branded strings (`__String`), arrays, type parameters, intersections and
construction-time assertions are retained rather than forced into interface or
union categories. In the union category the most common target is `Identifier`
(29); other includes 57 further `Identifier` targets. All such sites remain
individually reviewable.

## The hard case: Node as Identifier

`src/compiler/types.ts:942` declares **Node as an interface**, with
`readonly kind: SyntaxKind`; it is not a union of all concrete syntax nodes.
At `types.ts:1701`, `Identifier` narrows `kind` to `SyntaxKind.Identifier` and
requires `escapedText`. The assertion itself does not cause stock tsc to verify
that the object has those fields. Stock tsc accepts the assertion because the
sub-interface is structurally related to Node. Source branches often test kind
first, but checking that branch still does not validate the whole shape.

The useful runtime test is a nullish check followed by
`node.kind === SyntaxKind.Identifier` (the numeric enum value **80**). For a
multi-kind target, compare against every allowed kind. Evaluate the operand
once. No `instanceof Identifier` check exists: that interface has no constructor.
This test is sound **only with a proven invariant that every reachable Node with
kind Identifier is a fully initialized Identifier**, including its required
fields and their types. The invariant must hold at construction, through mutable
views, aliases, cloning and every kind write, and at external boundaries.

There is outside evidence that the invariant does not follow from the stock
interface alone. [kind-soundness.cjs](kind-soundness.cjs) creates a genuine stock
factory PlusToken, stores it as `ts.Node`, changes its kind through
`Mutable<ts.Node>`, and casts it to `ts.Identifier`. Stock tsc accepts with **zero
diagnostics**. Node prints **`true\nundefined\n`**: the kind check passes but
`escapedText` is absent. [logs/kind-soundness.json](logs/kind-soundness.json)
records the complete source and observations. This is a counterexample to the
unconditional tag-only rule, not an observation of a native miscompile.

A native compiler can use a proven closed union or a trusted construction
invariant. Alternatively it can check an actual runtime shape descriptor that
certifies the required initialized fields and their value types, or validate
those fields explicitly. Testing only for the existence of one field is not a
complete structural validation. None of these proofs is automatically available
from Node's interface declaration. Factory construction may assert a target
before populating its fields, and many Type/Flow interfaces use masks rather
than literal discriminants. Such sites need construction proofs or source
adaptations; where a check cannot establish the target, the ruling requires
refusal. This unit inventories them; it does not assert a whole-program proof.

## Non-null forms and a source-contract conflict

[nonnull-ledger.json](nonnull-ledger.json) gives all locations, in 49 files:
453 field operands, 202 call operands, 2 indexed operands, 111 literal
`undefined` operands, and 355 other operands (primarily identifiers). Parentheses
are ignored in assigning these forms. `operationArguments![operationIndex]!`
counts its identifier unwrap and indexed unwrap separately; a regex-call unwrap
followed by `[0]` is a call operand, not an indexed operand.

The ruling is a loud nullish check naming the expression, with one evaluation,
not erasure. Valid present values in these fixtures can therefore be compared
with plain Node stripping. However **111 original sites explicitly assert
`undefined!`**. For example `core.ts:1896` clears memoize's callback;
`scanner.ts:4025` clears tokenValue; `binder.ts:594` clears file. A loud unwrap
there must fail, while stock tsc runs these successfully. Fixture 06 preserves
this real case and Node prints `42\n42\n1\n`. Native source parity needs truthful
optional state declarations and ordinary undefined assignments at such sites.
This unit makes no adaptation. Do not mark fixture 06 Compiles simply because a
future checked unwrap compiles it: its successful output must still match Node.
The scanner's codePointAt helper also explicitly comments that its missing-value
contract is wrong; the fixture exercises the present inputs, not that unresolved
end-of-input contract.

## Fixtures and observed status

Each fixture starts with its upstream source location and exact census reason.
Nineteen retain complete upstream function bodies and signatures, including
nested helper functions extracted into a standalone file. Fixture 18 retains the
one real cache tuple statement at `checker.ts:1992`. The support types, dependency
stubs and drivers are reduced scaffolding. Numeric syntax/flag constants use
6.0.3 values. They are ordinary constant objects so no alternate enum/namespace
Node runner is required. No source function's statements were rewritten.

| Fixture | Usage represented | Stage0 |
| --- | --- | --- |
| 01 map_call | Generic Map.get call unwrap; hit and miss | Refused |
| 02 code_point_call | Scanner codePointAt call; BMP and supplementary input | Refused |
| 03 exports_field | Required symbol exports field before lookup | Refused |
| 04 field_then_call | Optional token spelling call before length | Refused |
| 05 regex_call_index | Regex exec unwrap, followed by indexed capture | Refused |
| 06 memoize_clear | Deliberate undefined unwrap used to clear state | Refused |
| 07 indexed_operation | Generator operation/argument arrays, including indexed unwrap | Refused |
| 08 optional_start | Scanner's optional start unwrap in setText | Refused |
| 09 interface_kind | Node interface to ObjectLiteralExpression by kind | Refused |
| 10 union_kind | Closed support union to ArrayLiteralExpression by kind | Compiles |
| 11 parenthesized_kind | Repeated interface kind downcast to peel parentheses | Refused |
| 12 union_target | Node interface to two access-expression interfaces | Refused |
| 13 flag_downcast | TransientSymbol selected by flag bit | Refused |
| 14 structural_cache | Untagged Type to optional iteration-cache interface | NotYet |
| 15 mutable_view | Generic readonly text range asserted writable | Refused |
| 16 unknown_chain | Polling option conversion through unknown | NotYet |
| 17 brand_upcast | Branded escaped name read as string | NotYet |
| 18 const_tuple | Real readonly signature-cache tuple | Compiles |
| 19 identifier_kind | Real Identifier and property-access switch | NotYet |
| 20 type_flag_mask | Type to ObjectFlagsType by flag mask | Refused |

Fixture 10 deliberately gives the unchanged function a closed support union;
upstream Node is still an interface. It isolates the available union mechanism
and does not classify the upstream site as a union. Fixture 16 models the two
numeric enum domains with distinct phantom brands, preserving the unchecked
conversion while avoiding unrelated enum syntax. Dependency stubs in 07 print
minimal AST text; they do not implement the generator emitter.

All 20 Node observations exit 0 with empty stderr. Stage0: **14 Refused,
4 NotYet, 2 Compiles, 0 Checker**. The two compiled binaries (10 and 18) match
Node's stdout, stderr and exit byte for byte. No silent miscompile was observed.
[status.json](status.json) is exactly the requested array schema; `what` retains
the complete diagnostic text, excluding only go run's `exit status 1` wrapper.
Raw build, Node and native results are in [logs](logs).

Other features can be the first blocker: 13/20 refuse numeric conditions;
14 fails indexed-access lowering; 16 fails its branded optional representation;
17 fails numeric unary lowering; 19 fails a nonconstant case. These are observed
outcomes on main, not predictions about the feature implementations in flight.

## Verification and mutants

`observe.py` ran `node --disable-warning=ExperimentalWarning oracle/node.mjs
<fixture>` and `go run ./cmd/adamic build <fixture> -o <scratch binary>` for every
row, then ran each successful binary and compared all three outputs. It stops
on a mismatch. The setup log records Go 1.27.1, clang 20.1.8, Node 24.19.0;
Go ready 0s, clang/Node/submodules ready 1s, build cache warm 367s, total 367s,
nproc 5, CPU quota 4.

The required mutant changes only `map.get(key)!` to `map.get(key)` in fixture
01, with no checker or compiler changes. Baseline **Refused** becomes
**Checker**, TS2322: `Type 'V | undefined' is not assignable to type 'V'.`
The complete diagnostic and edit are in [logs/mutant.json](logs/mutant.json).
This proves the checker rejects loss of the unwrap; current main refuses the
baseline, so it does not prove a future runtime nullish check can fail.

`audit.py --mutants` checks exact census location coverage, six exclusive counts,
all non-null counts, all 20 unchanged upstream spans, status schema and compiled
output equality. Removing one assertion row is caught by location coverage;
increasing the upcast count by one is caught by category recounting. These are
ledger mutants. The kind-rewrite probe separately falsifies a prospective
unconditional kind-only proof.

Reproduction from the Adamic root, after `cloud/setup.sh` and sourcing its env:

```sh
mkdir -p /tmp/assertions-work
# Extract these two census files without merging another worker's territory.
git show 429c1177f0130f785c19cf590d1860513b2ddbfc:stage3/census/data/sites.json > /tmp/assertions-work/sites.json
git show 429c1177f0130f785c19cf590d1860513b2ddbfc:stage3/census/data/files.json > /tmp/assertions-work/files.json
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/assertions-work/typescript > /tmp/assertions-clone.log 2>&1
npm ci --prefix stage3/fixtures/assertions/api --no-audit --no-fund > /tmp/assertions-npm.log 2>&1
ln -s "$PWD/stage3/fixtures/assertions/api/node_modules" /tmp/assertions-work/typescript/node_modules
node /tmp/assertions-work/typescript/scripts/processDiagnosticMessages.mjs /tmp/assertions-work/typescript/src/compiler/diagnosticMessages.json > /tmp/assertions-generate.log 2>&1
export NODE_PATH="$PWD/stage3/fixtures/assertions/api/node_modules"
node stage3/fixtures/assertions/ledger.cjs /tmp/assertions-work/typescript /tmp/assertions-work > /tmp/assertions-ledger.log 2>&1
node stage3/fixtures/assertions/make-fixtures.cjs /tmp/assertions-work/typescript > /tmp/assertions-extract.log 2>&1
python3 stage3/fixtures/assertions/observe.py > /tmp/assertions-observe.log 2>&1
python3 stage3/fixtures/assertions/audit.py /tmp/assertions-work /tmp/assertions-work/typescript --mutants > /tmp/assertions-audit.log 2>&1
node stage3/fixtures/assertions/kind-soundness.cjs > /tmp/assertions-kind.log 2>&1
go test ./cmd/adamic ./internal/load -count=1 > /tmp/assertions-package-tests.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(casts|cast_fails|unions)\.a$' -count=1 -timeout 30m -v > /tmp/assertions-oracle.log 2>&1
```

Package checks passed: `cmd/adamic` has no test files; `internal/load` passed in
2.359s. The filtered existing oracle ran **casts.a, cast_fails.a, unions.a**,
all passing, total 17.170s (six native and six Node cache misses). It includes
the existing rejected-tag panic case; it does not substitute for a `!` runtime
mutant. Full logs are committed. `git diff --check` also passed.

The committed api lockfile pins the same five npm packages as the observed run;
that run installed them in `/tmp/assertions-work/api`. The scripts accept any
pinned checkout/cache paths. `ledger.cjs` may take a fourth argument for an
alternate output directory. No npm dependency is needed by the `.a` fixtures.

Not covered: native `!` implementation or panic wording, a runtime check mutant
for `!` (main has no such check), whole-program kind/construction proofs,
source adaptations for 111 undefined unwraps, all 4,101 assertions as runnable
fixtures, full generator emission, or the full uncached repository/upstream
TypeScript gates. No compiler implementation, shared fixtures test, adaptation,
other bucket, or PR was changed or created.

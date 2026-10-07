# Type predicates and assertion functions

Pinned source: TypeScript 6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`.
Stage0 base: `ef3d907ecdc4c771b016f7d9c52372def057a340`, current main when this work started.
Census input: `origin/codex/tsc-census:stage3/census/data/sites.json`.
The exact census reason is `a type predicate`.

## What is observed

The census has 651 predicate syntax nodes, on 636 distinct start lines in 31 files.
These are not 651 executable functions. Resolving overload signatures to their
implementation yields 590 distinct bodies, 603 nodes with a body, and 48 nodes
without an implementation body. The latter include callback contracts and
interface/function-type declarations. All 16 `asserts` syntax nodes are in
`debug.ts`; overloads share bodies. Counts below count census nodes, retaining
that distinction rather than silently dropping contracts or counting overloads
as separate implementations.

[LEDGER.md](LEDGER.md) lists every location, class and implementation location.
[ledger.json](ledger.json) additionally retains each body, predicate spelling,
resolved predicate calls, all recognized features, and direct kind-chain status.

| Primary body class | Nodes | Files | Can the body prove the contract? |
| --- | ---: | ---: | --- |
| `kind ===`, possibly a chain | 345 | 13 | Yes for a closed discriminated input union with matching target; the original open `Node` interface alone is insufficient. |
| `typeof` | 8 | 5 | Yes for primitive tests with matching target; compound structural validators require all target properties checked. |
| `instanceof` | 0 | 0 | In principle, with a proven constructor/nominal class relationship; no census body uses it. |
| Call to another predicate | 136 | 16 | Compose proven callee contracts and argument identity; higher-order calls require a proven supplied predicate. |
| Flags mask | 24 | 6 | The mask itself proves no structural subtype. Other conjuncts may independently prove a target. |
| `Debug.assert`-style assertions | 16 | 1 | Only if every normal return implies the assertion and failure cannot return; disabled checks and empty bodies fail. |
| Something else | 74 | 16 | Case-specific control-flow/structural proof; includes kind switches, null checks, presence checks and array algorithms. |
| No implementation body | 48 | 6 | A contract creates a proof obligation for its implementations; no local body can discharge it. |

Classes are exclusive, but bodies can contain multiple features. Priority is:
assertion, instanceof, typeof, mask, kind equality, resolved predicate call,
other. `features` preserves overlaps. A kind switch is “something else” unless
another higher-priority feature appears. Calls are recognized through stock
TypeScript's resolved signatures, including inferred predicates, rather than
assuming every function whose name starts with `is` is a predicate. Nested
function bodies are classified separately. This is a reproducible body-shape
classification, not a claim that stock tsc verifies these predicate contracts.

## Exact proof for the bulk kind tests

All **227** predicates in `factory/nodeTests.ts` are a single direct kind equality.
There are **246** direct equality/OR chains in the whole census. Of the 103 nodes
in `utilities.ts`, four are direct chains; its remaining classes are 30 kind
comparison, 46 predicate call, 22 other, two masks, two typeof and one no-body
(the four chains are included in the 30 comparison count).

For a closed union `U`, let `K(T)` be the set of literal kinds of target members
and let `S` be the kinds tested by the return expression. The positive proof is:
`u in U && u.kind in S` implies `u in T`. An OR chain joins the true-path members;
a switch joins the selected case paths. To support TypeScript's negative-branch
exclusion too, false must imply `u not in T`; equivalently the target's members
within U must be exactly the members selected by S. Every return path must obey
these implications. A stricter boolean filter can satisfy the positive implication
while failing the negative one. Predicate names and declared returns are not proof.

**The original interface is not that closed union.** `types.ts:942` declares
`Node.kind: SyntaxKind`, while `Identifier` at `types.ts:1701` also requires
`escapedText` and inherited members/brands. A kind comparison proves the kind,
not those fields. Consequently the original signature of `isIdentifier` at
`factory/nodeTests.ts:318` does not establish the whole structural target from
its body alone. An implementation needs either a validated, closed node variant
representation whose construction preserves kind/shape correspondence, or
structural validation. A name-based SyntaxKind-to-interface whitelist without
that construction invariant would accept a lie. This unit supplies neither
compiler implementation nor that source/type adaptation. Under the given
body-proof ruling, such unproven original contracts must remain refused.

`isTransientSymbol` at `utilities.ts:655` illustrates the flag problem:
`(symbol.flags & SymbolFlags.Transient) !== 0` does not establish `links` required
by `TransientSymbol` (`types.ts:6139`). Similarly checker TypeFlags/ObjectFlags
and debug FlowFlags are masks, not structural certificates. Mixed cases such
as `isJSDocTypeReference` retain their kind/predicate features in the ledger;
the mask is not automatically a reason to discard an independently valid proof.

`Debug.assert` and `assertIsDefined` can prove truthiness and non-null presence
respectively if `fail` is proven `never`. The overload implementation for
`assertEachIsDefined` needs a loop invariant, sound element presence and a
verified callee. `assertEachNode`, `assertNode`, `assertNotNode`,
`assertOptionalNode`, `assertOptionalToken`, and `assertMissingNode` condition
checking on `shouldAssertFunction`; the disabled path returns with no evidence.
`Debug.type<T>` at `debug.ts:365` has an empty implementation at line 366.
Those bodies cannot discharge their assertions for arbitrary accepted inputs.

## Fixtures

Each extracted function and its statements are kept verbatim, with CRLF converted
to LF. Imports/namespaces are omitted, and support declarations retain only needed
fields. SyntaxKind and SymbolFlags support objects use the exact stock numeric
values; they are driver support, not edited function bodies or enum fixtures.
The ordinary Node runner on main is used unchanged; no enum/namespace runner
change is needed. `Node` support stays an open `kind: number` interface; it is
not strengthened into a closed union to manufacture a proof.

| Fixture | Usage represented | Real extracted caller |
| --- | --- | --- |
| 01_identifier | Single generated kind test; wrong-kind mutant | `isEffectiveModuleDeclaration` |
| 02_module_name | Predicate calls combined by OR | `isModuleName` |
| 03_void_zero | Kind proof followed by nested literal access | `isVoidZero` |
| 04_literal_or | OR kind comparison with a Node/FileReference cast and predicate composition | `isStringOrNumericLiteralLike` |
| 05_kind_switch | Four successful kind cases and default false | Driver only |
| 06_signed_numeric | Harder intersection target: outer kind, operator OR and nested operand predicate | `isSignedNumericLiteral` |
| 07_typeof | Primitive predicate and negative branch narrowing | `getTypeReferenceResolutionName` |
| 08_flags | Flags mask accepts a structurally incomplete Symbol | Driver only |
| 09_assert_defined | Generic assertion, never failure and wrapper returning narrowed value | `checkDefined` |
| 10_nullish | Null/undefined OR with upstream explicit any retained | Driver only |
| 11_empty_assert | Unproved generic assertion returns normally; claimed property is undefined | Driver only |

For 09, `fail` is a small harness dependency throwing `Error`, not a copy of the
full debugging/stack instrumentation. Successful caller behavior and caught
failure text are observed; this does not certify the upstream implementation of
`fail`. Fixture 11 is intentionally unproved. Fixture 08's accepted flags input
intentionally lacks the target subtype's links. These must be refused by a
future body verifier, even though Node correctly runs their original statements.

All eleven Node runs exit 0 with empty stderr. All eleven current-main Stage0
runs are `Refused`; the exact diagnostics, including locations and the `go run`
exit wrapper, are in `status.json`. No fixture compiled, so there is no native
output comparison and no observed silent miscompile. The filtered smoke check
`go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/functions\.a$' -count=1 -timeout 30m -v`
passed both matched fixtures (`functions.a` and `generic_functions.a`) in 9.172s;
its log is retained. Logs retain stdout/stderr
from both runs. Setup: Go 1.27.1, clang 20.1.8, Node 24.19.0; go/clang/node/
submodules 0s each, cache warm 100s, total 100s, nproc 5, CPU quota 4.

## Reproduction and mutants

Use stock TypeScript 6.0.3, the pinned upstream tree and the census `sites.json`.
From the repository root:

```sh
source /workspace/adamic-tools/env.sh
PREDICATES_TYPESCRIPT=/path/to/typescript/lib/typescript.js node stage3/fixtures/predicates/ledger.cjs /path/to/TypeScript /path/to/sites.json /tmp/predicates-ledger.json > /tmp/predicates-ledger.log 2>&1
PREDICATES_TYPESCRIPT=/path/to/typescript/lib/typescript.js node stage3/fixtures/predicates/extract.cjs /path/to/TypeScript > /tmp/predicates-extract.log 2>&1
python3 stage3/fixtures/predicates/record.py > /tmp/predicates-record.log 2>&1
python3 stage3/fixtures/predicates/verify.py /path/to/sites.json --mutants > /tmp/predicates-verify.log 2>&1
node --disable-warning=ExperimentalWarning oracle/node.mjs stage3/fixtures/predicates/01_identifier.a > /tmp/predicates-node.log 2>&1
go run ./cmd/adamic build stage3/fixtures/predicates/01_identifier.a -o /tmp/predicates-native > /tmp/predicates-stage0.log 2>&1
```

The wrong-kind mutant changes only `isIdentifier`'s comparison from Identifier
to StringLiteral. Its real caller, `isEffectiveModuleDeclaration`, changes stdout
from `true\nfalse\ntrue\n` to `false\ntrue\ntrue\n`, with exit 0 and empty stderr
on both runs. Exact caller stdout comparison kills it. `mutant.json` records both
observations. A second mutant removes one ledger entry; the census completeness
check kills it. Neither is a compiler mutant or a native body-verifier test.

## What remains incomplete

This delivers the full census location/body-shape ledger and class-level proof
requirements. It does **not** semantically verify all 590 distinct bodies or
report a proven accept/refuse verdict for each original contract. The unresolved
cases include recursive predicates, mutable alias effects, compound structural
validators, flag/type construction invariants and array predicate overloads.
Doing that accurately needs the body verifier or a separate semantic audit;
syntactic shapes and stock tsc's trusting checker cannot substitute for it.

Four fixtures have driver callers instead of extracted compiler callers; their
full callers require wider checker/transformer/config dependency slices. No real
instanceof-predicate fixture was fabricated because the census has zero such
bodies. Assertion-disable behavior is explained from the actual body ledger,
but has no separate runnable fixture. No upstream adaptation, production compiler
change, shared fixture harness, full integration gate, or upstream TypeScript
suite run is included. Only this bucket and the missing shared NOTICE are added.

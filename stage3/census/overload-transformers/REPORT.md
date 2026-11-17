# Transformer overloads — original census decision, #ev14ch9

Compiler subsequently accepted getProperties under the delegation. The current-main landing and fresh proof receipts are in [landing-932a6344/REPORT.md](landing-932a6344/REPORT.md). The original stop below records the correctly enforced pre-ruling gate; the two namedEvaluation sets remain stopped.

Base: cf468df0747075e50d326a97ff03e8530d7a1a88. TypeScript v6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.

**All three sets stop under road (iii). No landing branch, adaptation or ledger entry qualifies.** The getProperties type-only candidate removes the overload refusal but exposes a new unchecked-cast refusal, failing the required 0-new gate. The two namedEvaluation sets need node-kind precision that the available declaration/body proofs cannot establish.

| Set | Result sites in brief | Checker-resolved callers | Final road | Landing branch / SHA |
| --- | ---: | ---: | --- | --- |
| getProperties | 1 | 2 | (iii), attempted (i) fails no-new gate | none / none |
| injectClassNamedEvaluationHelperBlockIfMissing | 1 | 3 | (iii), kind-preserving result needs a ruling | none / none |
| transformNamedEvaluation | 1 | 23 | (iii), kind-specific assignment results need a ruling | none / none |

The evidence branch is typescript/overload-transformers-evidence. Its first report was pushed as 0bd4986bf before constructing any adapter; INITIAL_REPORT.md preserves that frozen proposal. Checkpoints 4fcba3c9f and 3abfb80bc preserve the controls/tooling and rejection respectively. No force push was used. Rejected source and whole-owner pins are retained under rejected-get-properties, outside stage3/adapt. Production lowering, API sanctions, worker counts and lane expectations were not edited.

## Proof receipts for the rejected getProperties candidate

| Required proof | Status | Observation |
| --- | --- | --- |
| Type-only JavaScript | PASS | Commented and comment-free transpiled owner-file output is byte-identical; LF/CRLF, drift, duplicates, half-adapted text, mixed endings, exact restore and zero-write replay controls pass |
| scan-all, owner file | NOT PROVEN | 5,070 events on both sides, 0 new/gone/changed. Both stop at core.ts:261:1, overload 1 of filter, before this set is reached; Completed=false. This is not a target disappearance pass |
| Full latent, utilities.ts | FAIL | 146 -> 146 events; 6 new, 6 gone, 0 changed semantic events. Three new Refused replays at one cast site, plus three Boundary records, replace the overload events |
| Full latent, ts.ts caller | PASS | 660 -> 656 events; 0 new, 4 gone, 0 changed |
| Full latent, classFields.ts caller | FAIL | 633 -> 624 events; 2 new, 11 gone, 1 changed |
| Restoring mutant | PASS | Exact baseline findings restored: 146 events, including all three overload-refusal replays; 0 delta versus baseline |
| Stock TypeScript 6.0.3 | PASS | Complete sorted 79 roots, locked Node declarations: 0 -> 0 diagnostics, 0 new |
| adamic types | PASS | Complete sorted 79 candidate roots, exit 0 and no diagnostics |
| Candidate stage3/apply.sh | NOT RUN | Gate failed before a candidate product was authorized. Baseline apply passed (997.83 s); direct candidate adapter and replay passed |
| Upstream npm ci / hereby dts | PASS | Unmodified and candidate builds pass; both public bundles byte-identical |
| TestProduct_Stage3AdaptedTree | NOT RUN | Rejected at the no-new-findings gate |
| Whole unfiltered stage3/lane | NOT RUN | Rejected at the no-new-findings gate |
| Ledger | NOT RUN | No qualifying landing change |
| Road (ii) counter / Node witnesses | NOT RUN | No JavaScript-changing road built |

Every full latent run uses LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE set to the measured absolute path, and CommonJSBundle. All runs finish with a relation_report trailer and no output-guard failure. Headers report zero checker diagnostics. The tool's fixed measurement label still says "measured on a checker-rejected program"; the headers, rather than that generic label, establish that these inputs are checker-clean.

For each of the two namedEvaluation stops, scan-all, full candidate latent census, restoring mutant, stock candidate check, adamic types, candidate apply, public declaration comparison, adapted-tree product, whole lane and ledger are NOT RUN: neither set has a candidate edit. No proof from the getProperties experiment is claimed for them.

## Why the getProperties proposal cannot land

Widening only the true overload removes its unneeded initialized-element claim. Full measurement then reaches the existing `as PropertyDeclaration[]` return at transformers/utilities.ts:590:12 and refuses `a cast the runtime can't check`. That is a new finding relative to the main census, even though the text of that cast already existed. It cannot be waved away as a pre-existing latent issue: the required gate counts newly exposed findings too.

Dropping the narrow overload instead exposes the same return cast. Narrowing the implementation result to initialized properties would be false for the false caller. A predicate annotation on the existing boolean helper would overclaim: it returns false for some genuine PropertyDeclaration values depending on initializer/static flags. Splitting the filter into multiple traversals would change JavaScript, intermediate allocations and AST reads; no whole-lane value proof was built for that road. Stop and report rather than ship an adaptation with a new refusal.

The classFields.ts caller census also fails independently: it exposes a new NotYet at binder.ts:2258:13 for writing a checked field that may hold undefined/null, plus its Boundary, and changes the Boundary at classFields.ts:2314:13 from an identifier-assignment stop to the existing core.ts filter-overload refusal. All of those records are listed below.

scan-all independently cannot establish this site's disappearance on this base: both import-closure runs abort at the existing filter overload in core.ts. Their identical event totals prove only the recorded scope. Full latent census supplies the actual getProperties before/after evidence.

## Public declaration bytes

- typescript.d.ts: 588452 bytes, SHA-256 `c967e03c8dbb4770f7e2f0b71b5d13593de679a41cc2b60266c4b69f1342a263`, byte-identical before/after.
- tsserverlibrary.d.ts: 866 bytes, SHA-256 `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd`, byte-identical before/after.

## Source and caller reading

The original reading/road analysis follows. Its inventory counts refer to the initial pre-generation checkout. The successful expanded inventory in whole-checkout-callers.json additionally covers JavaScript: after diagnostic generation, 81,470 files, 21,233 TypeScript and 18,210 JavaScript files, 710 src roots, 7 additional candidate roots and 832 program files. It confirms the same 28 direct calls. There are no function-value escapes: target references consist of 28 calls, 7 declarations and 6 imports. Candidate spelling filters only select work; callee symbols and getResolvedSignature declaration identity establish every reported caller. The original exhaustive TypeScript call traversal is retained at 0bd4986bf and upstream-callers.json.

## Read and caller census

Read all of internal/lower/census_small.go, overload_predicate_result.go, census_overload_proof.go and overload_result_check.go before choosing these roads. Exact upstream overloads, full implementations and all resolved call statements/types are in upstream-callers.json. callers.cjs uses stock TypeScript 6.0.3 getResolvedSignature and compares the resolved declaration's identity/location, including aliases. Identifier references are also recorded to detect non-call uses. It inventories the entire checkout: 81,468 files, 21,232 TypeScript files, all 709 src TypeScript roots (including harness and testRunner), and all four additional test/baseline TypeScript files containing a target spelling. Text only selects extra roots; it never identifies a caller. The 807-file program resolves 28 calls to these exact owners; none is in harness, testRunner or tests. Unrelated getProperties methods do not count. Upstream uses CRLF.

Ownership search covered every stage3/adapt owners.json, targets.json and adapt.cjs. Only 99-correctness-receiver-transformers/owners.json matched an exact target identifier, in the esnext.ts import. It pins none of the three function owners. getProperties therefore needs a new house-shape directory, ordered after existing dependencies.

### getProperties

Two overloads, one implementation. One true argument at ts.ts:858 feeds some solely for array nonemptiness (one caller); one false argument at classFields.ts:2312 binds instanceProperties for ordinary PropertyDeclaration processing (one caller). No caller requires InitializedPropertyDeclaration. The true result promises initialized elements, but the implementation's readonly PropertyDeclaration[] does not serve that promise under censusRelated. There is no presence difference, callback predicate or void promise. Its reached return has a cast, which censusReturnProof refuses; narrowOverload never accepts arrays.

Attempted road (i), subsequently rejected by the no-new-findings gate: widen the first overload's result to the already-served readonly PropertyDeclaration[]. Preserve both overloads, all parameters, the body, documentation and runtime values. No public API baseline declares this internal function. Byte-identical commented and comment-free emits will be asserted by the adapter, and the whole source/declaration checks remain required.

### injectClassNamedEvaluationHelperBlockIfMissing

Three callers: namedEvaluation.ts:250 passes a ClassExpression and immediately casts/checks the result as ClassExpression (one); esDecorators.ts:676 passes ClassLikeDeclaration and reassigns its union (one); esDecorators.ts:1062 passes ClassDeclaration and reassigns the ClassDeclaration variable (one). The last needs the narrow result. Body and overload are retained in upstream-callers.json in full.

censusRelated cannot serve Extract<ClassLikeDeclaration, Pick<T, "kind">> from ClassLikeDeclaration for rigid T. No callback/presence/void/phantom acceptance applies. The body proof stops on classHasExplicitlyAssignedName(node), on member-based tests, and on the assignment to node; its final local/parameter flow cannot prove the promised Extract. The implementation is non-generic, but its unresolved generic Extract promise is not an ordinary checkable literal-tag object here (the recorded refusal remains the task's starting evidence).

Candidate (i), dropping the overload or widening its result to ClassLikeDeclaration, loses the ClassDeclaration caller's kind. A generic implementation annotation requires proving the conditional result, not asserting it with a new cast. Candidate (ii), merely branching on node.kind or isClassDeclaration(node), is outside censusTrustedTest's bare-parameter leaves. Adding wrappers, assertions, a new shape or moving the problem into callers needs a ruling. Stop under (iii); no adaptation or ledger line is built.

### transformNamedEvaluation

23 callers, all reassigning the result: classFields.ts has 10, esDecorators.ts 9, esnext.ts 4. By argument kind/use: PropertyAssignment 2, ShorthandPropertyAssignment 2, VariableDeclaration 3, ParameterDeclaration 2, BindingElement 2, PropertyDeclaration 3 (including one private-name case whose Extract instantiates as never), ExportAssignment 2, assignment expressions 7. These are node-specific variables, not discarded results. Each full statement and instantiated result is recorded.

The implementation's inferred union of factory-update results does not serve the rigid Extract<NamedEvaluation, Pick<T, ...>> promise. There is no predicate callback, presence-only difference or void result. The body proof cannot read switch(node.kind), and a generic implementation would disable narrowOverload. The promised Extract includes intersection/name/operator restrictions rather than one ordinary checkable literal-tag node result.

Candidate (i), widening or dropping the overload, loses node-specific views at the assignment callers. Annotating the implementation with the generic Extract would need unproved casts. Candidate (ii), expanding the switch to if tests on node.kind, still fails censusTrustedTest. New type shapes, assertions or specialized wrappers require a ruling. Stop under (iii); no adaptation or ledger line is built.


## Complete finding deltas

Paths below are normalized only by removing their respective measured-tree prefixes. Event identity retains all fields except raw start/end byte offsets; the shorter annotation shifts later spans by 11 bytes. delta-latent-utilities-raw.json preserves all raw metadata differences, including eleven Boundary span shifts with identical site, kind, unit, phase, reason and words. These are not changed semantic findings. No production refusal is omitted. All newly exposed cast events are failures of the gate.

### Full latent utilities

```json
{
  "new": [
    {
      "depth": 0,
      "kind": "Refused",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "a cast the runtime can't check",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:590:12: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
      "unit": "src/compiler/transformers/utilities.ts:585:1",
      "where": "src/compiler/transformers/utilities.ts:590:12"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:590:12: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
      "unit": "src/compiler/transformers/utilities.ts:585:1",
      "where": "src/compiler/transformers/utilities.ts:590:5"
    },
    {
      "depth": 0,
      "kind": "Refused",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "a cast the runtime can't check",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:590:12: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
      "unit": "src/compiler/transformers/utilities.ts:587:1",
      "where": "src/compiler/transformers/utilities.ts:590:12"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:590:12: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
      "unit": "src/compiler/transformers/utilities.ts:587:1",
      "where": "src/compiler/transformers/utilities.ts:590:5"
    },
    {
      "depth": 0,
      "kind": "Refused",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "a cast the runtime can't check",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:590:12: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
      "unit": "src/compiler/transformers/utilities.ts:589:1",
      "where": "src/compiler/transformers/utilities.ts:590:12"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:590:12: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
      "unit": "src/compiler/transformers/utilities.ts:589:1",
      "where": "src/compiler/transformers/utilities.ts:590:5"
    }
  ],
  "gone": [
    {
      "depth": 0,
      "kind": "Refused",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:585:1: Adamic 0.1 refuses overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]; make the implementation result covariant with every overload result",
      "unit": "src/compiler/transformers/utilities.ts:585:1",
      "where": "src/compiler/transformers/utilities.ts:585:1"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "function signature or prologue failed; continued at next declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:585:1: Adamic 0.1 refuses overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]; make the implementation result covariant with every overload result",
      "unit": "src/compiler/transformers/utilities.ts:585:1",
      "where": "src/compiler/transformers/utilities.ts:585:1"
    },
    {
      "depth": 0,
      "kind": "Refused",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:585:1: Adamic 0.1 refuses overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]; make the implementation result covariant with every overload result",
      "unit": "src/compiler/transformers/utilities.ts:587:1",
      "where": "src/compiler/transformers/utilities.ts:585:1"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "function signature or prologue failed; continued at next declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:585:1: Adamic 0.1 refuses overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]; make the implementation result covariant with every overload result",
      "unit": "src/compiler/transformers/utilities.ts:587:1",
      "where": "src/compiler/transformers/utilities.ts:587:1"
    },
    {
      "depth": 0,
      "kind": "Refused",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:585:1: Adamic 0.1 refuses overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]; make the implementation result covariant with every overload result",
      "unit": "src/compiler/transformers/utilities.ts:589:1",
      "where": "src/compiler/transformers/utilities.ts:585:1"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "function signature or prologue failed; continued at next declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:585:1: Adamic 0.1 refuses overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]; make the implementation result covariant with every overload result",
      "unit": "src/compiler/transformers/utilities.ts:589:1",
      "where": "src/compiler/transformers/utilities.ts:589:1"
    }
  ],
  "changed": []
}
```

### Full latent ts

```json
{
  "new": [],
  "gone": [
    {
      "depth": 0,
      "kind": "NotYet",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "an array of InitializedPropertyDeclaration",
      "root_kind": "",
      "text": "src/compiler/core.ts:622:31: stage 0 can't lower an array of InitializedPropertyDeclaration yet",
      "unit": "src/compiler/transformers/ts.ts:855:5",
      "where": "src/compiler/core.ts:622:31"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "root_kind": "",
      "text": "src/compiler/core.ts:622:31: stage 0 can't lower an array of InitializedPropertyDeclaration yet",
      "unit": "src/compiler/transformers/ts.ts:855:5",
      "where": "src/compiler/core.ts:622:17"
    },
    {
      "depth": 0,
      "kind": "Refused",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:585:1: Adamic 0.1 refuses overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]; make the implementation result covariant with every overload result",
      "unit": "src/compiler/transformers/ts.ts:855:5",
      "where": "src/compiler/transformers/utilities.ts:585:1"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:585:1: Adamic 0.1 refuses overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]; make the implementation result covariant with every overload result",
      "unit": "src/compiler/transformers/ts.ts:855:5",
      "where": "src/compiler/transformers/ts.ts:857:9"
    }
  ],
  "changed": []
}
```

### Full latent classFields

```json
{
  "new": [
    {
      "depth": 0,
      "kind": "NotYet",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds",
      "root_kind": "",
      "text": "src/compiler/binder.ts:2258:13: stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/binder.ts:2258:13"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "root_kind": "",
      "text": "src/compiler/binder.ts:2258:13: stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/classFields.ts:2322:13"
    }
  ],
  "gone": [
    {
      "depth": 0,
      "kind": "Refused",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:585:1: Adamic 0.1 refuses overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]; make the implementation result covariant with every overload result",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/utilities.ts:585:1"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/utilities.ts:585:1: Adamic 0.1 refuses overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[]; make the implementation result covariant with every overload result",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/classFields.ts:2311:9"
    },
    {
      "depth": 0,
      "kind": "NotYet",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "reading instanceProperties",
      "root_kind": "",
      "text": "src/compiler/transformers/classFields.ts:2312:26: stage 0 can't lower reading instanceProperties yet",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/classFields.ts:2312:26"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/classFields.ts:2312:26: stage 0 can't lower reading instanceProperties yet",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/classFields.ts:2312:9"
    },
    {
      "depth": 0,
      "kind": "NotYet",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "assigning to an Identifier",
      "root_kind": "",
      "text": "src/compiler/transformers/classFields.ts:2314:13: stage 0 can't lower assigning to an Identifier yet",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/classFields.ts:2314:13"
    },
    {
      "depth": 0,
      "kind": "NotYet",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "reading properties",
      "root_kind": "",
      "text": "src/compiler/transformers/classFields.ts:2318:43: stage 0 can't lower reading properties yet",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/classFields.ts:2318:43"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/classFields.ts:2318:43: stage 0 can't lower reading properties yet",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/classFields.ts:2318:9"
    },
    {
      "depth": 0,
      "kind": "NotYet",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "reading needsConstructorBody",
      "root_kind": "",
      "text": "src/compiler/transformers/classFields.ts:2321:30: stage 0 can't lower reading needsConstructorBody yet",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/classFields.ts:2321:30"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/classFields.ts:2321:30: stage 0 can't lower reading needsConstructorBody yet",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/classFields.ts:2321:9"
    },
    {
      "depth": 0,
      "kind": "NotYet",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "reading properties",
      "root_kind": "",
      "text": "src/compiler/transformers/classFields.ts:2353:76: stage 0 can't lower reading properties yet",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/classFields.ts:2353:76"
    },
    {
      "depth": 0,
      "kind": "Boundary",
      "measurement": "measured on a checker-rejected program",
      "phase": "lowering",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "root_kind": "",
      "text": "src/compiler/transformers/classFields.ts:2353:76: stage 0 can't lower reading properties yet",
      "unit": "src/compiler/transformers/classFields.ts:2310:5",
      "where": "src/compiler/transformers/classFields.ts:2353:13"
    }
  ],
  "changed": [
    {
      "site": [
        "src/compiler/transformers/classFields.ts:2314:13",
        "Boundary"
      ],
      "before": [
        {
          "depth": 0,
          "kind": "Boundary",
          "measurement": "measured on a checker-rejected program",
          "phase": "lowering",
          "reason": "failed statement; state rolled back; continued at next statement or declaration",
          "root_kind": "",
          "text": "src/compiler/transformers/classFields.ts:2314:13: stage 0 can't lower assigning to an Identifier yet",
          "unit": "src/compiler/transformers/classFields.ts:2310:5",
          "where": "src/compiler/transformers/classFields.ts:2314:13"
        }
      ],
      "after": [
        {
          "depth": 0,
          "kind": "Boundary",
          "measurement": "measured on a checker-rejected program",
          "phase": "lowering",
          "reason": "failed statement; state rolled back; continued at next statement or declaration",
          "root_kind": "",
          "text": "src/compiler/core.ts:261:1: Adamic 0.1 refuses overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined; make the implementation result covariant with every overload result",
          "unit": "src/compiler/transformers/classFields.ts:2310:5",
          "where": "src/compiler/transformers/classFields.ts:2314:13"
        }
      ]
    }
  ]
}
```

scan-all: new=[], gone=[], changed=[]; target absent from both recorded scopes. Full streams, exact commands, real child exit codes, source hashes and metadata are in the adjacent JSON/JSONL files and logs/. Restoring-mutant events are in restoring-mutant-result.json; its baseline delta is delta-mutant.json.

## Non-proof attempts retained

The first adapter construction had a Python syntax error; consequent missing-module runs failed. Corrected *-fixed and adapter-controls receipts establish the actual contracts. Expanded caller-inventory attempts hit stock-TypeScript stack overflows in unrelated dependency/fixture exploration; the successful scoped-symbol/signature audit and initial exhaustive inventory agree. The first stock probe omitted Node declarations and is invalid; only the corrected *-with-node pair is used. No failed attempt is reported as a pass. Wrapper receipts preserve each child's own return code. No evidence is placed on a landing branch.

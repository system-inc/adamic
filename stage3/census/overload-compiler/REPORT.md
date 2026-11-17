# Compiler overload sets — #ev14ch9

**Outcome: four sets stopped; zero accepted adaptations and zero landing branches.** No compiler/lowering rule, public declaration, API sanction, oracle, or ledger was changed. `builder.ts:1663 createBuilderProgram` was left alone.

Base: `cf468df0747075e50d326a97ff03e8530d7a1a88`. Upstream v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. [setup.json](setup.json) records the initialized recursive submodule pins, tool versions and scratch paths. The source worktree was fresh and detached at the base before creating the evidence branch. Go temporary files/cache were in scratch; no new Go file was placed under stage3.

The pre-candidate report and per-set counts/roads were pushed first at `6e7c19568`. Baseline source adaptation had started for the caller audit before that checkpoint; no candidate existed until after the checkpoint push. Subsequent evidence checkpoints include `96f4633e0`, `9f75f5b8d` and `33489405f`. The union-array follow-up was predeclared at 33489405f before its construction. No force push was used.

## Per-set decision

| Set, upstream line | Bodyless overloads in adapted baseline | Resolved callers | Final road | Landing branch / SHA |
| --- | ---: | ---: | --- | --- |
| utilities.ts:6122 getDiagnostics | 2 | 11 | (iii): admitted empty string returns global diagnostics without locations | none / none |
| utilities.ts:10011 getSupportedExtensionsWithJsonIfResolveJsonModule | 2 | 4 | (iii): two type-only roads tried; both fail the census gate | none / none |
| builder.ts:558 convertOrRepopulateDiagnosticMessageChainArray | 1 | 3 | (iii): generic type-only amendment fails scan and full latent gates | none / none |
| commandLineParser.ts:3775 convertOptionsFromJson | 2 | 3 | (iii): watch-result promise needs a separately typed option boundary/new shape | none / none |

## Read and count

Read in full at the pinned base: `internal/lower/census_small.go`, `internal/lower/overload_predicate_result.go`, `internal/lower/census_overload_proof.go`, and `internal/lower/overload_result_check.go`. Overload parameters are related contravariantly, with mutable invariance and strict callback variance. Results first try relation, phantom/presence/predicate/void acceptances, then the limited body proof and call checks. A generic implementation excludes the call-check road; arrays and untagged interfaces are not checkable promises. The body proof refuses reached casts, parameter writes, returned locals and untrusted control flow.

[final-callers-decoded.json](final-callers-decoded.json) audits **21,236 TypeScript roots** over the completed adapted checkout, including src/harness, src/testRunner and tests. It reports **zero source changes during its checker walk**. [js-callers.json](js-callers.json) audits **18,289 roots** comprising checkout JavaScript and compiler source declarations, with **zero additional calls** to these source owners. Counts come only from `getResolvedSignature` declarations, never name-text matches. DiagnosticCollection interface signatures are included because they carry the returned nested function's contract. [CALLERS.md](CALLERS.md) lists every resolved caller and its use; the JSON retains full call/argument/context/signature text.

- getDiagnostics: **5 filename calls**, all filename property reads, feeding diagnostic returns/concatenation/directive filtering; **6 no-argument calls**, feeding one return, emitter output, iteration, and three equality assertions. Immediate uses: 6 bindings, 2 arguments, 1 return, 1 property initializer, 1 member read.
- Extensions: **1 readonly enum-array input**, **2 readonly string-array inputs**, **1 mutable string-array input**. Three results are flattened/read for extension tests or directory enumeration; one is returned for completions. No site writes an inner array.
- Builder array: **3 T[] inputs**, all cast from chain.next and paired with the same `(T)` repopulation callback. Two results become next fields; one is compared with chain.next to retain original chain identity. Its only caller file is builder.ts.
- Options: **2 discarded results** with CompilerOptions/TypeAcquisition default objects, retained and mutated by callers; **1 returned watch result**, passing undefined as its default.

The intentionally invalid import.defer fixture makes stock TypeScript throw. The JavaScript pass additionally hits generated import.defer baselines and two deep LKG calls. All **15 exception sites** are explicitly reviewed in [exceptions-reviewed.json](exceptions-reviewed.json): identifier callees bind to declarations in lib, outside the requested source owners; an import meta-property cannot denote one of these named functions. No exception was silently counted as a resolved call. The first aborted audit and preliminary reads while baseline adaptation was running are not final proof. The final audit uses TypeScript's decoder, rather than UTF-8 assumptions, for UTF-16/legacy fixtures.

Ownership search covered owners.json, targets.json and every adapt.cjs before candidate construction. The builder array owner is pinned whole by **99-correctness-axolotl-k**, so its candidate is an amendment there. No competing owner pin was introduced. The extension target has no whole-owner pin; the adjacent getSupportedExtensions pin belongs to a different function. Complete proposed extension union owner text/SHA pins are retained in [extensions-union-owner.json](extensions-union-owner.json), without installing a new adaptation directory.

## Why each stops

### getDiagnostics

`censusOverload` cannot serve mutable DiagnosticWithLocation[] from Diagnostic[]. Phantom, presence, predicate and void roads do not apply; arrays exclude narrowOverload. The map.get call, returned local, unshift and existing casts prevent the body proof.

There is also a real admitted counterexample: `getDiagnostics("")` takes the all-diagnostics branch. [diagnostics-empty-string-witness.json](diagnostics-empty-string-witness.json), using the actual stock 6.0.3 exported collection, records one returned global diagnostic, the original object identity, and no location. Changing truthiness to an undefined test would change that computed value. Dropping/widening the nested overload would no longer satisfy DiagnosticCollection's location-returning method contract. Candidate roads needing a ruling: a nonempty-name contract or a changed internal collection result shape plus consumer narrowing. Neither was built.

### Extensions

The original readonly applies only to the **outer** array. Its mutable Extension[] rows cannot be served by mutable string[] rows. The narrower result promise also cannot be proven from a string-array implementation result; arrays exclude call checking.

First road (i): drop the enum overload. JavaScript is byte-identical, and the stock checker accepts all 79 roots. **Scan-all: 5074 → 5075 events, 0 gone, 1 new, 0 changed.** The new event is program.ts:1588:119, `stringWideningRefusal`: `a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read`.

Second road (i), predeclared before construction: one signature taking/returning `readonly (Extension[] | string[])[]`, retaining the existing inner views as union members. JavaScript is byte-identical; the stock API checker and actual tsc CLI both pass. **Scan-all: 5074 → 5075, 0 gone, 1 new, 2 changed.** The new program.ts refusal remains, now against the union-array view; two builtin-return refusals change words. Full tuples are below. The baseline scan stops before it can prove the declaration site gone, so this is not an elimination claim.

Remaining candidate roads would need a new nested readonly result view propagated into the explicitly typed service helper, or a separately typed boundary. Those cross the brief's stop-for-new-shape/ruling boundary and were not built. Copying rows when options are off would violate argument identity and is not proposed.

### Builder array

The adapted overload admits arbitrary T[] and a matching `(T)` callback; the nongeneric implementation receives a mutable base-union array and base-union callback. This loses T and fails the parameter relation.

The road (i) candidate amends 99-correctness-axolotl-k: one generic implementation, a generic subtype-preserving local predicate, and separate sameMap input/output type arguments. [builder-candidate.diff](builder-candidate.diff), [builder-candidate-owners.json](builder-candidate-owners.json) and [builder-mutant.cjs.txt](builder-mutant.cjs.txt) retain the complete reviewable attempt. The owner after pin/SHA/declaration count were updated, previous after text retained for byte-identical JavaScript assertions, mixed line endings refused, and existing guard/copy/identity behavior retained.

The stock checker passes **0 diagnostics / 79 roots**. Both emitted-JavaScript comparisons pass. All **13 existing refusal controls**, unrelated-edit acceptance, and **zero-write replay** pass. Adamic types passes all 79 roots. These do not override the census failure:

- Scan-all: **5070 → 5067, 6 gone, 3 new, 0 changed**.
- Full single-file latent census, CommonJSBundle with LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1: **453 → 449, 11 gone, 7 new, 0 changed**. This covers owner and every resolved caller because all are in builder.ts. The overload refusal is gone, but mutable recursive-chain views, an unproven predicate, a generic function value, and newly reached runtime-uncheckable casts fail the gate.
- Restoring mutant: **453 → 453, 0 gone, 0 new, 0 changed** against baseline, including the original parameter-array overload refusal. The mutant's JavaScript equality assertion passes.

Changing the predicate passing convention alone would not settle the recursive T/next casts and mutable views. Further roads need a different recursive contract/new shape or a ruling. No amendment is committed to a landing branch.

### convertOptionsFromJson

The watch overload promises WatchOptions | undefined; the implementation can return a CompilerOptions or TypeAcquisition object as well. The other two callers discard the result. WatchOptions is untagged and cannot use narrowOverload; TRaw also makes the implementation generic. The body assigns defaultOptions, iterates and returns that parameter, so the body proof cannot establish the watch promise. The earlier adaptation already adds undefined to the non-watch overload's early-return promise.

Dropping overloads exposes the broad result at the watch worker; a cast would conceal the same unsupported claim. A separately typed accumulator/helper or a different internal option shape would require the ruling boundary specified in the brief. No candidate/public declaration edit was built.

## Proof receipts

| Required line | Diagnostics | Extensions | Builder array | JSON options |
| --- | --- | --- | --- | --- |
| scan-all over owner file | not run | **failed**, both roads; full differences below | **failed**, 3 new | not run |
| Full latent owner + every caller file | not run | **not completed**: owner/program jobs stopped after scan failures; commandLineParser/stringCompletions not run | **failed**, 7 new; complete owner/caller census | not run |
| Restoring mutant brings site back | not run | not run | **passed**, exact baseline census restored | not run |
| Stock 6.0.3 over 79 roots | not run | **passed**, 0 diagnostics, both roads; union CLI exit 0 | **passed**, 0 diagnostics | not run |
| adamic types over 79 roots | not run | **passed first road**; union road not run | **passed**, exit 0, empty diagnostic stderr | not run |
| stage3/apply.sh candidate | not run | not run | not run; owner contract passed | not run |
| npm ci + hereby dts, both public files identical | not run | not run | not run | not run |
| TestProduct_Stage3AdaptedTree | not run | not run | not run | not run |
| Whole unfiltered lane | not run | not run | not run | not run |
| Landing ledger line | not written | not written | not written | not written |

Baseline stage3/apply.sh **passed**; [baseline-apply.stdout](baseline-apply.stdout) retains its full output. This is not a candidate apply or product-test pass. The rejected early 77-root trial files predate generated inputs and contain preparation diagnostics; only the completed-tree 79-root stock receipts are acceptance checks. Product/dts/lane/ledger work was not started after failed census gates.

Four remaining extension census processes were deliberately stopped after both road failures. Their `.run.json` receipts record direct child return code **-15 (SIGTERM)**; the JSONL streams contain only an initial header and **no completed census/trailer**, so they are not evidence of zero findings. [stopped-measurements.json](stopped-measurements.json) records each exact process and reason. The comparator refuses incomplete latent streams. The original wrapper exposed a signalled child's negative return code modulo 256; it now re-raises child signals after recording the direct code, and [wrapper-signal-control.run.json](wrapper-signal-control.run.json) plus the observed wrapper exit 143 verify SIGTERM propagation. Normal measurement child exit codes are retained and propagated; no interrupted job is labelled passed.

Scan-all is a refusal-scan continuation plus the first declaration/body/tail stop, not complete lowering acceptance. All completed scan runs stop at a production declaration boundary. Full builder census has its no-output assertions enabled and CommonJSBundle in each header. Its checker-rejected label is the tool's static warning; the actual header records checker_rejected=false and zero diagnostics.

Comparison maps unchanged lines through the source diff and exact function-unit starts by unique named AST ancestry. Event multiplicity is retained. Changed words at the same rule/site are explicitly paired. Raw streams retain every metadata field; the following complete diff tuples list every new, gone and changed finding, without suppressing newly reached failures.

## Complete census differences

### scan-extensions-diff.json

```json
{
  "mode": "scan",
  "baseline_events": 5074,
  "candidate_events": 5075,
  "gone": [],
  "new": [
    {
      "count": 1,
      "event": [
        "scan stop",
        "refusal_scan/stringWideningRefusal",
        "src/compiler/program.ts:1588:119",
        "a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read"
      ]
    }
  ],
  "changed": [],
  "comparison": "Unchanged source lines are mapped by SequenceMatcher. Changed words at an equal rule/site are paired and listed in full; no changed words are suppressed."
}
```

### scan-extensions-union-diff.json

```json
{
  "mode": "scan",
  "baseline_events": 5074,
  "candidate_events": 5075,
  "gone": [],
  "new": [
    {
      "count": 1,
      "event": [
        "scan stop",
        "refusal_scan/stringWideningRefusal",
        "src/compiler/program.ts:1588:119",
        "a value of type readonly Extension[][] seen as readonly (string[] | Extension[])[], which can write string where Extension is read"
      ]
    }
  ],
  "changed": [
    {
      "count": 1,
      "before": [
        "scan stop",
        "refusal_scan/stringWideningRefusal",
        "src/compiler/utilities.ts:10019:64",
        "a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read"
      ],
      "after": [
        "scan stop",
        "refusal_scan/stringWideningRefusal",
        "src/compiler/utilities.ts:10019:64",
        "a value of type readonly Extension[][] seen as readonly (string[] | Extension[])[], which can write string where Extension is read"
      ]
    },
    {
      "count": 1,
      "before": [
        "scan stop",
        "refusal_scan/stringWideningRefusal",
        "src/compiler/utilities.ts:10020:63",
        "a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read"
      ],
      "after": [
        "scan stop",
        "refusal_scan/stringWideningRefusal",
        "src/compiler/utilities.ts:10020:63",
        "a value of type readonly Extension[][] seen as readonly (string[] | Extension[])[], which can write string where Extension is read"
      ]
    }
  ],
  "comparison": "Unchanged source lines are mapped by SequenceMatcher; exact function-unit starts are matched by unique named function ancestry in the TypeScript AST. Changed words at an equal rule/site are paired and listed in full; no changed words are suppressed."
}
```

### scan-builder-diff.json

```json
{
  "mode": "scan",
  "baseline_events": 5070,
  "candidate_events": 5067,
  "gone": [
    {
      "count": 1,
      "event": [
        "scan stop",
        "refusal_scan/constraintRelationRefusal",
        "src/compiler/builder.ts:580:106",
        "a value of type DiagnosticMessageChain seen as DiagnosticMessageChain | ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] | undefined where DiagnosticMessageChain[] | undefined is read"
      ]
    },
    {
      "count": 1,
      "event": [
        "scan stop",
        "refusal_scan/enumRefusal",
        "src/compiler/builder.ts:585:56",
        "an object refinement using an open numeric enum as a literal tag"
      ]
    },
    {
      "count": 1,
      "event": [
        "scan stop",
        "refusal_scan/enumRefusal",
        "src/compiler/builder.ts:585:75",
        "an object refinement using an open numeric enum as a literal tag"
      ]
    },
    {
      "count": 1,
      "event": [
        "scan stop",
        "refusal_scan/enumRefusal",
        "src/compiler/builder.ts:586:17",
        "an object refinement using an open numeric enum as a literal tag"
      ]
    },
    {
      "count": 1,
      "event": [
        "scan stop",
        "refusal_scan/enumRefusal",
        "src/compiler/builder.ts:586:45",
        "an object refinement using an open numeric enum as a literal tag"
      ]
    },
    {
      "count": 1,
      "event": [
        "scan stop",
        "refusal_scan/stringWideningRefusal",
        "src/compiler/builder.ts:580:137",
        "a value of type (chain: DiagnosticMessageChain | ReusableDiagnosticMessageChain) => RepopulateDiagnosticChainInfo | undefined seen as (chain: DiagnosticMessageChain) => RepopulateDiagnosticChainInfo | undefined, which can write ReusableDiagnosticMessageChain[] | undefined where DiagnosticMessageChain[] | undefined is read"
      ]
    }
  ],
  "new": [
    {
      "count": 1,
      "event": [
        "scan stop",
        "refusal_scan/constraintRelationRefusal",
        "src/compiler/builder.ts:574:134",
        "a value of type T & DiagnosticMessageChain seen as DiagnosticMessageChain | ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] | undefined where DiagnosticMessageChain[] | (ReusableDiagnosticMessageChain[] & DiagnosticMessageChain[]) | undefined is read"
      ]
    },
    {
      "count": 1,
      "event": [
        "scan stop",
        "refusal_scan/predicateArguments",
        "src/compiler/builder.ts:579:21",
        "an unproven predicate argument for parameter predicate (argument \"isFullChain\")"
      ]
    },
    {
      "count": 1,
      "event": [
        "scan stop",
        "refusal_scan/predicateArgumentsRefusal",
        "src/compiler/builder.ts:579:21",
        "an unproven predicate argument for parameter predicate (argument \"isFullChain\")"
      ]
    }
  ],
  "changed": [],
  "comparison": "Unchanged source lines are mapped by SequenceMatcher. Changed words at an equal rule/site are paired and listed in full; no changed words are suppressed."
}
```

### latent-builder-diff.json

```json
{
  "mode": "latent",
  "baseline_events": 453,
  "candidate_events": 449,
  "gone": [
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:572:1",
        "lowering",
        "Boundary",
        "src/compiler/builder.ts:572:1",
        "function signature or prologue failed; continued at next declaration"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:572:1",
        "lowering",
        "Refused",
        "src/compiler/builder.ts:567:5",
        "overload 1 of convertOrRepopulateDiagnosticMessageChainArray parameter array cannot be served by implementation parameter array"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:572:1",
        "refusal_scan",
        "Refused",
        "src/compiler/builder.ts:580:106",
        "a value of type DiagnosticMessageChain seen as DiagnosticMessageChain | ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] | undefined where DiagnosticMessageChain[] | undefined is read"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:572:1",
        "refusal_scan",
        "Refused",
        "src/compiler/builder.ts:580:137",
        "a value of type (chain: DiagnosticMessageChain | ReusableDiagnosticMessageChain) => RepopulateDiagnosticChainInfo | undefined seen as (chain: DiagnosticMessageChain) => RepopulateDiagnosticChainInfo | undefined, which can write ReusableDiagnosticMessageChain[] | undefined where DiagnosticMessageChain[] | undefined is read"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:584:5",
        "lowering",
        "Boundary",
        "src/compiler/builder.ts:585:9",
        "failed statement; state rolled back; continued at next statement or declaration"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:584:5",
        "lowering",
        "Refused",
        "src/compiler/builder.ts:585:16",
        "in that narrows a union of object types (any object with the key passes)"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:584:5",
        "refusal_scan",
        "Refused",
        "src/compiler/builder.ts:585:56",
        "an object refinement using an open numeric enum as a literal tag"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:584:5",
        "refusal_scan",
        "Refused",
        "src/compiler/builder.ts:585:75",
        "an object refinement using an open numeric enum as a literal tag"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:584:5",
        "refusal_scan",
        "Refused",
        "src/compiler/builder.ts:586:17",
        "an object refinement using an open numeric enum as a literal tag"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:584:5",
        "refusal_scan",
        "Refused",
        "src/compiler/builder.ts:586:45",
        "an object refinement using an open numeric enum as a literal tag"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:618:1",
        "lowering",
        "Refused",
        "src/compiler/builder.ts:567:5",
        "overload 1 of convertOrRepopulateDiagnosticMessageChainArray parameter array cannot be served by implementation parameter array"
      ]
    }
  ],
  "new": [
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:572:1",
        "refusal_scan",
        "Refused",
        "src/compiler/builder.ts:574:134",
        "a value of type T & DiagnosticMessageChain seen as DiagnosticMessageChain | ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] | undefined where DiagnosticMessageChain[] | (ReusableDiagnosticMessageChain[] & DiagnosticMessageChain[]) | undefined is read"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:572:1",
        "refusal_scan",
        "Refused",
        "src/compiler/builder.ts:579:21",
        "an unproven predicate argument for parameter predicate (argument \"isFullChain\")"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:618:1",
        "lowering",
        "Boundary",
        "src/compiler/builder.ts:579:5",
        "failed statement; state rolled back; continued at next statement or declaration"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:618:1",
        "lowering",
        "NotYet",
        "src/compiler/builder.ts:579:21",
        "a generic function as a value"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:618:1",
        "lowering",
        "Refused",
        "src/compiler/builder.ts:544:66",
        "a cast the runtime can't check"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:618:1",
        "lowering",
        "Refused",
        "src/compiler/builder.ts:550:66",
        "a cast the runtime can't check"
      ]
    },
    {
      "count": 1,
      "event": [
        "",
        "src/compiler/builder.ts:618:1",
        "lowering",
        "Refused",
        "src/compiler/builder.ts:553:65",
        "a cast the runtime can't check"
      ]
    }
  ],
  "changed": [],
  "comparison": "Unchanged source lines are mapped by SequenceMatcher; exact function-unit starts are matched by unique named function ancestry in the TypeScript AST. Changed words at an equal rule/site are paired and listed in full; no changed words are suppressed."
}
```

### latent-builder-mutant-diff.json

```json
{
  "mode": "latent",
  "baseline_events": 453,
  "candidate_events": 453,
  "gone": [],
  "new": [],
  "changed": [],
  "comparison": "Unchanged source lines are mapped by SequenceMatcher. Changed words at an equal rule/site are paired and listed in full; no changed words are suppressed."
}
```

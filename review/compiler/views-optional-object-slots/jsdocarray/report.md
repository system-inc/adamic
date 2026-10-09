# JSDocArray contract boundary

Roadmap step 09, #tzd3gjg. Evidence-only unit on compiler/views-optional-object-slots at 390cdc2905b66c915deec4068ded3c365d184d67, with V3 14690251f29b6f155c5d795be13eb420de00471a. No compiler admission is added. The user directed this unit to stop rather than implement V4 callables or V5 dictionaries/intersections.

## Exact contract

The hash-pinned adapted declaration is types.ts:963:1:

```text
interface JSDocArray extends Array<JSDoc> {
    jsDocCache?: readonly JSDocTag[] | undefined;
}
```

It is neither a NodeArray brand nor a syntactic intersection nor an element union. Its numeric index has element type JSDoc. The ancestry helper returns JSDoc[], and its only own field is optional jsDocCache. The checker does not classify the interface itself as an ordinary array. unsupportedViewFamily therefore classifies its inherited numeric index as dictionary. viewContract returns ViewUnknown, Unsupported=dictionary, with no element or own-field certificate. The registered ordinary-array adapter supplies an Element contract but no Fields. Replacing the interface with its array ancestor would discard the cache contract and cannot certify the full declared type. The separate own-field omission is a V3 augmented-array limitation; the dictionary dispatcher also needs an indexed-object composition contract. This report does not relabel the source declaration as an intersection.

More decisively, probing the real array ancestor also fails to supply a complete element graph. JSDoc at types.ts:3951 has parent: HasJSDoc (3953). The builder's exact diagnostic is:

```text
types.ts:963:1: stage 0 can't lower checked union representation for HasJSDoc yet
```

HasJSDoc at types.ts:1200 includes EndOfFileToken. EndOfFileToken at types.ts:1599 is the genuine intersection:

```text
Token<SyntaxKind.EndOfFileToken> & JSDocContainer
```

The probe checks every resolved HasJSDoc union member. EndOfFileToken is the sole member with no view representation; its family is intersection. Consequently the JSDoc element descriptor is ViewUnknown, Unsupported=representation conversion. The array ancestor has a reserved ViewArray descriptor, but that is not a complete contract or permission to erase element checks. contract-trace.json records both layers and the sole unknown member. This is an independent V5 intersection dependency even if V3's augmented-array dispatch is repaired. Under the unit's instruction, no V5 contract is built and all affected bodies remain pending.

Dependency: V5 intersections must represent the conjunctive EndOfFileToken object within HasJSDoc without losing either Token or JSDocContainer fields. The complete array view must also retain the optional own cache and recursive child contracts. No published compiler/views-v4* or compiler/views-v5* branch was returned by git ls-remote, so no implementing SHA is claimed. The 16 emit-helper callable bodies remain pending on V4. No other worker branch is merged.

## Pinned remeasurement

The existing local, unpushed predicates merge a1be057fc66b7c666282aa4fcfe56ffd93a4bb71 includes predicates acc2bfbcf, this candidate's production conversion, and V3 14690251. Executable IR output remains disabled by the measurement overlay. The same exact 138 coordinates from corpus 8a7ab17e are selected. All 79 source hashes and all 320 checker diagnostics match the previous measurement. after.json.gz is byte-equivalent as decoded JSON to the prior after-audited.json.gz. No local predicates merge is pushed.

| Group | Bodies | Logical proof before → after | Full admission before → after | Pending views before → after | .a refused before → after |
| --- | ---: | ---: | ---: | ---: | ---: |
| Direct kind | 133 | 133 → 133 | 0 → 0 | 133 → 133 | 133 → 133 |
| Delegation | 5 | 5 → 5 | 0 → 0 | 5 → 5 | 5 → 5 |
| Total | 138 | 138 → 138 | 0 → 0 | 138 → 138 | 138 → 138 |

First recorded descriptor failure is unchanged: JSDocArray 119; emit-helper callback 16; NodeArray<Modifier> 1; NodeArray<ModifierLike> 1; HasJSDoc 1. These are first boundaries, not mutually independent dependencies: fixing one does not promise the others disappear. Zero newly admitted. The 119 are not credited as passing just because a body proof or an ancestor descriptor exists.

## Observations and checks

The existing audit command, with --mutants, passes: 138 exact bodies, 79 identical source hashes, 320 preserved diagnostics; 138 pending, zero full admissions. It catches four independent evidence corruptions: drop-body, invent-pass, fabricate-check, source-hash. audit.log.gz records the intended catchers.

The Node-only array-own-field.a witness retains the reduced array/interface/cache declarations. String kind names replace the large tsc enum solely in this observation. Node 24.19.0 after stripTypeScriptTypes prints JSDoc, false, true, undefined, JSDocTag, JSDoc, one per line, and exits 0. stderr contains Node's experimental stripTypeScriptTypes warning, recorded separately. Missing and present undefined cache slots differ in own-property presence while array element reads continue to work. This is a Node observation, not a native or JavaScript backend pass.

No contract is implemented, so the requested element-check omission and wrong-kind acceptance mutants are not claimed run or caught. Both backend/sanitizer certifications remain pending on the V5 element graph and augmented-array own-field contract. No test leaf, oracle fixture or counts row is added or touched. There are therefore no new test-leaf seconds or changed count rows. No full gate or whole-package test suite is run. All new evidence is under this review directory, and the Go probe is named probe-hook.go.txt.

Reproduction uses the existing local merge and its executable-disabled after overlay, replacing only the scratch hatch predicate hook with probe-hook.go.txt. Build command:

```text
source /workspace/adamic-tools/env.sh
go build -buildvcs=false -overlay /workspace/scratch/jsdocarray-overlay/overlay.json -tags hatch_predicate_measurement -o /tmp/jsdocarray-measure ./stage3/scouts/step09/predicates/measure-probe
HATCH_JSDOC_TRACE=1 HATCH_BODY_ONLY=1 HATCH_SELECTION=/tmp/delegation-original-selection.json /tmp/jsdocarray-measure /workspace/scratch/predicates-corpus
python3 review/compiler/views-optional-object-slots/audit.py review/compiler/views-optional-object-slots/results.json review/compiler/views-optional-object-slots/after-audited.json.gz /tmp/jsdocarray-after.json /workspace/scratch/predicates-corpus --mutants
```

Build and measurement succeed with logs saved. The first scratch build without -buildvcs=false failed to obtain VCS status because the measurement worktree references the existing cohere submodule via symlink; disabling VCS stamping fixed the scratch build. No cohere source is copied.

Required setup used GOPROXY=https://proxy.golang.org|direct. Timing lines: Node 0.019s; Go 0.021s; markdown 0.059s; submodules 0.068s; clang 0.148s; build 36.854s; deferred test binaries 36.998s; cache warm 37.004s; done 37.031s. nproc=5, cgroup cpu.max=400000 100000 (four CPUs). Go 1.27.1, clang 20.1.8, Node 24.19.0. setup.log.gz retains the complete output.

## Remaining design question

V3's own-field array limitation remains separate from V5's genuine intersection blocker. Which slice owns the combined array-elements plus optional own-field adapter? The current dispatcher calls it dictionary, but the source is an interface extending Array. Regardless of ownership, admitting the 119 requires the independently observed V5 EndOfFileToken representation too. This unit leaves both obligations visible rather than treating array ancestry as a complete proof.

## Committed lane checks

After evidence commit 13bd54ab5, the required repository-root command passes with the toolchain environment sourced:

```text
lane checks 0.9 s: gofmt and tools on 63 Go files, t.Parallel on 7 test packages; no t.Parallel analyzer on this tree; vet 7 packages
```

The first invocation omitted the environment and stopped before checks with FileNotFoundError: gofmt. The sourced rerun succeeds; both logs are retained. No Go or test file changed in this unit. The inherited V3 tree still lacks the parallel analyzer, as the lane output states. The previous unit's independent integration-analyzer receipt remains in the parent review directory. The final evidence commit is checked again before the single push to this feature branch. No main or area branch is pushed, and the predicates measurement merge remains local.

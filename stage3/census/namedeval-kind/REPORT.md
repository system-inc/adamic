# namedEvaluation kind census — #0kkpw2e

Report only. No adaptation, ledger or internal/ change is made on this branch.

## Direct answers

**S1: yes, the helper leaves its overload-result stop.** The `injectClassNamedEvaluationHelperBlockIfMissing` Refused at namedEvaluation.ts:173:1 disappears. Its implementation unit at :179:1 then reaches **NotYet at namedEvaluation.ts:203:11: `a value of type ClassNamedEvaluationHelperBlock`**. The same existing helper-body wall is reached from esDecorators' visitClassDeclaration unit. These are reached walls in already-blocked units, not new defects. The unchanged refusal-scan readonly-widening findings at namedEvaluation.ts:205:68 and :212:18 already existed at B; they are distinct from this newly reached lowering frontier.

**The census exercises one of the three caller checks, not all three.** At upstream esDecorators.ts:1062 (adapted :1061:24), the S1 trace observes one successful checked result with concrete promise `ClassDeclaration`, marked=true, checked delta=1, proven=0, erased=0, error="". The old helper covariance refusal at this caller goes. The other two caller units fail their existing prologues before their calls, so this census cannot claim checks were constructed there. This does not mean S1 lacks a check route for those concrete Extract instances; it means the requested real-tsc units do not reach it yet.

**S2: yes, transformNamedEvaluation now reports the specified NotYet at its declaration.** At namedEvaluation.ts:469:1 (implementation unit :470:1), `an overload result Extract over intersection members the implementation's result doesn't serve` replaces the covariance Refused. Its associated prologue Boundary changes its underlying message too. Nothing else moves in the requested semantic finding sets: the other 72 namedEvaluation findings and all 587 esDecorators findings pair unchanged; no body is newly reached by S2. The coherent scan-all frontier is also unchanged.

## Frozen revisions and common tsc input

| Case | Compiler SHA |
| --- | --- |
| B | `932a6344af63dc7c8728fffac05584c176d0475f` |
| S1 | `dfff19bb0ab5722136f7c19ea01505a1ab0946df` |
| S2 | `12047c681cae67e2e8b83d2c2b57d969237bdd3f` |

B is the merge base of S1 and fetched origin/main `932a6344af63dc7c8728fffac05584c176d0475f`. S2 is a descendant of S1. Each compiler is built from its own detached worktree. Each has its own freshly generated scan-all overlay, freshly generated latent overlay and separately built probes. All six canonical binary SHA-256 digests are distinct. No probe or overlay is reused across SHAs; only unchanged Go dependency caches are shared. Exact build commands, own-child exit codes, source hashes, generated-overlay hashes and binary digests are in build-provenance.json and the *-build/*-overlay receipts.

Apply does **not** differ: all three keys equal `9a14bca24b6ba48dcf4eb67000e565e36a08f214ba76cc4319950bea61b1324d`. The one adapted tsc tree was produced by B's stage3/apply.sh, pin 050880ce59e30b356b686bd3144efe24f875ebc8 (TypeScript v6.0.3), and is shared unchanged across every run. inputs.json records all 79 compiler source hashes. This is the B tree, without the later getProperties landing.

Each canonical full run sets LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, exact LATENT_ONLY_FILE and CommonJSBundle, using that case's probe. All six complete streams contain their relation_report trailer and report zero checker diagnostics. The tools' generic "measured on a checker-rejected program" label does not override those checker-clean headers. No usable IR is exposed. Measurements remain blocked; process exit zero means measurement completed, not tsc native-compilation acceptance.

## Full single-file totals

| File | B | S1 | S2 |
| --- | ---: | ---: | ---: |
| transformers/namedEvaluation.ts | 54 | 74 | 74 |
| transformers/esDecorators.ts | 567 | 587 | 587 |

## Pairwise deltas and grades

Canonical pairing is exactly `(kind, phase, reason, message)` with source positions stripped. Multiplicity is retained. Same-position/unit matches are preferred only to attribute duplicates; positions and units are not part of the identity key. Full original words and positions are retained in every paired record. Only after canonical pairing, unmatched records at the same unchanged unit/phase/source boundary are presented as in-place **changed** diagnostics. Thus the Refused-to-NotYet change is not incorrectly paired as an unchanged identity. The JSON also retains strictCanonicalGone and strictCanonicalAppeared, making that distinction explicit.

An unmatched after-finding in a unit already blocked at B/S1 is **reached**, not new. New is reserved for an after-finding in a previously unblocked unit. Counts below include Boundary bookkeeping; each S1 row's 22 reached events are 11 diagnostic instances and 11 Boundaries. Its two gone events are the old covariance refusal and its Boundary. S2's two changed events are one production diagnostic and its prologue Boundary.

| Pair | File | New | Reached | Gone | Changed | Unchanged pairs |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| B → S1 | namedEvaluation.ts | 0 | 22 | 2 | 0 | 52 |
| B → S1 | esDecorators.ts | 0 | 22 | 2 | 0 | 565 |
| S1 → S2 | namedEvaluation.ts | 0 | 0 | 0 | 2 | 72 |
| S1 → S2 | esDecorators.ts | 0 | 0 | 0 | 0 | 587 |

Strict identity residues, before in-place classification: B→S1 is 2 unmatched before / 22 unmatched after in each file; S1→S2 is 2 / 2 in namedEvaluation and 0 / 0 in esDecorators. Those unmatched after records are not silently called new. Every full record appears in the appendices and *-delta.json files; unchanged pairs are retained in JSON too.

## Three real callers

Caller identities and concrete Extract result instances are resolved with stock TypeScript 6.0.3 getResolvedSignature, not identified by text alone (caller-mapping.stdout). Existing adaptations shift the two esDecorators call lines by -1; namedEvaluation's line is unchanged.

| Requested upstream caller | Adapted caller | Concrete result | Full-census observation |
| --- | --- | --- | --- |
| esDecorators.ts:1062 | esDecorators.ts:1061:24 | ClassDeclaration | One checked result constructed successfully; whole unit remains blocked at later existing helper-body walls |
| esDecorators.ts:676 | esDecorators.ts:675:20 | ClassDeclaration \| ClassExpression | Not reached: transformClassLike unit :669:5 stops at :669:14, `a function returning ImmediatelyInvokedArrowFunction` |
| namedEvaluation.ts:250 | namedEvaluation.ts:250:14 | ClassExpression | Not reached: finishTransformNamedEvaluation unit :236:1 stops at :238:5, `a value of type WrappedExpression<AnonymousFunctionDefinition>` |

The auxiliary S1-only trace reads counters around one invocation of the original overloadResultCheck method, preserves its returned IR expression/error, and writes metadata to stderr. It does not replace a canonical probe or relax any no-output guard. Its complete findings match the canonical S1 records exactly in both files. caller-observations.json contains the raw successful check event, the two blocked-prologue findings, concrete promises, positions and trace digests. This proves check **construction** at the one reached call; it is not execution of a native tsc program.

## scan-all first stop

All three canonical scans complete their measurement with Completed=false and the same production boundary, lowering.declareModule. The first coherent declaration/body/tail stop is:

```text
/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/core.ts:261:1: overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined
```

This is the existing filter overload at core.ts:261:1, before the namedEvaluation declarations. Therefore neither S1 nor S2 advances this whole-import-closure frontier. A zero delta here cannot be used to claim bodies beyond it pass.

For completeness, the first recorded refusal-scan event is also unchanged in every case:

```text
/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilitiesPublic.ts:778:16: a cast the runtime can't check
```

Each scan records 5,054 events. Canonical multiset deltas are empty for both pairs. Complete raw scans are *-scan-namedEvaluation.stdout; scan-first-stops.json preserves both meanings of "first stop", including full words/positions; scan-deltas.json records the pairwise equality. The scan is deliberately blind after its first production abort. Claims that nothing else moves are limited to the two full single-file censuses and this recorded scan scope.

## Full changed, gone and reached records

The records below preserve exact words, positions, unit attribution and phases. Source-position stripping is used only for matching, never to truncate displayed evidence.

### B-to-S1-namedEvaluation

New:

```json
[]
```

Gone:

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Refused",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:173:1",
    "reason": "overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> | Extract<...> cannot be served by implementation result ClassLikeDeclaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:173:1: Adamic 0.1 refuses overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> | Extract<...> cannot be served by implementation result ClassLikeDeclaration; make the implementation result covariant with every overload result"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "reason": "function signature or prologue failed; continued at next declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:173:1: Adamic 0.1 refuses overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> | Extract<...> cannot be served by implementation result ClassLikeDeclaration; make the implementation result covariant with every overload result",
    "start": 7070,
    "end": 8605
  }
]
```

Changed:

```json
[]
```

Reached:

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:203:11",
    "reason": "a value of type ClassNamedEvaluationHelperBlock",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:203:11: stage 0 can't lower a value of type ClassNamedEvaluationHelperBlock yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:203:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:203:11: stage 0 can't lower a value of type ClassNamedEvaluationHelperBlock yet",
    "start": 7486,
    "end": 7598
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/factory/emitNode.ts:139:17",
    "reason": "a function returning ExpressionStatement & { readonly expression: CallExpression & { readonly expression: Identifier; }; }",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/factory/emitNode.ts:139:17: stage 0 can't lower a function returning ExpressionStatement & { readonly expression: CallExpression & { readonly expression: Identifier; }; } yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:205:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/factory/emitNode.ts:139:17: stage 0 can't lower a function returning ExpressionStatement & { readonly expression: CallExpression & { readonly expression: Identifier; }; } yet",
    "start": 7620,
    "end": 7700
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/classThis.ts:72:1",
    "reason": "checked predicate overload target ClassThisAssignmentBlock",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/classThis.ts:72:1: stage 0 can't lower checked predicate overload target ClassThisAssignmentBlock yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:208:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/classThis.ts:72:1: stage 0 can't lower checked predicate overload target ClassThisAssignmentBlock yet",
    "start": 7707,
    "end": 7794
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:209:43",
    "reason": "reading insertionIndex",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:209:43: stage 0 can't lower reading insertionIndex yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:209:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:209:43: stage 0 can't lower reading insertionIndex yet",
    "start": 7794,
    "end": 7854
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:210:41",
    "reason": "reading insertionIndex",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:210:41: stage 0 can't lower reading insertionIndex yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:210:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:210:41: stage 0 can't lower reading insertionIndex yet",
    "start": 7854,
    "end": 7912
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:211:49",
    "reason": "reading leading",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:211:49: stage 0 can't lower reading leading yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:211:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:211:49: stage 0 can't lower reading leading yet",
    "start": 7912,
    "end": 8007
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Refused",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10657:6",
    "reason": "a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10657:6: Adamic 0.1 refuses a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10657:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10657:6: Adamic 0.1 refuses a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
    "start": 427673,
    "end": 427710
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Refused",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10647:6",
    "reason": "a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10647:6: Adamic 0.1 refuses a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10647:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10647:6: Adamic 0.1 refuses a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
    "start": 427403,
    "end": 427440
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:212:18",
    "reason": "reading members",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:212:18: stage 0 can't lower reading members yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:212:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:212:18: stage 0 can't lower reading members yet",
    "start": 8007,
    "end": 8049
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:221:13",
    "reason": "reading members",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:221:13: stage 0 can't lower reading members yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:214:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:221:13: stage 0 can't lower reading members yet",
    "start": 8049,
    "end": 8522
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:232:5",
    "reason": "writing a possibly absent optional own field",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:232:5: stage 0 can't lower writing a possibly absent optional own field yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:179:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:232:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:232:5: stage 0 can't lower writing a possibly absent optional own field yet",
    "start": 8522,
    "end": 8584
  }
]
```

### B-to-S1-esDecorators

New:

```json
[]
```

Gone:

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Refused",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:173:1",
    "reason": "overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> | Extract<...> cannot be served by implementation result ClassLikeDeclaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:173:1: Adamic 0.1 refuses overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> | Extract<...> cannot be served by implementation result ClassLikeDeclaration; make the implementation result covariant with every overload result"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1061:17",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:173:1: Adamic 0.1 refuses overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> | Extract<...> cannot be served by implementation result ClassLikeDeclaration; make the implementation result covariant with every overload result",
    "start": 51584,
    "end": 51682
  }
]
```

Changed:

```json
[]
```

Reached:

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:203:11",
    "reason": "a value of type ClassNamedEvaluationHelperBlock",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:203:11: stage 0 can't lower a value of type ClassNamedEvaluationHelperBlock yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:203:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:203:11: stage 0 can't lower a value of type ClassNamedEvaluationHelperBlock yet",
    "start": 7486,
    "end": 7598
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/factory/emitNode.ts:139:17",
    "reason": "a function returning ExpressionStatement & { readonly expression: CallExpression & { readonly expression: Identifier; }; }",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/factory/emitNode.ts:139:17: stage 0 can't lower a function returning ExpressionStatement & { readonly expression: CallExpression & { readonly expression: Identifier; }; } yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:205:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/factory/emitNode.ts:139:17: stage 0 can't lower a function returning ExpressionStatement & { readonly expression: CallExpression & { readonly expression: Identifier; }; } yet",
    "start": 7620,
    "end": 7700
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/classThis.ts:72:1",
    "reason": "checked predicate overload target ClassThisAssignmentBlock",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/classThis.ts:72:1: stage 0 can't lower checked predicate overload target ClassThisAssignmentBlock yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:208:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/classThis.ts:72:1: stage 0 can't lower checked predicate overload target ClassThisAssignmentBlock yet",
    "start": 7707,
    "end": 7794
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:209:43",
    "reason": "reading insertionIndex",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:209:43: stage 0 can't lower reading insertionIndex yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:209:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:209:43: stage 0 can't lower reading insertionIndex yet",
    "start": 7794,
    "end": 7854
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:210:41",
    "reason": "reading insertionIndex",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:210:41: stage 0 can't lower reading insertionIndex yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:210:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:210:41: stage 0 can't lower reading insertionIndex yet",
    "start": 7854,
    "end": 7912
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:211:49",
    "reason": "reading leading",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:211:49: stage 0 can't lower reading leading yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:211:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:211:49: stage 0 can't lower reading leading yet",
    "start": 7912,
    "end": 8007
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Refused",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10657:6",
    "reason": "a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10657:6: Adamic 0.1 refuses a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10657:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10657:6: Adamic 0.1 refuses a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
    "start": 427673,
    "end": 427710
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Refused",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10647:6",
    "reason": "a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10647:6: Adamic 0.1 refuses a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10647:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/utilities.ts:10647:6: Adamic 0.1 refuses a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
    "start": 427403,
    "end": 427440
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:212:18",
    "reason": "reading members",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:212:18: stage 0 can't lower reading members yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:212:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:212:18: stage 0 can't lower reading members yet",
    "start": 8007,
    "end": 8049
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:221:13",
    "reason": "reading members",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:221:13: stage 0 can't lower reading members yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:214:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:221:13: stage 0 can't lower reading members yet",
    "start": 8049,
    "end": 8522
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:232:5",
    "reason": "writing a possibly absent optional own field",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:232:5: stage 0 can't lower writing a possibly absent optional own field yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/esDecorators.ts:1053:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:232:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:232:5: stage 0 can't lower writing a possibly absent optional own field yet",
    "start": 8522,
    "end": 8584
  }
]
```

### S1-to-S2-namedEvaluation

New:

```json
[]
```

Gone:

```json
[]
```

Changed:

```json
[
  {
    "before": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:470:1",
      "phase": "lowering",
      "kind": "Refused",
      "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:469:1",
      "reason": "overload 1 of transformNamedEvaluation result Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> | ... 7 more ... | Extract<...> cannot be served by implementation result BinaryExpression | BindingElement | ExportAssignment | ParameterDeclaration | PropertyAssignment | PropertyDeclaration | ShorthandPropertyAssignment | VariableDeclaration",
      "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:469:1: Adamic 0.1 refuses overload 1 of transformNamedEvaluation result Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> | ... 7 more ... | Extract<...> cannot be served by implementation result BinaryExpression | BindingElement | ExportAssignment | ParameterDeclaration | PropertyAssignment | PropertyDeclaration | ShorthandPropertyAssignment | VariableDeclaration; make the implementation result covariant with every overload result"
    },
    "after": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:470:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:469:1",
      "reason": "an overload result Extract over intersection members the implementation's result doesn't serve",
      "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:469:1: stage 0 can't lower an overload result Extract over intersection members the implementation's result doesn't serve yet"
    },
    "grade": "changed diagnostic at the same unit/source boundary"
  },
  {
    "before": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:470:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:470:1",
      "reason": "function signature or prologue failed; continued at next declaration",
      "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:469:1: Adamic 0.1 refuses overload 1 of transformNamedEvaluation result Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> | ... 7 more ... | Extract<...> cannot be served by implementation result BinaryExpression | BindingElement | ExportAssignment | ParameterDeclaration | PropertyAssignment | PropertyDeclaration | ShorthandPropertyAssignment | VariableDeclaration; make the implementation result covariant with every overload result",
      "start": 20947,
      "end": 22316
    },
    "after": {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:470:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:470:1",
      "reason": "function signature or prologue failed; continued at next declaration",
      "text": "/workspace/scratch/namedeval-kind-runtime/adapted-B/src/compiler/transformers/namedEvaluation.ts:469:1: stage 0 can't lower an overload result Extract over intersection members the implementation's result doesn't serve yet",
      "start": 20947,
      "end": 22316
    },
    "grade": "changed diagnostic at the same unit/source boundary"
  }
]
```

Reached:

```json
[]
```

### S1-to-S2-esDecorators

New:

```json
[]
```

Gone:

```json
[]
```

Changed:

```json
[]
```

Reached:

```json
[]
```

## Audit and scope

All canonical cases and auxiliary cases have their own command/exit receipts. The adapted source hashes are checked again after measurement. B, S1 and S2 remain at their recorded SHAs with no production source edits. This branch changes only stage3/census/namedeval-kind/. It contains no adaptation, ledger or internal/ change. No feature was merged or proposed here. Nothing was stopped or fabricated to make a census pass.

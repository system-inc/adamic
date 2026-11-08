# Absent optional write: paired latent first-error census

Task g2aj1pb. Report only. No compiler or adaptation change is committed.

Parent and apply: `e3aa0bb7642163ebd004a71ba40a4657d27eab71`. Head: `e1d221ce929ec37a4677f7ae98e076206f522634`, exactly one commit on the parent. TypeScript 6.0.3 source: `050880ce59e30b356b686bd3144efe24f875ebc8`. Checker dependency: `06af2e399c6c1fb93fd570b774027942747400b4`.

The head keys `absentOptionalWrite` by field name and includes `decodeJson` as an omitted-field producer. Its wording at observed new sites is quoted verbatim below. The omission map is global by name; this census does not attribute a finding to a particular literal or decoded producer.

Method: [library-method-view reference](https://github.com/system-inc/adamic/blob/5f83ad9f1ea1aa994c11de2cd9a7e6e0184e742b/stage3/census/library-method-view-bbc1eeb2/REPORT.md). Each guarded entry driver and overlay is built from its own compiler SHA. Both run over the same unchanged parent apply, with `LATENT_FULL`, `LATENT_SPECULATIVE`, and `LATENT_ONLY_FILE` unset and `LATENT_ASSERT_NO_OUTPUT=1`. The entry is `src/tsc/tsc.ts`. Pairing uses the parent `stage3/census/pair` and stock TypeScript 6.0.3 outermost-access-chain normalization.

The streams have 81 reached source files and 0 checker diagnostics on each side. Headers and attempted eligibility match. This is a lowering observation ledger, not proof of previous native correctness or exhaustive expression coverage. Generic deferrals, isolated unit state and first-error stopping retain the reference method limits. No full census, replay, native build or full gate ran.

## Instrumentation correction required for this guard

The pinned latent driver does not call `noteOmittedOptionals`, although production lowering does. Its fresh unit lowerings therefore leave `omittedOptionals` nil. Both the declaration-keyed parent lookup and the name-keyed head lookup then miss every non-Union omission. A literal reference-method run would be blind to the requested change. The original parent run was interrupted with SIGTERM before any pairing; its partial stream, own child status and original overlays remain under [unprimed-interrupted](unprimed-interrupted/). No count or verdict is inferred from that incomplete run.

To measure this guard, [prime_overlay.py](prime_overlay.py), with the exact one-line changes in [prepass-change.go.txt](prepass-change.go.txt), inserts exactly `l.noteOmittedOptionals(program.Files())` after `noteAccessorNames` in each first-error unit setup, using the production implementation from that side. This symmetric measurement-only correction is the sole intentional instrumentation departure from the reference. It changes neither checked-in compiler code nor eligibility, stopping policy, access-chain normalization or no-output guards. Each attempted unit runs the actual prepass; no omission-map approximation or memoization is introduced. The counts below describe these corrected first-error drivers.

The [discriminating probe](prepass-proof/probe.ts) leaves `Mut.name` out and writes through a different interface declaring `name` again. With zero checker diagnostics, the original head driver records no finding, the corrected parent records no finding, and the corrected head records exactly `NotYet: writing a possibly absent optional own field` at line 5, column 1. [Probe results](prepass-proof/result.json) and all three streams are retained. The [decodeJson probe](prepass-proof/decoded.a) independently gives the same three-way result, with only the corrected head stopping at its write on line 6, column 5; [results](prepass-proof/decoded-result.json) retain the exact wording. The focused own-pin production head test for `other_interface.ts` also passed.

## Counts

| View | Cleared | New | Changed | New sites | First-stop changes | Regression entries |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Raw | 1 | 1 | 0 | 1 | 1 | 0 |
| Distinct outermost access chains | 1 | 1 | 0 | 1 | 1 | 0 |

Cleared means no finding remains at that identity. Changed means a removed finding has another reason remaining there. New includes replacement reasons, so new and changed need not be disjoint. Raw counts unique exact-position keys, not event multiplicity. Distinct keys include source hash, outermost access-chain start, and exact reason. A cleared first-error finding can be hidden by an earlier new stop and is not necessarily a compiler improvement. Unit first stops use lowering events only; refusal-scan events do not make a unit newly stopped.

| View | Before | After |
| --- | ---: | ---: |
| Raw | 5116 | 5116 |
| Distinct | 4914 | 4914 |

## Every newly stopped unit (0)

None.

## Every changed first stop in an already-blocked unit (1)

| Unit | Unit location | Parent stop | Head stop |
| --- | --- | --- | --- |
| processPragmasIntoFields | compiler/parser.ts:10619:1 | NotYet: a value of type PragmaPseudoMap[TKey][] &#124; PragmaPseudoMap[TKey] at compiler/parser.ts:10625:31 | NotYet: writing a possibly absent optional own field at compiler/parser.ts:10620:5 |

Cleared unit first stops: 0.


## New findings grouped by reason and file

Every distinct new finding is listed below, including findings in already-blocked units. The category considers all recording contexts and their parent lowering first stops. The file tables show one representative unit; the JSON retains every context. The paired JSON retains raw positions; the streams and new-findings JSON retain every recording context.

| Reason | File | Raw new | Distinct new |
| --- | --- | ---: | ---: |
| writing a possibly absent optional own field | compiler/parser.ts | 1 | 1 |

### writing a possibly absent optional own field

Exact head wording:

> NotYet: writing a possibly absent optional own field

#### compiler/parser.ts

Raw new: 1. Distinct new: 1.

| Site | Recording unit | Unit location | Phase | Category |
| --- | --- | --- | --- | --- |
| compiler/parser.ts:10620:5 | processPragmasIntoFields | compiler/parser.ts:10619:1 | lowering | already-blocked unit |

Exact diagnostic at `compiler/parser.ts:10620:5`:

```text
/workspace/aow-adapted/src/compiler/parser.ts:10620:5: stage 0 can't lower writing a possibly absent optional own field yet
```

Source lines 10618 through 10622:

```typescript
/** @internal */
export function processPragmasIntoFields(context: PragmaContext, reportDiagnostic: PragmaDiagnosticReporter): void {
    context.checkJsDirective = undefined;
    context.referencedFiles = [];
    context.typeReferenceDirectives = [];
```

## New findings inside already-blocked units (findings: 1; contexts: 1)

| Site | Already-blocked recording unit and location | Exact head reason |
| --- | --- | --- |
| compiler/parser.ts:10620:5 | processPragmasIntoFields at compiler/parser.ts:10619:1 | NotYet: writing a possibly absent optional own field |

## Raw-only position additions (0)

These raw new findings collapse into reasons already present on the same distinct access chain. They are included in the raw totals, not the distinct new-finding count.

| Site | Representative recording unit | Kind and exact reason |
| --- | --- | --- |
| None | | |

## Panics, resources and validation

Panics: 0 before, 0 after.

| Run | Child exit | Wall seconds | Peak RSS MiB |
| --- | ---: | ---: | ---: |
| parent | 0 | 2942.735 | 7811.0 |
| head | 0 | 3120.993 | 8012.8 |

Peak RSS uses Linux child getrusage. Wall times include polling and overlap after a memory-aware stagger, so they are shared-host observations, not benchmarks. Runners retain each child exit status and stop it after 30 minutes without progress. All logs go to files and are never piped into head or tail. Setup, npm ci, parent apply and both builds passed. The first overlay attempt failed because the scratch workspace redundantly replaced its listed checker module. Removing that replacement fixed the workspace; the failed attempt log and status are retained. No production source changed.

- [Parent raw stream](parent-run.jsonl), [head raw stream](head-run.jsonl), [paired JSON](paired.json), [paired ledger](paired.md).
- [Newly stopped units](newly-stopped.json), [changed first stops](changed-first-stops.json), [all distinct new findings and source context](new-findings.json).
- [Applied patches](patch-set.md), [frozen source hashes](source-hashes.json), [compiler and binary provenance](provenance.json), [guarded overlays](overlays/), [independent audit](audit.json).
- [Apply checkpoint](checkpoint-apply.json), [corrected-driver checkpoint](checkpoint-primed.json), and [completed-parent checkpoint](checkpoint-parent.json) are on separate suffixed refs, preserving one final commit directly on e3aa0bb7 without force-pushing. Main is never pushed.
- The audit independently recounts raw totals and first-stop changes and verifies all adapted and compiler source hashes. Every reached source matches the frozen manifest. Go evidence filenames end in `.go.txt`.

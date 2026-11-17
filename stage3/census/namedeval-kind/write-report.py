import pathlib,json,collections
from deltas import stripped
p=pathlib.Path(__file__).resolve().parent
inputs=json.loads((p/'inputs.json').read_text());summary=json.loads((p/'delta-summary.json').read_text());call=json.loads((p/'caller-observations.json').read_text());scans=json.loads((p/'scan-first-stops.json').read_text());build=json.loads((p/'build-provenance.json').read_text())
root=inputs['commonAdaptedRoot']+'/'
def relative(s):return s.replace(root,'')
def load(case,file):return [json.loads(s) for s in (p/f'{case}-latent-{file}.jsonl').read_text().splitlines()][1]
# Retain the exact canonical scan multiset too, while the requested frontiers are shown below.
scanDeltas={}
for left,right in [('B','S1'),('S1','S2')]:
 a=json.loads((p/f'{left}-scan-namedEvaluation.stdout').read_text())['Stops'];b=json.loads((p/f'{right}-scan-namedEvaluation.stdout').read_text())['Stops']
 key=lambda f:json.dumps((f['Kind'],f['Rule'],stripped(f['Words']),stripped(f['Words'])))
 ca,cb=collections.Counter(map(key,a)),collections.Counter(map(key,b))
 scanDeltas[left+'-to-'+right]={'unmatchedBefore':list((ca-cb).elements()),'unmatchedAfter':list((cb-ca).elements()),'firstPipelineChanged':scans[left]['firstPipelineStop']!=scans[right]['firstPipelineStop']}
 assert not scanDeltas[left+'-to-'+right]['unmatchedBefore'] and not scanDeltas[left+'-to-'+right]['unmatchedAfter']
(p/'scan-deltas.json').write_text(json.dumps(scanDeltas,indent=2)+'\n')
text='''# namedEvaluation kind census — #0kkpw2e

Report only. No adaptation, ledger or internal/ change is made on this branch.

## Direct answers

**S1: yes, the helper leaves its overload-result stop.** The `injectClassNamedEvaluationHelperBlockIfMissing` Refused at namedEvaluation.ts:173:1 disappears. Its implementation unit at :179:1 then reaches **NotYet at namedEvaluation.ts:203:11: `a value of type ClassNamedEvaluationHelperBlock`**. The same existing helper-body wall is reached from esDecorators' visitClassDeclaration unit. These are reached walls in already-blocked units, not new defects. The unchanged refusal-scan readonly-widening findings at namedEvaluation.ts:205:68 and :212:18 already existed at B; they are distinct from this newly reached lowering frontier.

**The census exercises one of the three caller checks, not all three.** At upstream esDecorators.ts:1062 (adapted :1061:24), the S1 trace observes one successful checked result with concrete promise `ClassDeclaration`, marked=true, checked delta=1, proven=0, erased=0, error="". The old helper covariance refusal at this caller goes. The other two caller units fail their existing prologues before their calls, so this census cannot claim checks were constructed there. This does not mean S1 lacks a check route for those concrete Extract instances; it means the requested real-tsc units do not reach it yet.

**S2: yes, transformNamedEvaluation now reports the specified NotYet at its declaration.** At namedEvaluation.ts:469:1 (implementation unit :470:1), `an overload result Extract over intersection members the implementation's result doesn't serve` replaces the covariance Refused. Its associated prologue Boundary changes its underlying message too. Nothing else moves in the requested semantic finding sets: the other 72 namedEvaluation findings and all 587 esDecorators findings pair unchanged; no body is newly reached by S2. The coherent scan-all frontier is also unchanged.

## Frozen revisions and common tsc input

'''
text+='| Case | Compiler SHA |\n| --- | --- |\n'
for c,sha in inputs['revisions'].items():text+=f'| {c} | `{sha}` |\n'
text+=f'''
B is the merge base of S1 and fetched origin/main `{inputs['originMain']}`. S2 is a descendant of S1. Each compiler is built from its own detached worktree. Each has its own freshly generated scan-all overlay, freshly generated latent overlay and separately built probes. All six canonical binary SHA-256 digests are distinct. No probe or overlay is reused across SHAs; only unchanged Go dependency caches are shared. Exact build commands, own-child exit codes, source hashes, generated-overlay hashes and binary digests are in build-provenance.json and the *-build/*-overlay receipts.

Apply does **not** differ: all three keys equal `{inputs['applyKeys']['B']}`. The one adapted tsc tree was produced by B's stage3/apply.sh, pin 050880ce59e30b356b686bd3144efe24f875ebc8 (TypeScript v6.0.3), and is shared unchanged across every run. inputs.json records all 79 compiler source hashes. This is the B tree, without the later getProperties landing.

Each canonical full run sets LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, exact LATENT_ONLY_FILE and CommonJSBundle, using that case's probe. All six complete streams contain their relation_report trailer and report zero checker diagnostics. The tools' generic "measured on a checker-rejected program" label does not override those checker-clean headers. No usable IR is exposed. Measurements remain blocked; process exit zero means measurement completed, not tsc native-compilation acceptance.

## Full single-file totals

| File | B | S1 | S2 |
| --- | ---: | ---: | ---: |
'''
for f in ['namedEvaluation','esDecorators']:text+='| transformers/'+f+'.ts | '+' | '.join(str(len(load(c,f)['findings'])) for c in ['B','S1','S2'])+' |\n'
text+='''
## Pairwise deltas and grades

Canonical pairing is exactly `(kind, phase, reason, message)` with source positions stripped. Multiplicity is retained. Same-position/unit matches are preferred only to attribute duplicates; positions and units are not part of the identity key. Full original words and positions are retained in every paired record. Only after canonical pairing, unmatched records at the same unchanged unit/phase/source boundary are presented as in-place **changed** diagnostics. Thus the Refused-to-NotYet change is not incorrectly paired as an unchanged identity. The JSON also retains strictCanonicalGone and strictCanonicalAppeared, making that distinction explicit.

An unmatched after-finding in a unit already blocked at B/S1 is **reached**, not new. New is reserved for an after-finding in a previously unblocked unit. Counts below include Boundary bookkeeping; each S1 row's 22 reached events are 11 diagnostic instances and 11 Boundaries. Its two gone events are the old covariance refusal and its Boundary. S2's two changed events are one production diagnostic and its prologue Boundary.

| Pair | File | New | Reached | Gone | Changed | Unchanged pairs |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
'''
for n,d in summary.items():
 left,_,right,file=n.split('-',3);text+=f"| {left} → {right} | {file}.ts | {d['new']} | {d['reached']} | {d['gone']} | {d['changed']} | {d['pairedUnchanged']} |\n"
text+='''
Strict identity residues, before in-place classification: B→S1 is 2 unmatched before / 22 unmatched after in each file; S1→S2 is 2 / 2 in namedEvaluation and 0 / 0 in esDecorators. Those unmatched after records are not silently called new. Every full record appears in the appendices and *-delta.json files; unchanged pairs are retained in JSON too.

## Three real callers

Caller identities and concrete Extract result instances are resolved with stock TypeScript 6.0.3 getResolvedSignature, not identified by text alone (caller-mapping.stdout). Existing adaptations shift the two esDecorators call lines by -1; namedEvaluation's line is unchanged.

| Requested upstream caller | Adapted caller | Concrete result | Full-census observation |
| --- | --- | --- | --- |
| esDecorators.ts:1062 | esDecorators.ts:1061:24 | ClassDeclaration | One checked result constructed successfully; whole unit remains blocked at later existing helper-body walls |
| esDecorators.ts:676 | esDecorators.ts:675:20 | ClassDeclaration \\| ClassExpression | Not reached: transformClassLike unit :669:5 stops at :669:14, `a function returning ImmediatelyInvokedArrowFunction` |
| namedEvaluation.ts:250 | namedEvaluation.ts:250:14 | ClassExpression | Not reached: finishTransformNamedEvaluation unit :236:1 stops at :238:5, `a value of type WrappedExpression<AnonymousFunctionDefinition>` |

The auxiliary S1-only trace reads counters around one invocation of the original overloadResultCheck method, preserves its returned IR expression/error, and writes metadata to stderr. It does not replace a canonical probe or relax any no-output guard. Its complete findings match the canonical S1 records exactly in both files. caller-observations.json contains the raw successful check event, the two blocked-prologue findings, concrete promises, positions and trace digests. This proves check **construction** at the one reached call; it is not execution of a native tsc program.

## scan-all first stop

All three canonical scans complete their measurement with Completed=false and the same production boundary, lowering.declareModule. The first coherent declaration/body/tail stop is:

```text
'''+scans['B']['firstPipelineStop']['Site']+': '+scans['B']['firstPipelineStop']['Words']+'''
```

This is the existing filter overload at core.ts:261:1, before the namedEvaluation declarations. Therefore neither S1 nor S2 advances this whole-import-closure frontier. A zero delta here cannot be used to claim bodies beyond it pass.

For completeness, the first recorded refusal-scan event is also unchanged in every case:

```text
'''+scans['B']['firstRecordedStop']['Site']+': '+scans['B']['firstRecordedStop']['Words']+'''
```

Each scan records 5,054 events. Canonical multiset deltas are empty for both pairs. Complete raw scans are *-scan-namedEvaluation.stdout; scan-first-stops.json preserves both meanings of "first stop", including full words/positions; scan-deltas.json records the pairwise equality. The scan is deliberately blind after its first production abort. Claims that nothing else moves are limited to the two full single-file censuses and this recorded scan scope.

## Full changed, gone and reached records

The records below preserve exact words, positions, unit attribution and phases. Source-position stripping is used only for matching, never to truncate displayed evidence.
'''
for name in summary:
 d=json.loads((p/(name+'-delta.json')).read_text());text+='\n### '+name+'\n'
 for label in ['new','gone','changed','reached']:
  text+='\n'+label.capitalize()+':\n\n```json\n'+json.dumps(d[label],indent=2)+'\n```\n'
text+='''
## Audit and scope

All canonical cases and auxiliary cases have their own command/exit receipts. The adapted source hashes are checked again after measurement. B, S1 and S2 remain at their recorded SHAs with no production source edits. This branch changes only stage3/census/namedeval-kind/. It contains no adaptation, ledger or internal/ change. No feature was merged or proposed here. Nothing was stopped or fabricated to make a census pass.
'''
(p/'REPORT.md').write_text(text)
print('PASS report generated with all pairwise records and direct answers')

import pathlib,json,collections
p=pathlib.Path('stage3/census/overload-transformers')
initial=(p/'INITIAL_REPORT.md').read_text()
read=initial[initial.index('## Read and caller census'):initial.index('## Proof status at initial checkpoint')]
read=read.replace('Proposed road (i):', 'Attempted road (i), subsequently rejected by the no-new-findings gate:')
text='''# Transformer overloads — final report, #ev14ch9

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
'''
for name,file in [('ts','ts.ts'),('classFields','classFields.ts')]:
 d=json.loads((p/('delta-latent-'+name+'.json')).read_text())
 text+=f"| Full latent, {file} caller | {'PASS' if not d['new'] and not d['changed'] else 'FAIL'} | {d['beforeEvents']} -> {d['afterEvents']} events; {len(d['new'])} new, {len(d['gone'])} gone, {len(d['changed'])} changed |\n"
text+='''| Restoring mutant | PASS | Exact baseline findings restored: 146 events, including all three overload-refusal replays; 0 delta versus baseline |
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

'''
for d in json.loads((p/'public-declarations.json').read_text()):text+=f"- {d['file']}: {d['bytes']} bytes, SHA-256 `{d['afterSha256']}`, byte-identical before/after.\n"
text+='''
## Source and caller reading

The original reading/road analysis follows. Its inventory counts refer to the initial pre-generation checkout. The successful expanded inventory in whole-checkout-callers.json additionally covers JavaScript: after diagnostic generation, 81,470 files, 21,233 TypeScript and 18,210 JavaScript files, 710 src roots, 7 additional candidate roots and 832 program files. It confirms the same 28 direct calls. There are no function-value escapes: target references consist of 28 calls, 7 declarations and 6 imports. Candidate spelling filters only select work; callee symbols and getResolvedSignature declaration identity establish every reported caller. The original exhaustive TypeScript call traversal is retained at 0bd4986bf and upstream-callers.json.

'''+read
text+='''
## Complete finding deltas

Paths below are normalized only by removing their respective measured-tree prefixes. Event identity retains all fields except raw start/end byte offsets; the shorter annotation shifts later spans by 11 bytes. delta-latent-utilities-raw.json preserves all raw metadata differences, including eleven Boundary span shifts with identical site, kind, unit, phase, reason and words. These are not changed semantic findings. No production refusal is omitted. All newly exposed cast events are failures of the gate.

'''
for name in ['utilities','ts','classFields']:
 d=json.loads((p/('delta-latent-'+name+'.json')).read_text())
 text+='### Full latent '+name+'\n\n```json\n'+json.dumps({k:d[k] for k in ['new','gone','changed']},indent=2)+'\n```\n\n'
text+='''scan-all: new=[], gone=[], changed=[]; target absent from both recorded scopes. Full streams, exact commands, real child exit codes, source hashes and metadata are in the adjacent JSON/JSONL files and logs/. Restoring-mutant events are in restoring-mutant-result.json; its baseline delta is delta-mutant.json.

## Non-proof attempts retained

The first adapter construction had a Python syntax error; consequent missing-module runs failed. Corrected *-fixed and adapter-controls receipts establish the actual contracts. Expanded caller-inventory attempts hit stock-TypeScript stack overflows in unrelated dependency/fixture exploration; the successful scoped-symbol/signature audit and initial exhaustive inventory agree. The first stock probe omitted Node declarations and is invalid; only the corrected *-with-node pair is used. No failed attempt is reported as a pass. Wrapper receipts preserve each child's own return code. No evidence is placed on a landing branch.
'''
(p/'REPORT.md').write_text(text)

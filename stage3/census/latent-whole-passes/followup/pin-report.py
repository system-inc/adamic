import gzip,hashlib,json,pathlib,shutil
R=pathlib.Path(__file__).resolve().parent;E=R/'evidence/pin';E.mkdir(parents=True,exist_ok=True);S=pathlib.Path('/workspace/lwp2')
def rows(p):return [json.loads(x)for x in p.read_text().splitlines()if x.strip()]
variants={'parent':('p3-parent','p3-parent.jsonl'),'87028b08':('p3-child','p3-child.jsonl'),'old-stop-mutant':('p3-old-mutant','p3-mutant.jsonl'),'isolated-parent':('p4-isolated-parent','p4-isolated-parent.jsonl'),'isolated-child':('p4-isolated-child','p4-isolated-child.jsonl')}
logs={}
for label,(name,log)in variants.items():
 d=S/name;out=E/label;out.mkdir(exist_ok=True);status=json.loads((d/'status.json').read_text());assert status['child_exit_code']==0 and not status['deadline_expired'] and status['wall_seconds']<1800;shutil.copy2(d/'status.json',out/'status.json')
 for p in d.iterdir():
  if p.is_file()and p.name!='status.json':(out/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
 b=(S/log).read_bytes();(out/'events.jsonl.gz').write_bytes(gzip.compress(b,mtime=0));logs[label]=rows(S/log)
 stream=rows(d/'census.jsonl');assert stream[0]['latent_mode']=='full'and not stream[0]['checker_rejected'] and stream[-1]['status']=='relation_report'
assert logs['parent']==logs['87028b08']
assert logs['isolated-parent']==logs['isolated-child']
expected={'src/compiler/checker.ts:'+x for x in ['21733:22','21783:34','21849:20','21857:20','21895:16']}
summary={}
for label in ['parent','87028b08','isolated-parent','isolated-child']:
 r=logs[label];sites=[x for x in r if x['event']=='statement_end'];assert len(sites)==5 and {x['call']for x in sites}==expected
 good=[x for x in sites if x.get('tail_result')=='admitted'];bad=[x for x in sites if 'body_error'in x];assert len(good)==4 and len(bad)==1 and bad[0]['call'].endswith('21783:34')and 'Map or Set interface'in bad[0]['body_error']
 combined=next(x for x in r if x['event']=='combined');assert combined['result']=='admitted'and combined['saved_sites']==4
 origins=next(x for x in r if x['event']=='generatorFunctions'and x['active']=='combined-five-source-statements');assert len(origins['functions'])==3 and {x['yields']for x in origins['functions']}=={'an object'}
 assert not any('a step names the one function'in str(x)for x in r)
 summary[label]={'statement_ir_admitted':4,'body_blocked':1,'combined_generator_functions':origins['functions'],'combined':combined}
m=next(x for x in logs['old-stop-mutant']if x['event']=='combined');assert 'a step names the one function'in m['error']
summary['obsolete_stop_mutant']=m;summary['positive_five_site_pin']='blocked: inline yield does not lower'
(E/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
text='''# Full-mode generator pin

Pins: parent **ae75c13b3** (`87028b08^`) and **87028b0889b4431a08c31ca70bc294fa0fd746f5**. Their compiler, census and adaptation inputs are identical; the child only edits the generator mutant report. As in REPORT.md, separate executions reuse the identical pinned compiler/adapted tree, not the current-main compiler. The prior `evidence/pin-trees.json` establishes the equal trees.

**Four targeted statements lower and their iterable views are admitted on both pins. The fifth is blocked before the tail at its inline yield.** No real run returns the obsolete single-function/multi-function-view stop. The requested *all-five-bodies-lower* acceptance condition is therefore still **not met**; this is not called a passing five-site pin.

| Original checker.ts call site | Parent, LATENT_FULL | 87028b08, LATENT_FULL |
|---|---|---|
| 21733:22 | statement IR emitted; checkIterableViews admits | same |
| 21783:34 | body stops at yield, 21784:29; no statement IR | same |
| 21849:20 | statement IR emitted; checkIterableViews admits | same |
| 21857:20 | statement IR emitted; checkIterableViews admits | same |
| 21895:16 | statement IR emitted; checkIterableViews admits | same |

The blocked site's exact reason is:

> a Map or Set interface implemented by a value that isn't a Map or Set

It arises at `yield elem` in the inline generator, through the container-fit guard (`container_fit.go`), before checkIterableViews. It repeats with preflight state isolated too. This is a different earlier wall, not a `generatorFunctions` differing-yields result or an observed regression to the obsolete multi-function gate. No guard was disabled to manufacture a fifth lowered statement.

## What was actually run

`LATENT_FULL=1` is asserted in the stream header. The root is the original parent-adapted tsc entry (checker-clean); the file selector is checker.ts. Full mode splits to nested declarations and retains statement rollback/continuation. This avoids the old first-error createTypeChecker wall at 2971; the full-mode findings, including other rolled-back statements, are retained.

The probe uses one canonical lowering/ID space. It lowers elaborateElementwise, the named generator producer bodies, elaborateJsxComponents, elaborateArrayLiteral and elaborateObjectLiteral, with full-mode continuation. Named producers have real generator flags, object Yields and emitted yield nodes. Signature-only producer stubs were an initial invalid setup and were not counted as lowered generators.

At the end of each of the five exact source statements, before its enclosing full-mode rollback handler returns, the probe records whether raw statement lowering produced IR or an error. On a disposable state snapshot it attaches that statement IR, retains the lowered consumer/producer functions, removes unrelated caller-body roots, runs class/accessor/ToPrimitive completion and **the tail prefix through checkIterableViews**. It does not run later ownership/CFG passes or claim whole-program acceptance. The statement's caller function context is recorded separately from the enclosing demand-unit label. Check/dispatch rewrites never leak into the live full-mode lowering.

This source-root filtering matters: retaining the unrelated `generateJsxChildren` call at 21762 in an already completed caller otherwise masks the requested sites with “whose one use isn't a for-of over it”. That setup result and its cause were identified before the final isolated-statement runs. The five target statements are source-derived IR, not invented test calls.

The four successful statements are also joined as Main in the **same** canonical ID space, with their actual producer bodies and the consumer's actual iterable view. We do not concatenate unrelated census-unit ID spaces. This combined check reaches `generatorFunctions` with:

| Producer | Yields representation |
|---|---|
| generateJsxAttributes | an object |
| generateLimitedTupleElements | an object |
| generateObjectLiteralElements | an object |

The two tuple sites share one producer. The blocked inline generator is absent; generateJsxChildren is a lowered producer but its unrelated call is outside these five roots. With three equal-yield functions, the combined view is **admitted on both pins**, exercising the multi-function path. Individual successful-statement probes each also admit their view. There are no differing-yield words in these normal runs because the observed named origins agree.

## Regression sensitivity and limits

A measurement-only negative control reinstates the obsolete condition when `len(functions)>1`, with its exact S2c-1 words:

> a view that generators of more than one function feed (a step names the one function whose throws it may leave) (#qk8rztp S2c-1)

The combined check turns red under that control. Thus the successful multi-origin observation would detect the old gate; it is not merely five one-origin checks that could never exercise it. Identify this regression by the condition/text, not a stale physical line number after S3c-2 moved the code. The normal parent/child results and isolated-preflight parent/child results are each event-for-event identical.

The full-mode consumer/callers contain other rolled-back failures; admitting these scoped views does not certify the entire functions or tsc program. The inline storage-fit wall leaves the exact requested all-five positive pin blocked. This report neither suppresses it nor substitutes the four-site result for five.

## Evidence

`followup/evidence/pin/` contains per-revision complete full-mode streams, statement/generatorFunctions events, child exit/time/memory records, isolated-preflight controls and the old-stop mutant. Every child finishes normally within its 1,700-second bound; statement errors are data, not hidden process exits. `followup/pin-report.py` asserts all five attempted sites, four admissions, the exact blocked fifth site, three equal-yield origins, unchanged revisions and the negative control before writing this report. Go scratch evidence is `followup/pin.go.txt`; overlay and bounded-wrapper sources are alongside it. No production source changes are committed.
'''
(R.parent/'PIN.md').write_text(text)
print(json.dumps({'parent_child_same':True,'four_admitted':True,'inline_blocked':True,'multi_origin_admitted':True,'old_stop_mutant_red':True},indent=2))

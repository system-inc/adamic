from pathlib import Path
import json,hashlib,shutil,subprocess
s=Path('/workspace/scratch/arrayof');repo=Path('/workspace/adamic');gate=json.loads((s/'RECONCILED.json').read_text());assert gate['complete'] and gate['counterexamples']==0 and gate['counts']['non-negative integer']==gate['returnedOrThrew'];assert all(gate['counts'][c]==0 for c in ['negative','fractional','NaN','infinite'])
search=subprocess.run(['rg','-n','arrayOf','stage3/adapt','--glob','adapt.cjs','--glob','*owners*.json','--glob','*sites*.json'],cwd=repo,capture_output=True,text=True);(s/'ownership-final.log').write_text(search.stdout+search.stderr);assert search.returncode==1
out=repo/'stage3/adapt/99-correctness-core-arrayof';out.mkdir(exist_ok=False);before=json.loads((s/'instrument.json').read_text())['before'];assert (s/'before/src/compiler/core.ts').read_text().count(before)==1
after=before.replace('const result = new Array(count);','const result: T[] = [];').replace('result[i] = f(i);','result.push(f(i));');assert before!=after and before.count('\n')==after.count('\n');hash=lambda x:hashlib.sha256(x.encode()).hexdigest();pin=dict(file='core.ts',owner='arrayOf',beforeDeclarations=1,afterDeclarations=1,before=before,after=after,ownerBeforeSha256=hash(before),ownerAfterSha256=hash(after),site='core.ts:1339:12',reason='a member typed any ([]) related to T at the return site');(out/'owners.json').write_text(json.dumps([pin],indent=2)+'\n');shutil.copyfile(s/'adapter-draft.cjs',out/'adapt.cjs');shutil.copyfile(s/'contract-draft.cjs',out/'contract.cjs')
for name in ['prepare-baseline.py','prepare-measure.py','launch-after-baseline.sh','count-pieces.py','preload.cjs','reconcile.py','verify-baseline.py','measurement-scope.cjs','witness.cjs','census.py','build-census.py','compare.cjs','compare-census.py','replay.py','bound.py','pack.py','build-adaptation.py']:shutil.copyfile(s/name,out/name)
(out/'COUNT.json').write_text(json.dumps(gate,indent=2)+'\n');(out/'README.md').write_text('''# arrayOf builds its result by push

JavaScript-changing established-shape adaptation, task b5w3ycg, site F2 at
adapted core.ts:1339:12. The whole arrayOf owner is pinned before and after.
The result is explicitly T[] and grows by push, removing the any[] return
relation without a cast or a signature change.

Ruling: Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts.

The count gate is complete before adapter construction. COUNT.json retains
all five input classes, the independent raw-value reclassification, exact
value checks, test multiset and baseline diff agreement with stage3/lane/run.sh.
Only non-negative integer counts occur. Negative, fractional, NaN and infinite
counts each occur zero times. Node witnesses for 0 and 3 agree; -1, 1.5 and NaN
show the documented out-of-count divergence. Full after-apply proofs are pending.
''');(out/'PROOF.md').write_text('# Proof checkpoint\n\nMeasurement gate passes. Full after apply and paired proofs are pending.\n')
line='- arrayOf, 99-correctness-core-arrayof: JavaScript-changing; Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts. Whole-lane count '+str(gate['returnedOrThrew'])+' non-negative integer calls, 0 counterexamples; negative, fractional, NaN and infinite calls each 0. Node -1, 1.5 and NaN differ outside the counted domain; see the adapter README.\n';p=repo/'stage3/upstream/LEDGER.md';old=p.read_text();assert '99-correctness-core-arrayof' not in old;p.write_text(old+('' if old.endswith('\n\n') else '\n')+line);print(str(out),flush=True)

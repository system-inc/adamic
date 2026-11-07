from pathlib import Path
import gzip,hashlib,json,shutil,subprocess,tarfile
S=Path('/workspace/scratch/sys-walls');ROOT=Path('/workspace/adamic');out=Path('/workspace/sys-recordreplay-report/stage3/census/sys-recordreplay');summary=json.loads((out/'COUNT.json').read_text());reconciled=json.loads((out/'reconciliation.json').read_text());assert summary['complete'] and reconciled['independentReconciliation']=='pass'
h=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
prepared=S/'recordreplay';frozen=json.loads((S/'recordreplay-frozen.json').read_text());current={str(p.relative_to(prepared)):h(p) for base in ['src','built/local'] for p in sorted((prepared/base).rglob('*')) if p.is_file()};assert frozen==current
baseline=S/'before';public=[]
for name in ['typescript.d.ts','tsserverlibrary.d.ts','typescript.internal.d.ts']:
 a=baseline/'built/local'/name;b=prepared/'built/local'/name;assert a.read_bytes()==b.read_bytes();public.append({'file':name,'byteIdentical':True,'sha256':h(a)})
(out/'declarations.json').write_text(json.dumps(public,indent=2)+'\n')
packed=S/'recordreplay-final-packed';packed.mkdir()
logs=['apply-before.log','upstream-dependencies.log','recordreplay-instrument.log','recordreplay-build.log','recordreplay-harness.log','recordreplay-prepare.log','recordreplay-seeds.log','recordreplay-lane.log','recordreplay-lane-resume.log','recordreplay-piece-04-check.log','recordreplay-piece-04-check-fixed.log','commands.jsonl','recordreplay-frozen.json','shard-plan-check.log']
for name in logs:shutil.copyfile(S/name,packed/name)
for name in ['sys-timers-setup.log','sys-timers-api.log','sys-timers-source.log']:shutil.copyfile(Path('/workspace/scratch')/name,packed/name)
shutil.copytree(S/'recordreplay-seeds',packed/'seeds')
for side,tree in [('baseline',baseline),('instrumented',prepared)]:
 manifest={str(p.relative_to(tree/'src')):h(p) for p in sorted((tree/'src').rglob('*')) if p.is_file()};(packed/(side+'-source-manifest.json')).write_text(json.dumps(manifest,indent=2)+'\n');directory=packed/'source'/side;directory.mkdir(parents=True);shutil.copyfile(tree/'src/compiler/sys.ts',directory/'sys.ts')
for name in ['shard-plan.json','expected.json']:shutil.copyfile(ROOT/'stage3/lane'/name,packed/name)
for name in ['run.py','recordreplay-pieces.py','recordreplay-reconcile.py','recordreplay-prepare.py','prepare-recordreplay.cjs','recordreplay-preload.cjs','recordreplay-seeds.cjs','package-recordreplay.py']:
 assert (out/name).read_bytes()==(S/name).read_bytes(),name
provenance={'base':'e8f3723bd41abcebedabfcc7546fdf104a9bbf08','typescript':subprocess.check_output(['git','-C','/workspace/typescript-source','rev-parse','HEAD'],text=True).strip(),'node':subprocess.check_output(['node','--version'],text=True).strip(),'scripts':{name:h(out/name) for name in ['run.py','recordreplay-pieces.py','recordreplay-reconcile.py','recordreplay-prepare.py','prepare-recordreplay.cjs','recordreplay-preload.cjs','recordreplay-seeds.cjs','package-recordreplay.py']},'apiParserPackage':h(ROOT/'stage3/api/node_modules/typescript/package.json'),'sourceScope':json.loads((out/'source-scope.json').read_text()),'frozenInputsUnchangedAfterAllUnits':True,'seedEvaluationsExcludedFromOracleCount':True};(out/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n');(packed/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
(packed/'SHA256SUMS').write_text(''.join(h(p)+'  '+str(p.relative_to(packed))+'\n' for p in sorted(packed.rglob('*')) if p.is_file()))
with tarfile.open(out/'preparation.tar.gz','w:gz') as tf:tf.add(packed,arcname='preparation')
(out/'preparation.sha256').write_text(h(out/'preparation.tar.gz')+'  preparation.tar.gz\n')
owner=gzip.decompress((out/'getNodeSystem.ts.txt.gz').read_bytes());assert hashlib.sha256(owner).hexdigest()=='9d0f66d64f8d6c9a713b73b514429f485254c918e26791dae9540331a4aa357d'
values='\n'.join('| '+json.loads(k)['type']+' | '+str(v)+' |' for k,v in summary['actualValues'].items());maxpiece=max(p['seconds'] for p in summary['pieces'])
report=f'''# sys.ts recordreplay: complete lane measurement

Task b5w3ycg. Base origin/main `e8f3723bd41abcebedabfcc7546fdf104a9bbf08`; pinned TypeScript 6.0.3 source `050880ce59e30b356b686bd3144efe24f875ebc8`. Measurement only; no adaptation or host-axiom row is proposed.

All **398 lane units** ran exactly once in 20 pieces. Every ordinary identity, API, baseline and count guard passed. Totals are **106366 passes, one expected failure, zero skips**. The underlying Node exit 1 for this expected failure is retained; its ordinary shard guard exits 0 after approving that expected result. The sole sanctioned failure is `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.

| Expression result | Evaluations |
| --- | ---: |
| Truthy | {summary['truthy']} |
| Falsy | {summary['falsy']} |
| Total | {summary['evaluations']} |

| Actual process.recordreplay value | Evaluations |
| --- | ---: |
{values}

The actual value above was captured at the original property read. Raw rows also record whether the receiver is Node's process and whether it has an own property. Reconciliation records these distributions without assuming a value for an unobserved branch. This establishes the measured lane's observations; Compiler judges whether an adaptation is established shape or whether a host-axiom row with Language is appropriate.

## Method and validation

The sole scratch source change replaces exactly `!!(process as any).recordreplay` with `!!(globalThis as any).__sysRecordReplay((process as any).recordreplay, process)`. It preserves the preceding logical-or short circuit and performs one property read at the same point. The observer records the actual value and returns that exact value before Boolean conversion. It is installed in the Node global and the three VM context entry paths. Observer-ready events distinguish an active observer with zero evaluations from missing activation. Missing instrumentation throws instead of returning a quiet fallback.

Independent seeds exercise truthy and falsy values, a getter read exactly once, a preceding true short circuit, and all three VM entry paths. Seed observations are stored separately and excluded from the oracle count. The baseline and instrumented source manifests compare all 746 source files; only this expression in sys.ts differs. The complete prepared source/build manifest stays unchanged across every piece and after the final unit.

The unchanged stage3/lane shard.py and shard-plan.json execute the whole run.sh oracle inventory. Four concurrent units use the same 90 second shard deadline as the record-table measurement; the upstream 40 second test timeout, selected tests and expected results are unchanged. Each piece retains raw process observations, child exits, full ordinary shard reports and logs, and is committed and plainly pushed. The largest piece took {maxpiece:.3f} seconds. No command reached the 30 minute no-output stop, and no unit was cut short.

| Proof | Result |
| --- | --- |
| Required environment setup and pinned source | All accepted children exit 0 |
| Full unchanged apply --build-product | Child exit 0 |
| Instrumented compiler and test harness build | Children exit 0 |
| Observer seeds | Child exit 0 |
| generate_shards --check | Child exit 0 |
| Every ordinary lane shard | Child exit 0; all expected identity/API/baseline/count guards pass |
| Public declarations and typescript.internal.d.ts | Byte-identical to the unchanged baseline |
| Independent reconciliation | Reopens every raw stream and shard report, verifies archive hashes, exact unit inventory and complete counts |
| Source/build input freeze | Exact hashes unchanged through all 398 units |

The fourth piece's first packaging check exited 2 because the exact raw owner artifact used CRLF. Retaining those exact bytes as getNodeSystem.ts.txt.gz resolved the whitespace check; its LF quote below is complete. The already completed piece was pushed, then execution resumed at unit 81. No unit was rerun and no oracle result changed. That check failure and its successful correction are retained alongside the actual child exits.

COUNT.json is the cumulative measured ledger; reconciliation.json independently rebuilds its counts from raw archives and reopens all 398 ordinary reports. Every piece archive has its adjacent SHA-256. preparation.tar.gz retains setup, full apply, build, seed and command/child-exit logs; both source manifests; the prepared source/build hash manifest; exact sys.ts input bytes; the unchanged lane plan; and seed traces. Its internal SHA256SUMS and preparation.sha256 verify those bytes. Scripts, provenance.json and declarations.json identify the exact measured work. No result is submitted to Loom.

## Who pins the enclosing owner

**93-builtin-imports and 99-wasm-crypto-import both pin the whole getNodeSystem function**, including the recordreplay initializer, with before and after SHA-256 values in their proofs/targets.json files. Their adapt.cjs files validate the whole owner. owner-pins.json retains both manifests. The final full-apply function below has raw CRLF SHA-256 `9d0f66d64f8d6c9a713b73b514429f485254c918e26791dae9540331a4aa357d`, exactly the crypto adaptation's after-owner pin. getNodeSystem.ts.txt.gz preserves those raw bytes; this full quote normalizes only line endings to LF.

```ts
'''+owner.decode().replace('\r\n','\n')+'\n```\n'
assert '\u2013' not in report and '\u2014' not in report
(out/'REPORT.md').write_text(report)
print(json.dumps({'units':len(summary['completedUnits']),'evaluations':summary['evaluations'],'truthy':summary['truthy'],'falsy':summary['falsy'],'actualValues':summary['actualValues'],'files':len(list(out.iterdir()))}),flush=True)

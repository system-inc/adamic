from pathlib import Path
import gzip,shutil,json,hashlib
w=Path('/workspace/scratch/abrq61p');e=Path('/workspace/adamic/stage3/census/deferred-reach/evidence/steps2-3');e.mkdir(parents=True,exist_ok=True)
compressed=['head.jsonl','head-entry.jsonl','head-requests.jsonl','classification-references.json','full-core.jsonl','full-core-utilities.jsonl','pair-identity.json','head-times.jsonl','head-entry-times.jsonl','base-times.jsonl','base-entry-times.jsonl','probe-requests.jsonl','callback-bindings.json']
for prefix in ['caller-probe','caller-probe-extra','caller-probe-callback']:
 compressed += [prefix+suffix for suffix in ['.jsonl','-stages.jsonl','-requests.jsonl','-audit.json']]
if (w/'full-core-callers-result.json').exists():compressed+=['full-core-callers.jsonl','full-core-callers-times.jsonl']
for n in compressed:(e/(n+'.gz')).write_bytes(gzip.compress((w/n).read_bytes(),mtime=0))
plain=['corpus-identity.json','meter-identity.json','head-reader-validation.log','meter-reader-tests.log','pair-tests.log','reader-inventory-search.txt','latent-readers.txt','caller-probe-units.txt','caller-probe-extra-units.txt','caller-probe-callback-units.txt','probe-binding-counts.log','probe-computed-options.log','classification-references.log','binary-overlay-hashes.json','full-core-callers-files.json','full-core-callers-result.json','callback-bindings.log','callback-bindings.exit']
for n in plain:
 if (w/n).exists():shutil.copyfile(w/n,e/n)
for n in ['head-directory','head-entry','caller-probe','caller-probe-extra','caller-probe-callback','full-core','full-core-utilities','full-core-callers','probe-binding-counts','probe-computed-options','classification-references','witness-mutant','head-audit','statecopy-whole']:
 for suffix in ['.log','.exit']:
  if (w/(n+suffix)).exists():shutil.copyfile(w/(n+suffix),e/(n+suffix))
for name in ['classification-model.py','classify-references.cjs','callback-bindings.cjs','audit-caller-probes.py','prepare-caller-probes.py','report.py','validate-head.py','meter-identity.py','run.py','full-callers.py','archive-report.py']:
 shutil.copyfile(Path('/workspace/scratch')/('abrq61p-'+name),e/name)
for p in (w/'probe').glob('*.go.txt'):shutil.copyfile(p,e/p.name)
invocation=json.loads((e/'invocations.json').read_text())
invocation['probes']['unit_files']=['caller-probe-units.txt','caller-probe-extra-units.txt','caller-probe-callback-units.txt']
invocation['broader_full_attempt']=dict(script='full-callers.py',selection='full-core-callers-files.json',result='full-core-callers-result.json',memory_guard_bytes=14*1024**3,time_budget_seconds=1200)
(e/'invocations.json').write_text(json.dumps(invocation,indent=2)+'\n')
(e/'SHA256.json').write_text(json.dumps({p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(e.iterdir()) if p.name!='SHA256.json'},indent=2)+'\n')
print('evidence files',len(list(e.iterdir())),'bytes',sum(p.stat().st_size for p in e.iterdir()))

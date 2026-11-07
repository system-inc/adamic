from pathlib import Path
import hashlib,json
s=Path('/workspace/scratch/arrayof');assert (s/'apply-before-proof.exit').read_text().strip()=='0';fresh=s/'before-proof';measured=s/'before';manifest={};different=[]
for p in sorted(fresh.rglob('*')):
 rel=p.relative_to(fresh)
 if '.git' in rel.parts or not p.is_file():continue
 h=hashlib.sha256(p.read_bytes()).hexdigest();manifest[str(rel)]=h
 q=measured/rel
 if not q.is_file() or hashlib.sha256(q.read_bytes()).hexdigest()!=h:different.append(str(rel))
assert not different,different
(s/'fresh-baseline-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n');(s/'fresh-baseline-equivalence.json').write_text(json.dumps(dict(fullApplyExit=0,filesCompared=len(manifest),differences=[],failedAttemptRetained=True,failedAttempt='adapter phase complete, product packaging failed after an early build; not counted as full apply proof'),indent=2)+'\n')
p=s/'baseline-artifact/ready.json';ready=json.loads(p.read_text());ready['phases']['apply']=dict(exit=0,log=str(s/'apply-before-proof.log'),fullFreshApply=str(fresh),allAppliedInputFilesByteIdentical=True);ready['reuse']='fresh complete baseline full apply verified byte-for-byte against every applied input in the built baseline';p.write_text(json.dumps(ready,indent=2)+'\n');print(len(manifest),'input files identical; artifact apply provenance corrected',flush=True)

phasefile=s/"before-phases.json"
phases=json.loads(phasefile.read_text()); phases["apply"]=ready["phases"]["apply"];phasefile.write_text(json.dumps(phases,indent=2)+"\n")

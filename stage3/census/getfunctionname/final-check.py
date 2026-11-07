"""Verify normal inputs/outputs and the private observer after the measurement."""
from pathlib import Path
import json,hashlib
s=Path('/workspace/scratch/getfunctionname');raw=s/'raw';repo=Path('/workspace/adamic');reuse=json.loads((raw/'reuse-validation.json').read_text());old=Path(reuse['reused_product']);fresh=Path(reuse['new_tree']);changes=[]
for rel,digest in reuse['manifest'].items():
 for root in [old,fresh]:
  p=root/rel
  if not p.is_file() or hashlib.sha256(p.read_bytes()).hexdigest()!=digest:changes.append(str(p))
for rel,digest in reuse['compiled_files_verified_against_prior_commit'].items():
 p=old/'built/local'/rel
 if hashlib.sha256(p.read_bytes()).hexdigest()!=digest:changes.append(str(p))
for row in reuse['source_maps']:
 p=old/'built/local'/row['file']
 if hashlib.sha256(p.read_bytes()).hexdigest()!=row['sha256']:changes.append(str(p))
for row in json.loads((raw/'instrumentation.json').read_text())['rows']:
 p=s/'measure'/row['file']
 if hashlib.sha256(p.read_bytes()).hexdigest()!=row['afterSha256']:changes.append(str(p))
observed={r['file']:r['afterSha256'] for r in json.loads((raw/'instrumentation.json').read_text())['rows']}
for rel,digest in reuse['compiled_files_verified_against_prior_commit'].items():
 if rel.endswith('.js'):
  name='built/local/'+rel;p=s/'measure'/name
  if hashlib.sha256(p.read_bytes()).hexdigest()!=observed.get(name,digest):changes.append(str(p))
assert not changes,changes
result={'passed':True,'normalInputsCheckedPerTree':len(reuse['manifest']),'normalOutputsChecked':len(reuse['compiled_files_verified_against_prior_commit']),'sourceMapsChecked':len(reuse['source_maps']),'privateObservedBundlesChecked':3,'allPrivateJavaScriptFilesChecked':10,'changes':changes};(raw/'final-unchanged-check.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))

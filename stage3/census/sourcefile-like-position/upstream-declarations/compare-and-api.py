import difflib, hashlib, importlib.util, json, os, sys
from pathlib import Path
sys.dont_write_bytecode=True
repo=Path('/workspace/scratch/sourcefile-like-position')
out=repo/'stage3/census/sourcefile-like-position/upstream-declarations'
main=Path('/workspace/scratch/sourcefile-like-position-main-tree-2')
trial=Path('/workspace/scratch/sourcefile-like-position-trial-tree')
report={'files':{},'status':'pass'}
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
    a,b=[(tree/'built/local'/name).read_bytes() for tree in [main,trial]]
    report['files'][name]={'main_sha256':hashlib.sha256(a).hexdigest(),'candidate_sha256':hashlib.sha256(b).hexdigest(),'bytes':len(a),'identical':a==b}
    if a!=b:report['status']='fail'
(out/'comparison.json').write_text(json.dumps(report,indent=2)+'\n')
assert report['status']=='pass', 'declaration bytes differ'
results=out/'api-observation'
results.mkdir()
(results/'adapted-tree').symlink_to(trial,target_is_directory=True)
(results/'oracle').mkdir()
reference=trial/'tests/baselines/reference/api/typescript.d.ts'
local=trial/'tests/baselines/local/api/typescript.d.ts'
local.parent.mkdir(parents=True,exist_ok=True)
local.write_bytes((trial/'built/local/typescript.d.ts').read_bytes().replace(b'\r\n',b'\n'))
(results/'oracle/baseline.diff').write_text(''.join(difflib.unified_diff(reference.read_text().splitlines(True),local.read_text().splitlines(True),fromfile='reference/api/typescript.d.ts',tofile='local/api/typescript.d.ts')))
sanction=repo/'stage3/lane/sanctioned-api.json'
before=sanction.read_bytes()
spec=importlib.util.spec_from_file_location('lane_check',repo/'stage3/lane/check.py')
module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
errors,details=module.api_check(results,json.loads(before))
assert sanction.read_bytes()==before
api={'status':'fail' if errors else 'pass','errors':errors,'details':details,'sanctioned_api_sha256':hashlib.sha256(before).hexdigest(),'sanctioned_api_changed':False}
(out/'api-check.json').write_text(json.dumps(api,indent=2)+'\n')
print(json.dumps({'declarations':report,'api':api},indent=2))
assert not errors, errors

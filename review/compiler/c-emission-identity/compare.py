from pathlib import Path
import json,hashlib,subprocess,shutil
root=Path('/workspace/adamic/review/compiler/c-emission-identity');m=json.loads((root/'manifest.json').read_text());dirs={s:Path('/tmp/cemit-final-'+s) for s in ['parent','cut']}
expected={r['Name']+'.c' for r in m['programs']}
for side,d in dirs.items():assert {p.name for p in d.glob('*.c')}==expected,side
with (root/'emitted-c.diff').open('w') as log:
 code=subprocess.run(['diff','-ru',str(dirs['parent']),str(dirs['cut'])],stdout=log,stderr=subprocess.STDOUT,timeout=30).returncode
result={'diff_exit_code':code,'program_count':len(expected),'tests':{}}
for row in m['programs']:
 row['C']={}
 for side,d in dirs.items():
  b=(d/(row['Name']+'.c')).read_bytes();row['C'][side]={'bytes':len(b),'sha256':hashlib.sha256(b).hexdigest()}
for test,names in m['tests'].items():
 for side in dirs:
  d=Path('/tmp/cemit-test-trees')/side/test;d.mkdir(parents=True,exist_ok=True)
  for name in names:shutil.copyfile(dirs[side]/(name+'.c'),d/(name+'.c'))
 with (root/(test+'.diff')).open('w') as log:
  exit_code=subprocess.run(['diff','-ru','/tmp/cemit-test-trees/parent/'+test,'/tmp/cemit-test-trees/cut/'+test],stdout=log,stderr=subprocess.STDOUT,timeout=30).returncode
 result['tests'][test]={'program_count':len(names),'diff_exit_code':exit_code}
# A valid added C comment changes bytes only; no compiler or clang can catch it.
for side in ['before','mutant']:
 d=Path('/tmp/cemit-diff-control')/side;d.mkdir(parents=True,exist_ok=True);shutil.copyfile(dirs['cut']/'six-released.c',d/'six-released.c')
with Path('/tmp/cemit-diff-control/mutant/six-released.c').open('a') as f:f.write('/* planted emitted-byte disagreement */\n')
with (root/'comparator-mutant.diff').open('w') as log:
 mutant=subprocess.run(['diff','-ru','/tmp/cemit-diff-control/before','/tmp/cemit-diff-control/mutant'],stdout=log,stderr=subprocess.STDOUT,timeout=30).returncode
assert mutant==1;result['comparator_mutant_exit_code']=mutant
(root/'results.json').write_text(json.dumps(result,indent=2)+'\n');(root/'manifest.json').write_text(json.dumps(m,indent=2)+'\n');print(json.dumps(result,indent=2))

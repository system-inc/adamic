from pathlib import Path
import subprocess,os,json
r=Path('/workspace/scratch/write-file-construct-run');repo=Path('/workspace/scratch/write-file-construct-final');tree=r/'lane/adapted-tree';env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),TMPDIR=str(r/'tmp'));result=[]
for directory,name,files in [('99-correctness-write-file-view','actual-view-before',['src/compiler/sys.ts']),('99-correctness-write-file-construct','actual-construct-before',['src/compiler/sys.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts'])]:
 for f in files:
  target=r/name/f;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(subprocess.check_output(['git','-C',str(tree),'show',f'refs/stage3/{directory}-before:{f}']))
commands=[('93-actual-legacy-check',[repo/'stage3/adapt/93-builtin-imports/proofs/check.cjs',r/'actual93-before',r/'actual93-after']),('93-actual-legacy-scope-mutant',[repo/'stage3/adapt/93-builtin-imports/proofs/scope-mutant.cjs',r/'actual93-before',r/'actual93-after']),('view-actual-contract',[repo/'stage3/adapt/99-correctness-write-file-view/contract.cjs',r/'actual-view-before']),('construct-actual-contract',[repo/'stage3/adapt/99-correctness-write-file-construct/adapt.cjs','--proof',r/'actual-construct-before'])]
for label,args in commands:
 with (r/(label+'.log')).open('w') as log:p=subprocess.run(['node',*map(str,args)],env=env,stdout=log,stderr=subprocess.STDOUT)
 result.append({'name':label,'command':['node',*map(str,args)],'exit':p.returncode});(r/'actual-small-proofs.json').write_text(json.dumps(result,indent=2)+'\n');print(label,p.returncode,flush=True)

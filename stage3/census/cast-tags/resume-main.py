from pathlib import Path
import subprocess,os,json,sys,tempfile,importlib.util
r=Path('/workspace/scratch/cast-tags');work=r/'work';tree=r/'main-tree';e=r/'evidence'
spec=importlib.util.spec_from_file_location('apply',work/'stage3/apply.py');a=importlib.util.module_from_spec(spec);sys.path.insert(0,str(work/'stage3'));spec.loader.exec_module(a)
def git(*args,**kw):return subprocess.check_output(['git','-C',str(tree),*args],text=True,**kw).strip()
refs={x.split(' ',1)[0]:x.split(' ',1)[1] for x in git('for-each-ref','--format=%(refname) %(objectname)','refs/stage3').splitlines()}
(e/'interrupted-apply-refs.json').write_text(json.dumps(refs,indent=2)+'\n')
dirs,env=a.adaptation_inputs();done=[]
for d in dirs:
 if 'refs/stage3/'+d.name+'-after' in refs:done.append(d)
 else:break
assert len(done)>0 and done[-1].name=='99-correctness-core-multimap'
previous=refs['refs/stage3/'+done[-1].name+'-after'];print('RESUME',done[-1].name,previous,flush=True)
subprocess.run(['git','-C',str(tree),'read-tree','--reset','-u',previous],check=True)
with tempfile.TemporaryDirectory(prefix='resume-index-',dir=r/'tmp') as temp:
 indexenv=dict(os.environ,GIT_INDEX_FILE=str(Path(temp)/'index'));subprocess.run(['git','-C',str(tree),'read-tree',previous],env=indexenv,check=True)
 for d in dirs[len(done):]:
  print('START',d.name,flush=True);subprocess.run(['git','-C',str(tree),'update-ref','refs/stage3/'+d.name+'-before',previous],check=True)
  subprocess.run(['node',str(d/'adapt.cjs'),str(tree)],env=env,check=True)
  subprocess.run(['git','-C',str(tree),'add','--all'],env=indexenv,check=True);subprocess.run(['git','-C',str(tree),'add','--all','--force','src'],env=indexenv,check=True)
  previous=git('write-tree',env=indexenv);subprocess.run(['git','-C',str(tree),'update-ref','refs/stage3/'+d.name+'-after',previous],check=True);print('END',d.name,previous,flush=True)
refs={x.split(' ',1)[0]:x.split(' ',1)[1] for x in git('for-each-ref','--format=%(refname) %(objectname)','refs/stage3').splitlines()}
(e/'completed-apply-refs.json').write_text(json.dumps(refs,indent=2)+'\n')
rows=['# Patch set','','Resumed unchanged main apply after a runtime interruption from its last completed whole-tree ref.','','| Adaptation | Files | Lines added | Lines removed |','|---|---:|---:|---:|']
for d in dirs:
 before=refs['refs/stage3/'+d.name+'-before'];after=refs['refs/stage3/'+d.name+'-after'];nums=git('diff','--numstat',before,after).splitlines();adds=deletes=0
 for n in nums:
  x,y,_=n.split('\t',2);adds+=int(x);deletes+=int(y)
 rows.append(f'| {d.name} | {len(nums)} | {adds} | {deletes} |')
(tree/'patch-set.md').write_text('\n'.join(rows)+'\n');(e/'main-resume-status.json').write_text(json.dumps({'exit':0,'final_tree':previous,'resumed_after':done[-1].name})+'\n')

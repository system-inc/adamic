import os, pathlib, shutil, subprocess, time, json
repo=pathlib.Path('/workspace/adamic');s=pathlib.Path('/workspace/scratch/multimap');tree=s/'counted';out=repo/'stage3/adapt/99-correctness-core-multimap';begin=time.monotonic();records={}
shutil.copytree('/workspace/multimap-base',tree,symlinks=True,ignore=shutil.ignore_patterns('.git','node_modules','built'));(tree/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),NODE_OPTIONS='--require='+str(out/'measure-preload.cjs'),MM_BODIES=str(s/'bodies.json'))
for name,cmd in [('instrument',['node',str(out/'instrument.cjs'),str(tree),str(s/'bodies.json')]),('witnesses',['node',str(out/'witness.cjs'),str(s/'bodies.json'),str(s/'witnesses.json')]),('build',['npm','run','build']),('enumerations',['node',str(out/'instrument-js.cjs'),str(tree)])]:
 started=time.monotonic()
 with (s/(name+'.log')).open('w') as log:r=subprocess.run(cmd,cwd=tree,env={k:v for k,v in env.items() if name!='instrument' or k!='NODE_OPTIONS'},stdout=log,stderr=subprocess.STDOUT,timeout=1400)
 records[name]={'exit':r.returncode,'seconds':time.monotonic()-started};(s/'prepare-count.json').write_text(json.dumps(records,indent=2)+'\n');print(name,r.returncode,flush=True);assert r.returncode==0
(tree/'.git').symlink_to('/workspace/multimap-base/.git',target_is_directory=True)
(tree/'node_modules').unlink()
print('measurement tree ready',time.monotonic()-begin,flush=True)

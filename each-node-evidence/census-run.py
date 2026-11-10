from pathlib import Path
import os,subprocess,time,json,shutil,difflib,sys
s=Path('/workspace/scratch/each-node');repo=Path('/workspace/adamic');unit=repo/'stage3/adapt/99-correctness-core-each-node';env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),ADAMIC_TYPESCRIPT_SOURCE=str(s/'upstream'))
def run(name,args,cwd=repo,extra=None):
 start=time.monotonic();effective=dict(env,**(extra or {}))
 with (s/(name+'.log')).open('w') as f:p=subprocess.run([str(a) for a in args],cwd=cwd,env=effective,stdout=f,stderr=subprocess.STDOUT)
 record=dict(command=[str(a) for a in args],cwd=str(cwd),exit=p.returncode,seconds=time.monotonic()-start);(s/(name+'-result.json')).write_text(json.dumps(record,indent=2)+'\n');print(name,json.dumps(record),flush=True)
 if p.returncode:raise SystemExit(p.returncode)
 return record
mode='census'
if mode=='candidate':
 gate=json.loads((s/'reconciled.json').read_text());assert gate['passed'] and gate['totals']['exposed']==0
 before=(s/'before').resolve();after=s/'after';after.mkdir()
 for p in before.iterdir():
  if p.name in ['.git','built','node_modules']:continue
  if p.name=='src' or p.name=='scripts':shutil.copytree(p,after/p.name)
  elif p.is_dir():(after/p.name).symlink_to(p,target_is_directory=True)
  else:shutil.copy2(p,after/p.name)
 (after/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
 run('candidate-adapt',['node',unit/'adapt.cjs',after]);run('candidate-guards',['node',unit/'guards.cjs',before,after,s/'guards-result.json']);run('candidate-witness',['node',unit/'witness.cjs',before,after,unit/'witness-result.json']);run('candidate-build',['npm','run','build'],cwd=after)
 run('compare-built',['node','--stack_size=8192',s/'compare-built.cjs',before,after,s/'build-comparison.json']);run('after-instrument',['node',s/'instrument.cjs',after,s/'measure-after']);run('after-built-control',['node',s/'built-control.cjs',s/'measure-after/built/local/typescript.js','after'],extra=dict(NODE_OPTIONS='--require='+str(s/'preload.cjs'),EACH_NODE_COUNTS=str(s/'after-control-counts')));control=[json.loads(p.read_text()) for p in (s/'after-control-counts').glob('*.json')];assert len(control)==1 and tuple(control[0][k] for k in ['defined','exposed','throws'])==(1,1,0);run('unused-import',['node',s/'unused-import.cjs',after,s/'unused-import.json'])
 beforeText=(before/'src/compiler/debug.ts').read_bytes().decode();afterText=(after/'src/compiler/debug.ts').read_bytes().decode();(s/'source-hunks.patch').write_bytes(''.join(difflib.unified_diff(beforeText.splitlines(True),afterText.splitlines(True),fromfile='before/src/compiler/debug.ts',tofile='after/src/compiler/debug.ts')).encode())
 # The restoring mutant is the candidate src with exactly the ruled owner and its required import restored.
 shutil.copytree(after/'src',s/'mutant/src');(s/'mutant/node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True);(s/'mutant/src/compiler/debug.ts').write_bytes(beforeText.encode());(s/'restoring-mutant.patch').write_bytes(''.join(difflib.unified_diff(afterText.splitlines(True),beforeText.splitlines(True),fromfile='after/src/compiler/debug.ts',tofile='mutant/src/compiler/debug.ts')).encode())
 # Compare every normal JavaScript file and preserve every emitted hunk.
 emitted=[];hunks=[]
 for a in sorted((before/'built/local').rglob('*.js')):
  relative=a.relative_to(before/'built/local');b=after/'built/local'/relative;x=a.read_text();y=b.read_text()
  if x!=y:
   diff=list(difflib.unified_diff(x.splitlines(True),y.splitlines(True),fromfile='before/'+str(relative),tofile='after/'+str(relative)));hunks+=diff;emitted.append(dict(file=str(relative),hunks=[line.strip() for line in diff if line.startswith('@@')]))
 (s/'emitted-hunks.patch').write_text(''.join(hunks));(s/'emitted-hunk-inventory.json').write_text(json.dumps(emitted,indent=2)+'\n')
elif mode=='census':
 for side in ['mutant','after']:
  tree=s/('census-mutant' if side=='mutant' else side);clean={k:v for k,v in env.items() if not k.startswith('LATENT_')};clean.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/debug.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(s/(side+'-census-times.jsonl')))
  start=time.monotonic();args=[str(s/'latent'),str(tree/'src/compiler'),str(s/(side+'-census.jsonl'))]
  with (s/(side+'-census.log')).open('w') as f:p=subprocess.run(args,cwd=repo,env=clean,stdout=f,stderr=subprocess.STDOUT)
  result=dict(command=args,exit=p.returncode,seconds=time.monotonic()-start,flags={k:v for k,v in clean.items() if k.startswith('LATENT_')});(s/(side+'-census-result.json')).write_text(json.dumps(result,indent=2)+'\n');print(side,json.dumps(result),flush=True)
  if p.returncode:raise SystemExit(p.returncode)

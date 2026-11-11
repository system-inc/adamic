import pathlib,subprocess,sys,shutil,gzip,json
root=pathlib.Path('/workspace/overload-head'); rel='stage3/census/overload-result-10b6c33'; out=root/rel
out.mkdir(parents=True,exist_ok=True)
def copy(src,dst):
 p=pathlib.Path(src)
 if p.exists():
  if p.stat().st_size>1000000:
   with p.open('rb') as i,gzip.open(out/(dst+'.gz'),'wb') as o: shutil.copyfileobj(i,o)
  else: shutil.copyfile(p,out/dst)
for p in pathlib.Path('/tmp').glob('overload-*'):
 if p.is_file(): copy(p,p.name)
for label in ['parent','head']:
 d=pathlib.Path('/workspace/overload-build-'+label)
 for p in d.glob('*'):
  if p.is_file() and (p.suffix in ['.log','.json','.jsonl']): copy(p,label+'-'+p.name)
for name in ['run.py','build.sh','checkpoint.py']:
 copy('/workspace/overload-'+name,name)
if not (out/'REPORT.md').exists(): (out/'REPORT.md').write_text('# Overload result census\n\nCheckpoint: '+sys.argv[1]+'. Final census pending.\n')
def run(*args):
 p=subprocess.run(args,cwd=root,text=True,capture_output=True)
 print('child_exit='+str(p.returncode)+' '+args[0],flush=True)
 if p.returncode: print(p.stdout,p.stderr);sys.exit(p.returncode)
 return p.stdout.strip()
run('git','add','-f',rel)
tree=run('git','write-tree')
ref='refs/heads/typescript/overload-result-census-checkpoint'
parent=subprocess.run(['git','rev-parse','--verify',ref],cwd=root,text=True,capture_output=True)
sha=run('git','commit-tree',tree,'-p',parent.stdout.strip() if parent.returncode==0 else '10b6c33','-m','Census checkpoint: '+sys.argv[1])
run('git','update-ref',ref,sha)
run('git','push','origin',ref+':'+ref)
print('checkpoint='+sha)

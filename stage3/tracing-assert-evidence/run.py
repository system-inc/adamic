from pathlib import Path
import subprocess,os,json,time,sys
root=Path('/workspace/adamic');raw=Path('/workspace/scratch/baobab-evidence')
def run(name,args,cwd=root,env=None):
 start=time.monotonic()
 with (raw/(name+'.log')).open('w') as log:
  p=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT)
 result=dict(command=[str(x) for x in args],cwd=str(cwd),exit=p.returncode,seconds=time.monotonic()-start)
 (raw/(name+'.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
 if p.returncode:sys.exit(p.returncode)
if __name__=='__main__':
 mode=sys.argv[1]
 if mode=='build':
  side=sys.argv[2];tree=Path('/workspace/scratch/baobab-'+side)
  modules=tree/'node_modules'
  if not modules.exists():modules.symlink_to(root/'stage3/typescript/node_modules',target_is_directory=True)
  run(side+'-build',['npm','run','build'],cwd=tree)
 elif mode=='types':
  side=sys.argv[2];roots=sorted(Path('/workspace/scratch/baobab-'+side+'/src/compiler').rglob('*.ts'));assert len(roots)==79
  run(side+'-types',[raw/'adamic','types',*roots])
 elif mode=='census':
  side=sys.argv[2];tree=Path('/workspace/scratch/baobab-'+side)
  env=dict(os.environ,BAOBAB_CENSUS_ROOT=str(tree/'src/compiler/core.ts'))
  run(side+'-root-census',[raw/'root-census',tree/'src/compiler',raw/(side+'-root-census.jsonl')],env=env)

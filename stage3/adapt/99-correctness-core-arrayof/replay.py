from pathlib import Path
import subprocess,json,os,shutil,hashlib,sys,time
s=Path('/workspace/scratch/arrayof');repo=Path('/workspace/adamic');label=sys.argv[1];source=s/label;tree=s/('replay-'+label);shutil.copytree(source,tree,symlinks=True,ignore=shutil.ignore_patterns('node_modules'));(tree/'node_modules').symlink_to(s/'before/node_modules',target_is_directory=True)
def snapshot():
 return {str(p.relative_to(tree/'src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
initial=snapshot();(s/('replay-'+label+'-initial.json')).write_text(json.dumps(initial,indent=2)+'\n');rows=[];env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for d in sorted((repo/'stage3/adapt').iterdir()):
 if not d.is_dir():continue
 begin=time.monotonic()
 with (s/('replay-'+label+'.log')).open('a') as log:
  r=subprocess.run(['node',str(d/'adapt.cjs'),str(tree)],cwd=repo,env=env,capture_output=True,text=True,timeout=1700);log.write(d.name+'\n'+r.stdout+r.stderr)
 row=dict(adapter=d.name,exit=r.returncode,stdout=r.stdout.replace(str(tree),'TREE'),stderr=r.stderr.replace(str(tree),'TREE'));rows.append(row);(s/('replay-'+label+'.json')).write_text(json.dumps(rows,indent=2)+'\n');print(d.name,r.returncode,round(time.monotonic()-begin,3),flush=True)
(s/('replay-'+label+'-final.json')).write_text(json.dumps(snapshot(),indent=2)+'\n')

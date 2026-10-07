import argparse,json,subprocess,pathlib,re
parser=argparse.ArgumentParser(description="Record Node and two stage-0 fixture snapshots")
parser.add_argument("--feature-root",type=pathlib.Path,required=True)
parser.add_argument("--logs",type=pathlib.Path,required=True)
args=parser.parse_args()
args.logs.mkdir(parents=True,exist_ok=True)
root=pathlib.Path(__file__).resolve().parents[3]; feature=args.feature_root.resolve(); bucket='stage3/fixtures/nested-functions/'
(feature/'stage3/fixtures').mkdir(parents=True,exist_ok=True)
if not (feature/bucket).exists(): (feature/bucket).symlink_to(root/bucket,target_is_directory=True)
manifest=[]
for path in sorted((root/bucket).glob("[0-9][0-9]_*.a")):
 text=path.read_text()
 manifest.append({"file":path.name,"tsc":re.findall(r"^// From TypeScript 6.0.3, (.+)$",text,re.M),"reason":"a function inside a function (a closure)"})
results={}
for name,repo in [('main',root),('nested-functions',feature)]:
 rows=[]
 for e in manifest:
  e=e.copy();file=bucket+e['file'];log=args.logs/(name+'-'+e['file'])
  n=subprocess.run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',file],cwd=root,capture_output=True)
  e['node']={'stdout':n.stdout.decode(),'stderr':n.stderr.decode(),'exit':n.returncode}
  log.with_suffix('.node.log').write_bytes(n.stdout+n.stderr)
  binary=str(log)+'.bin'
  p=subprocess.run(['go','run','./cmd/adamic','build',file,'-o',binary],cwd=repo,capture_output=True)
  log.with_suffix('.build.log').write_bytes(p.stdout+p.stderr)
  diagnostic=(p.stdout+p.stderr).decode()
  if p.returncode==0:
   outcome='Compiles';what='';r=subprocess.run([binary],cwd=repo,capture_output=True);log.with_suffix('.native.log').write_bytes(r.stdout+r.stderr)
   if (r.stdout,r.stderr,r.returncode)!=(n.stdout,n.stderr,n.returncode): print('SILENT MISCOMPILE',name,file,repr(r.stdout),repr(n.stdout),flush=True)
  else:
   outcome='NotYet' if 'not yet:' in diagnostic.lower() or 'notyet' in diagnostic.lower() or "can't lower" in diagnostic.lower() else 'Refused' if 'refused:' in diagnostic.lower() or 'refuses' in diagnostic.lower() else 'Checker'
   what=diagnostic
  e['stage0']={'outcome':outcome,'what':what};rows.append(e)
  print(name,e['file'],outcome,repr(diagnostic[:350]),flush=True)
 results[name]=rows
(root/bucket/'status.json').write_text(json.dumps(results['main'],indent=2)+'\n')
(root/bucket/'nested-functions-status.json').write_text(json.dumps(results['nested-functions'],indent=2)+'\n')

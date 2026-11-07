"""Replay both complete trees, retaining every outcome and write, even refusals."""
import concurrent.futures, hashlib, json, os, pathlib, re, shutil, subprocess, sys
before_repo,after_repo,before_tree,after_tree,output=map(lambda x:pathlib.Path(x).resolve(),sys.argv[1:]);output.mkdir();hook=after_repo/'stage3/adapt/95-checker-axolotl-h/replay-hook.cjs'
def snapshot(tree):return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src').rglob('*') if p.is_file()}
def run(side,repo,source):
 tree=output/side;shutil.copytree(source,tree,symlinks=True,ignore=lambda d,names:[n for n in names if pathlib.Path(d)==source and n in ['.git','node_modules','built']]);(tree/'node_modules').symlink_to(after_repo/'stage3/typescript/node_modules',target_is_directory=True);logs=output/(side+'-logs');logs.mkdir();rows=[]
 for directory in sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir()):
  name=directory.name;old=snapshot(tree);writes=logs/(name+'.writes.json');env=dict(os.environ,NODE_PATH=str(after_repo/'stage3/api/node_modules'),REPLAY_TREE=str(tree),REPLAY_LOG=str(writes),NODE_OPTIONS='--require='+str(hook));
  with (logs/(name+'.stdout')).open('w') as stdout,(logs/(name+'.stderr')).open('w') as stderr:r=subprocess.run(['node',str(directory/'adapt.cjs'),str(tree)],cwd=repo,env=env,stdout=stdout,stderr=stderr,timeout=1380)
  new=snapshot(tree)
  def normalize(text):
   text=text.replace(str(tree),'TREE').replace(str(repo),'REPO').replace(str(after_repo),'REPO').replace(str(before_repo),'REPO');return re.sub(r'\bat .*\n','',text) if False else text
  row={'adapter':name,'exit':r.returncode,'stdout':normalize((logs/(name+'.stdout')).read_text()),'stderr':normalize((logs/(name+'.stderr')).read_text()),'changedFiles':sorted(k for k in old.keys()|new.keys() if old.get(k)!=new.get(k)),'filesWritten':sorted(item for line in writes.read_text().splitlines() for item in json.loads(line)) if writes.exists() else []};rows.append(row);(output/(side+'.json')).write_text(json.dumps(rows,indent=2)+'\n');print(side,name,r.returncode,flush=True)
 return rows
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 a=pool.submit(run,'before',before_repo,before_tree);b=pool.submit(run,'after',after_repo,after_tree);a,b=a.result(),b.result()
subprocess.run(['python3',str(after_repo/'stage3/adapt/99-correctness-core-multimap/replay-review.py'),str(before_repo),str(after_repo),str(before_tree),str(after_tree),str(output)],check=True)

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
left={r['adapter']:r for r in a};right={r['adapter']:r for r in b};new={'94-correctness-multimap','99-correctness-core-multimap'};assert right.keys()-left.keys()==new;different=[]
for name in left:
 if name=='95-checker-axolotl-h':
  x,y=left[name],right[name];assert x['exit']==y['exit']!=0;assert 'writer owner shape: createProgram' in x['stderr'] and 'writer owner shape: createProgram' in y['stderr'];assert x['changedFiles']==y['changedFiles']==[];different.append({'adapter':name,'result':'same existing createProgram owner refusal; no writes','amendment':'two call sites and one import gap'})
 else:assert left[name]==right[name],name
for name in new:assert right[name]['exit']==0 and not right[name]['changedFiles'] and not right[name]['filesWritten'],name
(output/'summary.json').write_text(json.dumps({'passed':True,'baselineAdapters':len(a),'adaptedAdapters':len(b),'identicalOtherAdapters':len(a)-1,'amendment':different,'newAdaptersReplayWithoutWrites':sorted(new)},indent=2)+'\n')

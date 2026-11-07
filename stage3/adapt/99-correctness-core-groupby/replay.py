import pathlib,subprocess,shutil,os,json,time,hashlib,concurrent.futures
R=pathlib.Path('/workspace/scratch/core-wave');env=dict(os.environ,NODE_PATH='/workspace/adamic/stage3/api/node_modules')
def manifest(tree):return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
for name,repo in [('clone',pathlib.Path('/workspace/adamic')),('groupby',pathlib.Path('/workspace/adamic-groupby'))]:
 own='99-correctness-core-'+name;trees={}
 for label,source in [('before',pathlib.Path('/workspace/core-before')),('after',pathlib.Path('/workspace/core-'+name))]:
  dest=R/('replay-'+name+'-'+label);shutil.copytree(source,dest,symlinks=True);trees[label]=dest
 records=[]
 for directory in sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir()):
  def run(label):
   tree=trees[label];before=manifest(tree);start=time.monotonic()
   p=subprocess.run(['node',str(directory/'adapt.cjs'),str(tree)],env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True,timeout=1700)
   after=manifest(tree);changed=[f for f in sorted(set(before)|set(after)) if before.get(f)!=after.get(f)]
   return label,{'exit':p.returncode,'stdout':p.stdout.replace(str(tree),'TREE'),'stderr':p.stderr.replace(str(tree),'TREE'),'changedSources':changed},time.monotonic()-start
  with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:results=list(pool.map(run,['before','after']))
  row={'adapter':directory.name,**{k:v for k,v,_ in results},'seconds':{k:t for k,_,t in results}};records.append(row);(R/('replay-'+name+'.json')).write_text(json.dumps(records,indent=2)+'\n')
  print(name,directory.name,row['before']['exit'],row['after']['exit'],flush=True)
  if directory.name!=own:assert row['before']==row['after'],row
  else:
   assert row['before']['exit']==row['after']['exit']==0,row
   assert json.loads(row['before']['stdout'])=={'files':1,'edits':1,'javascriptChanging':False},row
   assert json.loads(row['after']['stdout'])=={'files':0,'edits':0,'javascriptChanging':False},row
   assert row['before']['changedSources']==['src/compiler/core.ts'] and not row['after']['changedSources'],row
 assert manifest(trees['before'])==manifest(trees['after'])
 (R/('replay-'+name+'-result.json')).write_text(json.dumps({'pass':True,'adapters':len(records),'onlyDifferentOutcome':own,'finalSourceManifestsIdentical':True,'untouched30':next(r for r in records if r['adapter']=='30-indexed-reads')},indent=2)+'\n')

import hashlib,json,pathlib,shutil,subprocess,tarfile,tempfile
repo=pathlib.Path('/workspace/adamic');source=pathlib.Path('/workspace/scratch/or-predicate');destination=repo/'stage3/adapt/99-correctness-core-or-predicate'
active=[]
for side in ['before','after']:
 if not (source/('checker-'+side+'-result.json')).exists():active.append('checker-'+side)
with tempfile.TemporaryDirectory(prefix='or-predicate-evidence-') as temp:
 stage=pathlib.Path(temp)
 for p in source.iterdir():
  if not p.is_file() or p.suffix not in ['.json','.jsonl','.log','.py','.cjs']:continue
  if any(p.name.startswith(name) and not p.name.startswith(name+'-excluded-') for name in active):continue
  target=stage/p.name;shutil.copy2(p,target)
 for name in ['comparison','replay']:
  origin=source/name;target=stage/name;target.mkdir()
  for p in origin.rglob('*'):
   if not p.is_file():continue
   relative=p.relative_to(origin)
   if name=='replay' and relative.parts[0] in ['before','after']:continue
   q=target/relative;q.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(p,q)
 for origin in source.iterdir():
  if not origin.is_dir() or not origin.name.startswith('witness-'):continue
  for p in origin.rglob('*'):
   if p.is_file() and p.name!='native':
    q=stage/origin.name/p.relative_to(origin);q.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(p,q)
 compiler_diff=subprocess.check_output(['git','diff','origin/main','--','internal','cmd','go.mod','go.sum'],cwd=repo,text=True);assert not compiler_diff
 status={'main':'3214dd465e1155d7bf4d41653e95b29a3f1060a6','typescriptSource':'050880ce59e30b356b686bd3144efe24f875ebc8','productionCompilerDiffFromMain':compiler_diff,'binaryHashes':{name:hashlib.sha256((source/name).read_bytes()).hexdigest() for name in ['adamic','latent']},'backendWitnessAccepted':False,'backendBlockers':['retained or overload has a non-array rest parameter','instantiated helper has an unreified checked predicate claim of R1'],'originalAndMutantBinderRefusalsReproduced':True,'nodeSiteWitnessDifferences':0,'runningBestEffortCensuses':active,'laneCountReused':{'predicateCalls':2018000,'restCalls':168858,'valueDifferences':0},'excludedStartupNotes':['The initial preview was mistakenly applied to baseline. It was reversed and all 746 baseline source hashes verified before restarting the accepted checker census. The cancelled attempt is retained as excluded-preview.','A scratch file named types.py shadowed a Python standard module before the compiler started; renamed check-roots.py.','The first Node fixture attached its JSDoc comment to the wrong syntax context; a separate JavaScript @import node fixed fixture construction.','The first native CLI probe placed --sanitize before the source; corrected to the documented source-first form.','The first JS comparer preserved printer layout; structural AST comparison and explicit generated dead-export annotations now account for the complete output.','A first diff-report write ran in the read-only sandbox; the accepted diff emitter retained all child results and complete hunks.']}
 (stage/'proof-status.json').write_text(json.dumps(status,indent=2)+'\n')
 for side in ['before','after']:
  tree=pathlib.Path('/workspace/or-predicate-'+side);rows=[]
  for p in sorted((tree/'src').rglob('*')):
   if p.is_file():rows.append({'file':str(p.relative_to(tree)),'sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'bytes':p.stat().st_size})
  (stage/(side+'-source-manifest.json')).write_text(json.dumps(rows,indent=2)+'\n')
 manifest=[]
 for p in sorted(stage.rglob('*')):
  if p.is_file():manifest.append(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(stage)))
 (stage/'SHA256SUMS').write_text('\n'.join(manifest)+'\n')
 archive=destination/'PROOF-EVIDENCE.tar.gz'
 with tarfile.open(archive,'w:gz') as tar:
  for p in sorted(stage.rglob('*')):
   if p.is_file():tar.add(p,arcname=str(p.relative_to(stage)),recursive=False)
 with tarfile.open(archive,'r:gz') as tar:
  assert not any(n.endswith('.go') for n in tar.getnames())
  for line in manifest:
   digest,name=line.split('  ',1);assert hashlib.sha256(tar.extractfile(name).read()).hexdigest()==digest
 (destination/'PROOF-EVIDENCE.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  PROOF-EVIDENCE.tar.gz\n')
 (destination/'proof-status.json').write_text(json.dumps(status,indent=2)+'\n')
 print(json.dumps({'members':len(manifest)+1,'archiveBytes':archive.stat().st_size,'manifestVerified':True,'runningCensuses':active,'filesUnderDirectory':sum(p.is_file() for p in destination.rglob('*'))}))

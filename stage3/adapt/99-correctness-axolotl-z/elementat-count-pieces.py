import hashlib,json,os,shutil,signal,subprocess,tarfile,time
from collections import Counter,defaultdict
from concurrent.futures import ThreadPoolExecutor,as_completed
from pathlib import Path
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/elementat-b');out=repo/'stage3/adapt/99-correctness-axolotl-z';plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());units=plan['units'];results=s/'count-results';results.mkdir(exist_ok=True)
classes=['in-range non-negative integer','in-range negative integer','out-of-range integer','non-integer','array undefined']
summary={'ruling':{'task':'b5w3ycg','at':'2026-11-06 21:49Z','road':'b, fallback only, admitted only on a count'},'libRuling':{'task':'axochk','at':'2026-11-01 16:06Z','text':'tsc is checked against the lib its own tsconfig declares'},'unitInventory':len(units),'completedUnits':[],'pieces':[],'counts':dict.fromkeys(classes,0),'nonIntegerOffsetsIncludingUndefinedArrays':0,'callers':{c:{} for c in classes if c!=classes[0]},'oracleCounts':{'pass':0,'fail':0,'skip':0},'countScope':'one union of the unchanged 398 lane manifest shards; planted probes excluded','complete':False}
if (out/'ELEMENTAT-COUNT.json').exists():summary=json.loads((out/'ELEMENTAT-COUNT.json').read_text())
def snapshot():
 return {str(p.relative_to(s/'counted')):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((s/'counted'/base).rglob('*')) if p.is_file()}
frozen=snapshot();(s/'raw/counted-manifest.json').write_text(json.dumps(frozen,indent=2)+'\n')
for file in ['count-preload.cjs','instrument.cjs','prepare-count.py','count-pieces.py']:shutil.copyfile(s/file,out/('elementat-'+file))
(out/'README.md').write_text('''# 99z elementAt road (b), measurement first\n\nStep A in progress. The existing adapter remains the Step 1 no-op.\n\nAdamic, task axochk, Nov 1 16:06Z: "tsc is checked against the lib its own tsconfig declares."\nAdamic, task b5w3ycg, Nov 6 21:49Z: road (b), retain only elementAt's fallback\nas a JavaScript-changing adaptation, admitted only on a count. Lib es2022 is refused.\n\nELEMENTAT-COUNT.json records each completed, committed piece of the single\n398-shard lane union. Each shard runs through the unchanged stage3/lane/shard.py\nand its ordinary identity, baseline and API guards. The first two completed pieces use two units concurrently and the normal\n30-second budget. After deadline failures, subsequent measurement pieces use\nfour units and the runner-supported 90-second deadline. Tests and all identity,\nbaseline and API guards are unchanged. Failed attempts and diagnostic probes\nare excluded from the census and retained separately. Each piece has at most\n20 units and an\nindependent wall bound below 25 minutes. Raw call lines retain offset, length,\nclass and the required callers. Recorder v2 captures stacks only outside the\nfirst class and resolves source maps offline; a 263-call comparison proves the\nsame census fields and caller identities as v1. Seeded calls use a separate excluded directory.\nThe wrapper is installed only in the scratch counted source, never in adapt.cjs.\nThe raw source and built-file manifest is checked after every piece.\n\nThe earlier MEASURE-ELEMENTAT.md remains historical. Its missing lib ruling is\nnow supplied above. Step B requires the completed count and a zero non-integer\ncount. Native at and the fallback differ for fractional offsets and NaN; the\nrequested equal witnesses for absent non-integer cases need clarification.\n''')
def git(*args):
 with (s/'git-count.log').open('a') as log:subprocess.run(['git',*args],cwd=repo,stdout=log,stderr=log,check=True)
def run(unit,dir):
 name=unit['name'];env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/'counted'),STAGE3_RESULTS=str(results),EA_CALL_DIR=str(dir/'calls'),EA_SHARD=name,STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(s/'count-preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1')
 with (dir/(name+'.log')).open('w') as log:
  process=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),name,str(results)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=process.wait(timeout=100)
  finally:
   try:os.killpg(process.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   process.wait()
 report=json.loads((results/name/'report.json').read_text());return unit,code,report
for start in range(0,len(units),20):
 if start<len(summary['completedUnits']):continue
 begun=time.monotonic();piece=start//20+1;suffix='-retry' if (s/f'piece-{piece:02d}').exists() else '';dir=s/f'piece-{piece:02d}{suffix}';dir.mkdir(exist_ok=False);(dir/'calls').mkdir();subset=units[start:start+20]
 with ThreadPoolExecutor(max_workers=4) as pool:
  observations=[future.result() for future in as_completed([pool.submit(run,u,dir) for u in subset])]
 assert time.monotonic()-begun<1400
 assert frozen==snapshot(),'counted input changed'
 failed=[]
 for unit,code,report in observations:
  if code!=0 or report['status']!='pass':failed.append({'unit':unit['name'],'exit':code,'report':report})
  shutil.copytree(results/unit['name'],dir/unit['name'],ignore=shutil.ignore_patterns('adapted-tree'))
 counts=Counter();callers=defaultdict(Counter);noninteger=0;lines=0
 for file in sorted((dir/'calls').glob('*.jsonl')):
  for line in file.open():
   row=json.loads(line);counts[row['classification']]+=1;lines+=1;noninteger+=row['nonIntegerOffset']
   if row['classification']!=classes[0]:
    assert row['caller'],row
    callers[row['classification']][row['caller']]+=1
 record={'piece':piece,'seconds':time.monotonic()-begun,'units':[u['name'] for u in subset],'calls':lines,'counts':{c:counts[c] for c in classes},'nonIntegerOffsetsIncludingUndefinedArrays':noninteger,'callers':{c:dict(v) for c,v in callers.items()},'failed':failed,'recorderVersion':2,'unitDeadlineSeconds':90,'attemptSuffix':suffix}
 (dir/'piece.json').write_text(json.dumps(record,indent=2)+'\n')
 # First-piece archive also contains measurement provenance and excluded seed results.
 if piece==1:
  shutil.copytree(s/'raw',dir/'preparation');shutil.copytree(s/'seeds',dir/'excluded-seeds')
 for c in classes:summary['counts'][c]+=counts[c]
 summary['nonIntegerOffsetsIncludingUndefinedArrays']+=noninteger
 for c,sites in callers.items():
  for caller,n in sites.items():summary['callers'][c][caller]=summary['callers'][c].get(caller,0)+n
 if not failed:
  summary['completedUnits'] += [u['name'] for u in subset]
  for _,_,report in observations:
   for k,n in report['counts'].items():summary['oracleCounts'][k]+=n
 summary['pieces'].append(record);summary['complete']=len(summary['completedUnits'])==len(units)
 (out/'ELEMENTAT-COUNT.json').write_text(json.dumps(summary,indent=2)+'\n')
 files=sorted(p for p in dir.rglob('*') if p.is_file());(dir/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(dir))+'\n' for p in files))
 archive=out/f'elementat-count-{piece:02d}{suffix}.tar.gz'
 with tarfile.open(archive,'w:gz') as tar:tar.add(dir,arcname=f'piece-{piece:02d}')
 (out/f'elementat-count-{piece:02d}{suffix}.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
 git('add',str(out.relative_to(repo)));git('commit','-m',f'never cut alone: lands with chg_pyyr04zd, because the census reads its stop\n\n99z: counted lane piece {piece}, {len(subset)} shards, {lines} elementAt calls')
 git('push','origin','typescript/elementat-any')
 print(json.dumps({'piece':piece,'seconds':record['seconds'],'completed':len(summary['completedUnits']),'counts':summary['counts'],'failedUnits':len(failed)}),flush=True)
 for unit,_,_ in observations:shutil.rmtree(results/unit['name']/'adapted-tree',ignore_errors=True)
 if failed:raise SystemExit('Counted lane unit failed; checkpoint pushed with incomplete count')
assert summary['oracleCounts']==plan['counts'],summary['oracleCounts']
print('COMPLETE',json.dumps(summary['counts']),flush=True)

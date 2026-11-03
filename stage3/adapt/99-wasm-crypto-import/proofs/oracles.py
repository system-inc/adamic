#!/usr/bin/env python3
"""Main first, verify its unchanged lane predicate, then with-99-wasm; no concurrent work."""
from pathlib import Path
import os,sys,json,subprocess as s,time,hashlib,gzip,shutil,signal
sys.dont_write_bytecode=True
repo=Path(__file__).resolve().parents[4];proof=Path(__file__).resolve().parent;scratch=Path(os.environ.get('CRYPTO_PROOF_SCRATCH','/workspace/scratch/wasm-crypto-proof'))
assert json.loads((proof/'preparation.json').read_text())['status']=='pass'
os.environ.update(NODE_PATH=str(repo/'stage3/api/node_modules'),NODE_OPTIONS='--max-old-space-size=3072',npm_config_cache=str(scratch/'npm-cache'),TMPDIR=str(scratch/'tmp'))
(scratch/'tmp').mkdir(exist_ok=True)
sys.path.insert(0,str(repo/'stage3/lane'));from check import check_results,api_check
expected=json.loads((repo/'stage3/lane/expected.json').read_text());sanctioned=json.loads((repo/'stage3/lane/sanctioned-api.json').read_text())
resume=os.environ.get('CRYPTO_PROOF_RESUME_AFTER')=='1'
records=[r for r in json.loads((proof/'oracle-progress.json').read_text()) if r['label']=='control'] if resume else []
if resume:
 assert len(records)==1 and records[0]['lane']['status']=='pass' and records[0]['memoryBefore']==records[0]['memoryAfter']
for label in (['after'] if resume else ['control','after']):
 tree=scratch/label;out=scratch/('oracle-'+label)
 # Release only clean file-cache pages of finished task artifacts, never file bytes.
 seen=set();evicted=0
 for root in [Path('/workspace/scratch'),Path('/home/agent/.cache')]:
  for base,dirs,files in os.walk(root,followlinks=False):
   for name in files:
    file=Path(base)/name
    try:
     st=file.stat();identity=(st.st_dev,st.st_ino)
     if not file.is_file() or st.st_size<1 or identity in seen:continue
     seen.add(identity);fd=os.open(file,os.O_RDONLY)
     try:os.posix_fadvise(fd,0,0,os.POSIX_FADV_DONTNEED)
     finally:os.close(fd)
     evicted+=1
    except (OSError,ValueError):pass
 (proof/'logs'/('cache-pages-'+label+'.json')).write_text(json.dumps(dict(files=evicted,memoryCurrent=Path('/sys/fs/cgroup/memory.current').read_text())))
 # The stock host reads this generated cache at host.ts:204 and writes it at
 # :600. Each comparison gets the same cold, size-based packing as a fresh tree.
 perf=tree/'.parallelperf.json'
 if perf.exists():
  tracked=s.run(['git','-C',str(tree),'ls-files','--error-unmatch','.parallelperf.json'],stdout=s.DEVNULL,stderr=s.DEVNULL)
  assert tracked.returncode!=0,'refuse to remove a tracked input'
  with gzip.open(proof/'logs'/('generated-parallelperf-'+label+'.json.gz'),'wb') as saved:saved.write(perf.read_bytes())
  perf.unlink()
 assert shutil.disk_usage(scratch).free>3*1024**3,'less than 3 GiB free before oracle'
 memory=Path('/sys/fs/cgroup/memory.events').read_text();command=['bash','stage3/oracle/run.sh',str(tree),str(out),'--workers','8','--limit-seconds','1800']
 with (proof/'logs'/('free-'+label+'.log')).open('w') as f:
  s.run(['free','-m'],stdout=f,check=True);f.write('\n'+memory+'\nloadavg '+Path('/proc/loadavg').read_text())
 start=time.monotonic()
 with (proof/'logs'/('oracle-'+label+'.log')).open('w') as f:
  child=s.Popen(command,cwd=repo,stdout=f,stderr=s.STDOUT,start_new_session=True)
  minimum=shutil.disk_usage(scratch).free
  while child.poll() is None:
   minimum=min(minimum,shutil.disk_usage(scratch).free)
   if time.monotonic()-start>2400 or minimum<3*1024**3:
    os.killpg(child.pid,signal.SIGTERM);child.wait(timeout=10);raise RuntimeError('oracle time/disk bound')
   time.sleep(1)
  r=s.CompletedProcess(command,child.returncode)
 report=json.loads((out/'report.json').read_text());after=Path('/sys/fs/cgroup/memory.events').read_text();lane=scratch/('lane-'+label);lane.mkdir();(lane/'adapted-tree').symlink_to(tree,target_is_directory=True);(lane/'oracle').symlink_to(out,target_is_directory=True)
 execution={'platform':{'os':'linux','arch':'x64','node':s.check_output(['node','--version'],text=True).strip()},'apply_exit':0,'oracle_exit':r.returncode,'base':'a0ba9dd52b23326880b16a7898d446597cfb5a88','adaptation':None if label=='control' else '99-wasm-crypto-import'}
 (lane/'execution.json').write_text(json.dumps(execution,indent=2)+'\n');verdict=check_results(lane)
 row={'label':label,'command':command,'exit':r.returncode,'seconds':round(time.monotonic()-start,3),'counts':report.get('counts'),'minFreeBytes':minimum,'memoryBefore':memory,'memoryAfter':after,'lane':verdict}
 records.append(row);(proof/'oracle-progress.json').write_text(json.dumps(records,indent=2)+'\n');(proof/('oracle-'+label+'.json')).write_text(json.dumps(report,indent=2)+'\n');print(json.dumps({'label':label,'counts':report.get('counts'),'lane':verdict['status'],'seconds':row['seconds']}),flush=True)
 assert memory==after,('cgroup pressure during the oracle',row)
 assert verdict['status']=='pass',verdict.get('errors')
 with (proof/'logs'/('lane-'+label+'.log')).open('w') as output:s.run(['python3',str(repo/'stage3/lane/check.py'),str(lane)],cwd=repo,stdout=output,stderr=s.STDOUT,check=True,timeout=180)
 assert report['cause_counts']=={'timeout':0,'baseline-content':1,'baseline-missing':0,'exception':0,'other':0}
old=(scratch/'oracle-control/baseline.diff').read_bytes();new=(scratch/'oracle-after/baseline.diff').read_bytes();assert old==new,'baseline diff changed'
a=(scratch/'control/built/local/typescript.d.ts').read_bytes();b=(scratch/'after/built/local/typescript.d.ts').read_bytes();assert a==b,'public API changed'
# A planted extra baseline line must fail the very lane byte check used above.
diff=scratch/'oracle-after/baseline.diff'
try:
 diff.write_bytes(new+b'\nextra baseline mutant\n');errors,_=api_check(scratch/'lane-after',sanctioned);assert any('baseline.diff does not exactly describe' in e for e in errors)
 with (proof/'logs/oracle-baseline-mutant.log').open('w') as output:
  mutant=s.run(['python3',str(repo/'stage3/lane/check.py'),str(scratch/'lane-after')],cwd=repo,stdout=output,stderr=s.STDOUT,timeout=180)
 assert mutant.returncode==1,'extra baseline line survived the real lane CLI'
 (proof/'baseline-mutant.json').write_text(json.dumps({'caught':True,'errors':errors,'laneCliExit':mutant.returncode},indent=2)+'\n')
finally:diff.write_bytes(new)
with (proof/'logs/lane-after-restored.log').open('w') as output:
 s.run(['python3',str(repo/'stage3/lane/check.py'),str(scratch/'lane-after')],cwd=repo,stdout=output,stderr=s.STDOUT,check=True,timeout=180)
# Verify the actual built-and-tested source retains the imports-and-declarations scope.
check=proof/'check.cjs'
expression=f"const p=require({str(check)!r});console.log(JSON.stringify({{scope:p.TestCryptoSourceScope({str(scratch/'control')!r},{str(scratch/'after')!r}),transpile:p.TestCryptoTranspileScope({str(scratch/'control')!r},{str(scratch/'after')!r})}}))"
with (proof/'post-suite-scope.json').open('w') as output:s.run(['node','-e',expression],stdout=output,check=True,timeout=60)
with (proof/'logs/post-suite-idempotence.log').open('w') as output:s.run(['node',str(proof.parent/'adapt.cjs'),str(scratch/'after')],stdout=output,stderr=s.STDOUT,check=True,timeout=180)
assert json.loads((proof/'logs/post-suite-idempotence.log').read_text())['files']==0
(proof/'baseline.diff').write_bytes(new)
result={'status':'pass','runs':records,'baselineDiff':{'byteIdentical':True,'bytes':len(new),'sha256':hashlib.sha256(new).hexdigest()},'api':{'byteIdentical':True,'bytes':len(a),'sha256':hashlib.sha256(a).hexdigest()},'laneCheckerUnchanged':True,'baselineByteMutantCaught':True}
(proof/'oracle-comparison.json').write_text(json.dumps(result,indent=2)+'\n');print('PASS serial oracle comparison',flush=True)

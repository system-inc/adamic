"""Bounded scan-only batches; ten paired files per non-forced checkpoint."""
import concurrent.futures,gzip,hashlib,json,os,signal,subprocess,sys,time
from pathlib import Path
BASE=Path('/workspace/scratch/c4v7-census')
OUT=Path(__file__).resolve().parent
first,last=map(int,sys.argv[1:]); assert 1<=first<=last<=79 and last-first<10
roots=sorted((BASE/'adapted-head/src/compiler').rglob('*.ts'));assert len(roots)==79

def run(task):
 label,index=task
 tree=BASE/f'adapted-{label}/src/compiler';rel=roots[index-1].relative_to(BASE/'adapted-head/src/compiler');root=tree/rel
 dest=OUT/'part2'/label;dest.mkdir(parents=True,exist_ok=True)
 name=f'{index:02}';raw=dest/(name+'.jsonl');log=dest/(name+'.log')
 env=dict(os.environ,LATENT_ONLY_FILE=str(root),LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle')
 start=time.monotonic();timed=False
 with log.open('wb') as f:
  child=subprocess.Popen([str(BASE/f'scan-{label}'),str(tree),str(raw)],env=env,cwd=BASE/label,stdout=f,stderr=f,start_new_session=True)
  try:
   try:rc=child.wait(timeout=1500)
   except subprocess.TimeoutExpired:
    timed=True;os.killpg(child.pid,signal.SIGTERM)
    try:rc=child.wait(timeout=10)
    except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);rc=child.wait()
  finally:
   try:os.killpg(child.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   child.wait()
 status=dict(file=str(rel),commit_label=label,exit_code=rc,seconds=time.monotonic()-start,timeout=timed,process_group_cleaned=True,command=[str(BASE/f'scan-{label}'),str(tree),str(raw)])
 if rc==0:
  rows=[json.loads(x) for x in raw.read_text().splitlines()]
  assert rows[0]['module_format']=='CommonJSBundle'
  records=[x for x in rows if x.get('file')]
  assert len(records)==1 and records[0]['file']==str(root)
  assert records[0]['units']==[]
  assert rows[-1]['status']=='relation_report'
  assert not rows[-1]['findings'],rows[-1]['findings']
  assert not rows[-1]['relation_overflow']['after_load'] and not rows[-1]['relation_overflow']['end']
  status['relation_report_clean']=True
  status.update(findings=len(records[0]['findings']),zero_lowering_units=True,validated=True)
 for f in (raw,log):
  if f.exists():
   data=f.read_bytes();status[f.suffix+'_sha256']=hashlib.sha256(data).hexdigest();f.with_suffix(f.suffix+'.gz').write_bytes(gzip.compress(data,mtime=0));f.unlink()
 (dest/(name+'-status.json')).write_text(json.dumps(status,indent=2)+'\n')
 print(label,index,str(rel),rc,status.get('findings'),flush=True)
 return rc
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
 codes=list(pool.map(run,[(label,i) for i in range(first,last+1) for label in ('main','head')]))
assert all(x==0 for x in codes),codes
subprocess.run(['git','add',str(OUT)],check=True)
subprocess.run(['git','commit','--quiet','-m',f'census: refusal scan paired batch {first}-{last}'],check=True)
subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/c4v7-census'],check=True)

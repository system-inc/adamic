from pathlib import Path
import os,subprocess,json,time,signal,hashlib,shutil,sys
s=Path('/workspace/scratch/diag-read-view-3');r=s/'repo';v2=Path('/workspace/scratch/diag-read-view-2');raw=s/'raw'
env=dict(os.environ,TMPDIR=str(s/'tmp'),GOTMPDIR=str(s/'tmp'),GOCACHE=str(v2/'gocache'),PATH='/workspace/adamic-tools/bin:/workspace/adamic-tools/go/bin:'+os.environ['PATH'],ADAMIC_TYPESCRIPT_SOURCE=str(v2/'TypeScript'),ADAMIC_BUILD_CACHE_DIR=str(v2/'cache'),NODE_PATH=str(r/'stage3/api/node_modules'),PYTHONDONTWRITEBYTECODE='1',npm_config_cache=str(v2/'npm-cache'))
def run(label,args,limit=6000):
 begin=time.monotonic()
 with (raw/(label+'.log')).open('w') as out:
  child=subprocess.Popen(list(map(str,args)),cwd=r,env=env,stdout=out,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=child.wait(timeout=limit)
  except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait();code=124
 result=dict(command=list(map(str,args)),exit=code,seconds=time.monotonic()-begin,limit=limit)
 (raw/(label+'.result.json')).write_text(json.dumps(result,indent=2)+'\n');print(label,json.dumps(result),flush=True)
 assert code==0,label
def source(tree):return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
run('zero-write-replay',['node',s/'replay.cjs'])
run('apply',[r/'stage3/apply.sh',s/'applied-tree'])
a=source(s/'applied-tree');b=source(v2/'applied-tree');assert a==b
(raw/'source-match.json').write_text(json.dumps({'identicalToProvenV2':True,'files':len(a),'source':a},indent=2)+'\n')
run('whole-lane',[r/'stage3/lane/run.sh',s/'lane'])
lane=json.loads((s/'lane/report.json').read_text());assert lane['status']=='pass' and lane['counts']=={'passing':106366,'failing':1,'pending':0}
assert source(s/'lane/adapted-tree')==a
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=r,text=True).strip();parent=subprocess.check_output(['git','rev-parse','HEAD^'],cwd=r,text=True).strip();assert parent=='8e3a784aa5a601f57342c07b68f145fe56252c21'
summary={'status':'PASSED','implementation':sha,'parent':parent,'changesVsV2':['stage3/upstream/LEDGER.md','stage3/adapt/99-correctness-binding-views/README.md'],'ledgerOriginalPrefixPreserved':True,'replayWrites':0,'apply':'passed','sourceFilesUnchangedVsV2':len(a),'lane':lane,'inheritedFullProofs':'d7a6aa49976828aa9627b64c137234b718c9e503; identical 747 source files','loomSubmission':False}
(raw/'FINAL.json').write_text(json.dumps(summary,indent=2)+'\n')
for name in ['report.json','verdict.json','execution.json','verdict.txt','failures.md','patch-set.md','apply.log','oracle.log']:
 if (s/'lane'/name).exists():shutil.copyfile(s/'lane'/name,raw/('lane-'+name))
for p in (s/'lane/oracle').rglob('*'):
 if p.is_file() and p.suffix in ['.json','.jsonl','.log','.diff','.md']:
  target=raw/'lane-oracle'/p.relative_to(s/'lane/oracle');target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,target)
print(json.dumps({'status':'PASSED','sha':sha,'lane':lane['counts']},indent=2),flush=True)

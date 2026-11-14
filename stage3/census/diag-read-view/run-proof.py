import os,sys,json,subprocess,time,signal,concurrent.futures
from pathlib import Path
s=Path('/workspace/scratch/diag-read-view-1');r=s/'repo';raw=s/'raw';raw.mkdir(exist_ok=True)
env=dict(os.environ,TMPDIR=str(s/'tmp'),GOTMPDIR=str(s/'tmp'),GOCACHE=str(s/'gocache'),PATH='/workspace/adamic-tools/bin:/workspace/adamic-tools/go/bin:'+os.environ['PATH'],ADAMIC_TYPESCRIPT_SOURCE=str(s/'TypeScript'),ADAMIC_BUILD_CACHE_DIR=str(s/'cache'),NODE_PATH=str(r/'stage3/api/node_modules'))
def run(label,args,cwd=r,extra={},limit=6000):
 begin=time.monotonic();e=dict(env,**extra)
 with (raw/(label+'.log')).open('w') as f:
  p=subprocess.Popen(list(map(str,args)),cwd=cwd,env=e,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
  try: code=p.wait(timeout=limit)
  except subprocess.TimeoutExpired:
   os.killpg(p.pid,signal.SIGKILL);p.wait();code=124
 result=dict(label=label,args=list(map(str,args)),cwd=str(cwd),extra=extra,exit=code,seconds=time.monotonic()-begin,limit=limit)
 (raw/(label+'.result.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True);return code
mode=sys.argv[1]
if mode=='census':
 side=sys.argv[2];tree=s/(side+'-tree');files=sys.argv[3:] or ['program.ts','utilitiesPublic.ts','utilities.ts','types.ts']
 def census(file):return run(side+'-'+file+'-census',[s/'latent',tree/'src/compiler',raw/(side+'-'+file+'.jsonl')],extra=dict(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(tree/'src/compiler'/file),CENSUS_MODULE_FORMAT='CommonJSBundle'))
 with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:codes=list(pool.map(census,files))
 sys.exit(max(codes))
elif mode=='dts':
 side=sys.argv[2];tree=s/(side+'-tree');sys.exit(run(side+'-dts',[tree/'node_modules/.bin/hereby','dts'],cwd=tree))
elif mode=='apply':sys.exit(run('candidate-apply',[r/'stage3/apply.sh',s/'applied-tree']))

elif mode=='lane':sys.exit(run('whole-lane',[r/'stage3/lane/run.sh',s/'lane']))

elif mode=='build':
 side=sys.argv[2];sys.exit(run(side+'-build-compiler',['npm','run','build:compiler'],cwd=s/(side+'-tree')))

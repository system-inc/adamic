import concurrent.futures,json,os,pathlib,subprocess,time,collections
root=pathlib.Path('/workspace/adamic');out=pathlib.Path('/tmp/void-split-evidence')
paths=subprocess.check_output(['git','ls-files','-z'],cwd=root).decode().split('\0')
paths=[p for p in paths if p.endswith(('.a','.ts')) and not p.endswith('.d.ts') and not p.startswith('cohere/') and 'node_modules' not in pathlib.PurePosixPath(p).parts]
original={json.loads(line)['path'] for line in (out/'admission-ad029.jsonl').read_text().splitlines()}
assert len(original)==3699
current=set(paths); missing=original-current; paths=sorted(current|original)
historical=pathlib.Path('/tmp/adamic-void-main')
assert all((historical/p).is_file() for p in missing)
env=dict(os.environ,GOMAXPROCS='1');started=time.monotonic()
prefix=os.environ.get('VOID_SCAN_PREFIX','admission')
head_binary=os.environ.get('VOID_HEAD_BIN','/tmp/adamic-void-split-commit')
def inspect(path):
 observations={}
 directory=root if path in current else historical
 for label,binary in [('base','/tmp/adamic-void-split-base-bin'),('head',head_binary)]:
  before=time.monotonic()
  try:
   result=subprocess.run([binary,'c',path],cwd=directory,env=env,stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,text=True,timeout=180)
   observations[label]={'exit':result.returncode,'stderr':result.stderr,'seconds':round(time.monotonic()-before,3)}
  except subprocess.TimeoutExpired as error:
   observations[label]={'exit':None,'timeout':True,'stderr':(error.stderr or b'').decode() if isinstance(error.stderr,bytes) else (error.stderr or '')}
 return dict(path=path,original=path in original,current=path in current,source_root=str(directory),**observations)
new=[];timeouts=[];results=[]
with (out/(prefix+'.jsonl')).open('w') as log, concurrent.futures.ThreadPoolExecutor(max_workers=8) as pool:
 for result in pool.map(inspect,paths):
  log.write(json.dumps(result)+'\n');log.flush();results.append(result)
  if any(result[k].get('timeout') for k in ['base','head']):timeouts.append(result['path'])
  if 'adamic/no-void-widening' in result['head']['stderr'] and 'adamic/no-void-widening' not in result['base']['stderr']:new.append(result)
  if len(results)%100==0:print(f'{len(results)}/{len(paths)} in {time.monotonic()-started:.1f}s',flush=True)
summary={'files':len(paths),'current_files':len(current),'original_files':len(original),'historical_files':sorted(missing),'additional_files':sorted(current-original),'completed':len(results),'seconds':time.monotonic()-started,'timeouts':timeouts,'exit_pairs':dict(collections.Counter(str((r['base']['exit'],r['head']['exit'])) for r in results)),'newly_refused':new,'panics':[r for r in results if r['base']['exit']==2 or r['head']['exit']==2]}
(out/(prefix+'-summary.json')).write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k not in ['panics','newly_refused']}),flush=True)
for result in new:print(result['path'],result['head']['stderr'].strip(),flush=True)

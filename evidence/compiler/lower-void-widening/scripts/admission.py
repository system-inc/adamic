import concurrent.futures, json, os, pathlib, subprocess, time
root=pathlib.Path('/workspace/adamic')
paths=subprocess.check_output(['git','ls-files','-z'],cwd=root).decode().split('\0')
paths=[p for p in paths if p.endswith(('.a','.ts')) and not p.endswith('.d.ts') and not p.startswith('cohere/') and 'node_modules' not in pathlib.PurePosixPath(p).parts]
env=dict(os.environ,GOMAXPROCS='1')
prefix=os.environ.get('VOID_ADMISSION_PREFIX','admission')
started=time.monotonic()
def inspect(path):
 observations={}
 for label,binary in [('main','/tmp/adamic-main'),('head','/tmp/adamic-void')]:
  before=time.monotonic()
  try:
   result=subprocess.run([binary,'c',path],cwd=root,env=env,stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,text=True,timeout=180)
   observations[label]={'exit':result.returncode,'stderr':result.stderr,'seconds':round(time.monotonic()-before,3)}
  except subprocess.TimeoutExpired as error:
   observations[label]={'exit':None,'timeout':True,'stderr':(error.stderr or b'').decode() if isinstance(error.stderr,bytes) else (error.stderr or '')}
 return dict(path=path,**observations)
new=[]; timeouts=[]; count=0
with open(f'/tmp/void-evidence/{prefix}.jsonl','w') as out, concurrent.futures.ThreadPoolExecutor(max_workers=8) as pool:
 for result in pool.map(inspect,paths):
  out.write(json.dumps(result)+'\n'); out.flush(); count+=1
  if any(result[k].get('timeout') for k in ['main','head']): timeouts.append(result['path'])
  if 'adamic/no-void-widening' in result['head']['stderr'] and 'adamic/no-void-widening' not in result['main']['stderr']:
   new.append(result)
  if count%100==0: print(f'{count}/{len(paths)} in {time.monotonic()-started:.1f}s',flush=True)
summary={'files':len(paths),'completed':count,'seconds':time.monotonic()-started,'timeouts':timeouts,'newly_refused':new}
pathlib.Path(f'/tmp/void-evidence/{prefix}-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({'completed':count,'timeouts':len(timeouts),'newly_refused':len(new),'seconds':summary['seconds']}),flush=True)
for result in new: print(result['path'],result['head']['stderr'].strip(),flush=True)

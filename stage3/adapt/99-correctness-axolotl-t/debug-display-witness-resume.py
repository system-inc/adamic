import concurrent.futures,hashlib,json,os,pathlib,re,subprocess,sys,threading
root=pathlib.Path('/workspace/scratch/debug-display-before');old=pathlib.Path('/workspace/scratch/debug-display-witness');work=pathlib.Path('/workspace/scratch/debug-display-witness-resume');work.mkdir(exist_ok=True)
prior={r['index']:r for line in pathlib.Path('/workspace/scratch/debug-display-witness.log').read_text().splitlines() if line.startswith('{') for r in [json.loads(line)]}
cases=[]
for p in sorted((root/'tests/cases/compiler').glob('*.ts')):
 if re.search(r'@(?:filename|link|symlink|libFiles|includeBuiltFile)\s*:',p.read_text(errors='replace'),re.I):continue
 cases.append(p)
 if len(cases)==200:break
assert len(cases)==200
base=dict(os.environ);base.pop('NODE_OPTIONS',None)
for key in ['VSCODE_INSPECTOR_OPTIONS','NODE_INSPECTOR_IPC']:base.pop(key,None)
stop=threading.Event()
def meta(child):return {'exit':child.returncode,**{stream+'_sha256':hashlib.sha256(getattr(child,stream)).hexdigest() for stream in ['stdout','stderr']},**{stream+'_bytes':len(getattr(child,stream)) for stream in ['stdout','stderr']}}
def run(item):
 if stop.is_set():return None
 index,p=item;command=['node',str(root/'built/local/tsc.js'),'--ignoreConfig','--pretty','false','--noEmit','--skipLibCheck',str(p)];pair=[]
 if index in prior:
  result=prior[index];assert result['case']==str(p.relative_to(root)) and result['command']==command and result['identical'];pair=[result['plain'],result['debug']];pair_root=old
  for label,m in zip(['plain','debug'],pair):
   for stream in ['stdout','stderr']:
    data=(old/f'{index:03d}-{label}.{stream}').read_bytes();assert hashlib.sha256(data).hexdigest()==m[stream+'_sha256'] and len(data)==m[stream+'_bytes']
 else:
  pair_root=work
  for debug in [False,True]:
   env=dict(base)
   if debug:env['VSCODE_INSPECTOR_OPTIONS']='{}'
   child=subprocess.run(command,cwd=root,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE);pair.append(meta(child))
   label='debug' if debug else 'plain'
   for stream in ['stdout','stderr']:(work/f'{index:03d}-{label}.{stream}').write_bytes(getattr(child,stream))
 identical=pair[0]['exit']==pair[1]['exit'] and all((pair_root/f'{index:03d}-plain.{s}').read_bytes()==(pair_root/f'{index:03d}-debug.{s}').read_bytes() for s in ['stdout','stderr'])
 if not identical:
  stop.set();(work/'first-difference.json').write_text(json.dumps({'index':index,'pair':pair,'case':str(p)},indent=2));raise RuntimeError('paired witness differs')
 counted=list(command);counted[1]='/workspace/scratch/debug-display-counted/built/local/tsc.js';env=dict(base,VSCODE_INSPECTOR_OPTIONS='{}',NODE_OPTIONS='--require=/workspace/scratch/debug-display-preload.cjs',DEBUG_DISPLAY_COUNTS=str(work/f'{index:03d}-counts.jsonl'))
 child=subprocess.run(counted,cwd=root,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE);observed=meta(child)
 for stream in ['stdout','stderr']:(work/f'{index:03d}-observed.{stream}').write_bytes(getattr(child,stream))
 assert observed==pair[1],('observer differs',index,observed,pair[1])
 records=[json.loads(l) for l in (work/f'{index:03d}-counts.jsonl').read_text().splitlines()];counts={n:sum(r['counts'][n] for r in records) for n in ['attachFlowNodeDebugInfo','attachNodeArrayDebugInfo','enableDebugInfo']}
 result={'index':index,'case_sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'plain':pair[0],'debug':pair[1],'identical':True,'pair_reused_from_interrupted_controller':index in prior,'observed':observed,'counts':counts}
 (work/f'{index:03d}-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({'index':index,'identical':True,'counts':counts}),flush=True);return result
results=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
 futures=[pool.submit(run,item) for item in enumerate(cases)]
 for future in concurrent.futures.as_completed(futures):
  try:
   result=future.result()
   if result is not None:results.append(result)
  except BaseException:
   stop.set()
   for pending in futures:pending.cancel()
   raise
results.sort(key=lambda r:r['index']);assert len(results)==200
counts={n:sum(r['counts'][n] for r in results) for n in ['attachFlowNodeDebugInfo','attachNodeArrayDebugInfo','enableDebugInfo']};assert all(v>0 for v in counts.values())
(work/'summary.json').write_text(json.dumps({'cases':200,'all_identical':True,'observed_replays_identical':True,'debug_counts':counts,'reused_pairs':len(prior),'selection':'First 200 sorted single-file compiler .ts cases, excluding harness filesystem directives. Fixed flags: ignoreConfig, pretty false, noEmit, skipLibCheck.','rows':results},indent=2)+'\n')

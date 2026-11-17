"""Run pinned probes without changing either source checkout; preserve every child exit."""
import concurrent.futures,datetime,gzip,hashlib,json,os,pathlib,subprocess,sys,threading,time
scratch=pathlib.Path(sys.argv[1]).resolve();tree=scratch/'adapted';results=scratch/'results';results.mkdir(exist_ok=True)
batch=sys.argv[2] if len(sys.argv)>2 else 'all'
assert batch in ('entry','full','all')
manifestPath=results/(batch+'-manifest.json')
lock=threading.Lock();manifest=[]
def now():return datetime.datetime.now(datetime.timezone.utc).isoformat()
def run_task(task):
 label,mode,relative,root=task
 directory=results/label/mode;directory.mkdir(parents=True,exist_ok=True)
 stem=relative.replace('/','__').removesuffix('.ts') if mode!='entry' else 'tsc'
 output=directory/(stem+('.json' if mode=='scan' else '.jsonl'));log=directory/(stem+'.log');env=dict(os.environ)
 env.update(CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TAIL='1')
 for k in ['LATENT_SPECULATIVE','LATENT_REPLAY','LATENT_ONLY_FILE','LATENT_FULL'] :env.pop(k,None)
 if mode=='scan':args=[str(scratch/(label+'-tools')/'scan-all'),str(tree/root)]
 else:
  env['LATENT_FULL']='0' if mode=='entry' else '1'
  if mode=='full':env['LATENT_ONLY_FILE']=str(tree/relative)
  env['LATENT_TIMES']=str(directory/(stem+'.times.jsonl'))
  args=[str(scratch/(label+'-tools')/'entry-census'),str(tree/root),str(output)]
 row=dict(label=label,mode=mode,file=relative,root=root,args=args,start=now(),state='running',output=str(output.relative_to(scratch)),log=str(log.relative_to(scratch)),env={k:env[k] for k in ['CENSUS_MODULE_FORMAT','LATENT_FULL','LATENT_ONLY_FILE','LATENT_ASSERT_NO_OUTPUT','LATENT_TAIL'] if k in env})
 with lock:manifest.append(row);manifestPath.write_text(json.dumps(manifest,indent=2)+'\n')
 start=time.monotonic()
 with log.open('wb') as stderr:
  with (output if mode=='scan' else directory/(stem+'.stdout')).open('wb') as stdout:
   child=subprocess.Popen(args,env=env,cwd=scratch/label,stdout=stdout,stderr=stderr)
   row['pid']=child.pid
   _,status,usage=os.wait4(child.pid,0);child.returncode=os.waitstatus_to_exitcode(status)
 row.update(exit=child.returncode,end=now(),seconds=time.monotonic()-start,maxRssKiB=usage.ru_maxrss,state='complete' if child.returncode==0 else 'failed')
 try:
  if mode=='scan':
   value=json.loads(output.read_text());row['scanCompleted']=value.get('Completed');row['boundary']=value.get('Boundary');row['findings']=len(value.get('Stops') or [])
  else:
   records=[json.loads(line) for line in output.read_text().splitlines() if line.strip()]
   row['streamComplete']=bool(records and records[-1].get('status')=='relation_report');row['header']=records[0] if records and records[0].get('status')=='blocked' else {k:records[0].get(k) for k in ['status','latent_mode','module_format','checker_rejected']} if records else None
   row['files']=[r['file'] for r in records if 'file' in r];row['findings']=sum(len(r.get('findings') or []) for r in records)
   if mode=='full':row['selectionExact']=row['files']==[str(tree/relative)]
 except Exception as error:row['readError']=str(error)
 if output.exists():row['sha256']=hashlib.sha256(output.read_bytes()).hexdigest()
 with lock:manifestPath.write_text(json.dumps(manifest,indent=2)+'\n')
 print(json.dumps({k:row.get(k) for k in ['label','mode','file','exit','seconds','maxRssKiB','streamComplete','findings','boundary']}),flush=True)
 return row
while not ((tree/'patch-set.md').exists() and all((scratch/(label+'-tools')/'provenance.json').exists() for label in ['base','head'])):time.sleep(2)
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 entry=[pool.submit(run_task,(label,'entry','src/tsc/tsc.ts','src/tsc/tsc.ts')) for label in ['base','head']] if batch in ('entry','all') else []
 for f in entry:f.result()
 while batch in ('full','all') and not (scratch/'callers.json').exists():time.sleep(2)
 selection=json.loads((scratch/'callers.json').read_text())['fullCensusFiles'] if batch in ('full','all') else []
 # Prioritize the owner, then the complete exact caller-file list.
 selection.sort(key=lambda r:(r['file']!='src/compiler/core.ts',r['file']))
 tasks=[(label,'full',r['file'],r['root']) for r in selection for label in ['base','head']]
 tasks += [(label,'scan',relative,relative) for relative in ['src/compiler/core.ts','src/tsc/tsc.ts'] for label in ['base','head']] if batch in ('full','all') else []
 for _ in pool.map(run_task,tasks):pass
failed=any(r['exit'] or r.get('readError') or r['mode']!='scan' and (not r.get('streamComplete') or r['mode']=='full' and not r.get('selectionExact')) for r in manifest)
print(json.dumps({'finished':now(),'runs':len(manifest),'failed':failed}),flush=True)
sys.exit(1 if failed else 0)

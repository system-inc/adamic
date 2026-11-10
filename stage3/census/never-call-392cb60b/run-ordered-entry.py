import pathlib,subprocess,os,sys,json,time,shutil
repo=pathlib.Path('/workspace/never-call-report');out=repo/'stage3/census/never-call-392cb60b';live=pathlib.Path('/workspace/never-call-live');live.mkdir(exist_ok=True)
pending={};completed={}
for side in ['parent','head']:
 binary=pathlib.Path('/workspace/never-call-tools')/side/'census-entry-ordered'
 if not binary.is_file():raise SystemExit('build both ordered meters before launching: '+str(binary))
for side in ['parent','head']:
 name='tsc-entry-ordered-first-'+side;env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
 env.update(LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(live/(name+'.units.jsonl')),LATENT_PREPASS_TIMES=str(live/(name+'.prepasses.log')))
 cmd=['python3','/tmp/never-call-run.py',name,str(repo if side=='parent' else pathlib.Path('/workspace/never-call-head')),str(live),'/workspace/never-call-tools/'+side+'/census-entry-ordered','/workspace/never-call-adapted/src/tsc/tsc.ts',str(live/(name+'.jsonl'))]
 pending[subprocess.Popen(cmd,env=env)]=(side,name)
try:
 while pending:
  for p,(side,name) in list(pending.items()):
   rc=p.poll()
   if rc is None:continue
   p.wait();print(name,'wrapper child exit code:',rc,flush=True)
   for source in live.glob(name+'.*'):shutil.copyfile(source,out/source.name)
   completed[side]=rc;del pending[p]
   (out/'ordered-entry-execution.json').write_text(json.dumps(dict(concurrent=True,reason='bounded session time after correcting observation order',completed_child_exit_codes=completed),indent=2)+'\n')
   for command in [['python3','/tmp/never-call-archive.py'],['python3','/tmp/never-call-checkpoint.py',name]]:
    r=subprocess.run(command);print('post-run child exit code:',r.returncode,flush=True)
    if r.returncode:raise RuntimeError('checkpoint failed: '+str(r.returncode))
  if pending:time.sleep(1)
finally:
 for p in pending:
  if p.poll() is None:p.terminate()
  try:p.wait(timeout=10)
  except subprocess.TimeoutExpired:p.kill();p.wait()
if any(completed.values()):sys.exit(next(v for v in completed.values() if v))

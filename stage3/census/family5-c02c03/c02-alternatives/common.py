import os,sys,json,time,subprocess,shutil,hashlib,re,signal,traceback,datetime,collections,fcntl,gzip
from pathlib import Path
R=Path('/tmp/family5-land'); W=Path('/tmp/family5-c02-alternatives'); MAIN='fbe5b72ba115a8f3060d4b5229fd578574f4505f'; C02='12b9e04184663f899acb76505043ad292966f236'
W.mkdir(parents=True,exist_ok=True);os.chdir(R)
os.environ.update(ADAMIC_TYPESCRIPT_SOURCE='/workspace/mapset-upstream',ADAMIC_BUILD_CACHE_DIR='/workspace/family5-build-cache',NODE_PATH=str(R/'stage3/api/node_modules'))
lock=(W/'workflow.lock').open('w');fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
statefile=W/'state.json';state=json.loads(statefile.read_text()) if statefile.exists() else {'main':MAIN,'c02':C02,'entry_pushes':{},'candidates':{}}
class Stop(Exception):pass
def save():statefile.write_text(json.dumps(state,indent=2)+'\n')

def git(*args):return subprocess.check_output(['git',*args],cwd=R,text=True).strip()

def digest(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()

def tail(p):
 try:
  with Path(p).open('rb') as f:f.seek(max(0,f.seek(0,2)-1800));lines=f.read().decode(errors='replace').splitlines()
  for l in reversed(lines):
   try:
    x=json.loads(l)
    if 'where' in x and 'wall_seconds' in x:return 'last unit '+x['where']+' ('+str(round(x['wall_seconds'],1))+'s)'
   except ValueError:pass
 except OSError:pass
 return ''

def run(name,cmd,env=None,cwd=R,allow=False,inputs=None,force=False):
 out=W/name;status=out/'status.json'
 if status.exists() and not force:
  old=json.loads(status.read_text())
  if inputs is not None and old.get('inputs')!=inputs:raise Stop(name+': input fingerprint changed')
  if old['exit_code']!=0 and not allow:raise Stop(name+': recorded child exit '+str(old['exit_code']))
  return old['exit_code']
 if status.exists():
  dest=out.parent/(out.name+'-attempt-'+str(time.time_ns()));shutil.move(out,dest)
 out.mkdir(parents=True,exist_ok=True);cmd=list(map(str,cmd));environment=dict(os.environ) if env is None else env
 begin=time.monotonic();stamp=datetime.datetime.now(datetime.timezone.utc).isoformat();current={'proof':name,'started_utc':stamp,'command':cmd};(W/'current.json').write_text(json.dumps(current,indent=2)+'\n')
 print('START '+name,flush=True)
 with (out/'stdout.log').open('wb') as stdout,(out/'stderr.log').open('wb') as stderr:
  child=subprocess.Popen(cmd,cwd=cwd,env=environment,stdout=stdout,stderr=stderr,start_new_session=True)
  (out/'child.json').write_text(json.dumps({'pid':child.pid,'supervisor_pid':os.getpid(),'started_utc':stamp,'proc_stat_at_launch':Path('/proc/'+str(child.pid)+'/stat').read_text()},indent=2)+'\n')
  while True:
   try:code=child.wait(timeout=50);break
   except subprocess.TimeoutExpired:print(f'RUNNING {name}: {time.monotonic()-begin:.0f}s '+tail(out/'stderr.log'),flush=True)
  try:os.killpg(child.pid,signal.SIGTERM)
  except ProcessLookupError:pass
 result={'command':cmd,'cwd':str(cwd),'environment':{k:v for k,v in environment.items() if k.startswith(('ADAMIC_','LATENT_','CENSUS_','FAMILY5_'))},'inputs':inputs,'started_utc':stamp,'ended_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'exit_code':code,'wall_seconds':round(time.monotonic()-begin,3)}
 status.write_text(json.dumps(result,indent=2)+'\n');print(('PASS ' if code==0 else 'FAIL ')+name+' real_child_exit='+str(code),flush=True)
 if code and not allow:raise Stop(name+': child exit '+str(code))
 return code

def source_manifest(tree):return {str(p.relative_to(tree)):digest(p) for p in sorted((tree/'src').rglob('*')) if p.is_file()}

def read_stream(p):
 records=[json.loads(l) for l in p.read_text().splitlines()];assert records and records[-1].get('status')=='relation_report',p
 assert records[0].get('checker_rejected') is False and records[0].get('diagnostics')==[],p
 assert records[0]['module_format']=='CommonJSBundle' and records[0]['latent_mode']=='full',p
 return [f for row in records for f in row.get('findings',[])]

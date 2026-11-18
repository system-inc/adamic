import os,sys,json,time,subprocess,shutil,hashlib,re,signal,traceback,datetime,collections,fcntl
from pathlib import Path
R=Path('/tmp/family5-land'); W=Path('/tmp/family5-final-proofs'); C02='12b9e04184663f899acb76505043ad292966f236'; MAIN='fbe5b72ba115a8f3060d4b5229fd578574f4505f'
os.chdir(R)
os.environ.update(ADAMIC_TYPESCRIPT_SOURCE='/workspace/mapset-upstream',ADAMIC_BUILD_CACHE_DIR='/workspace/family5-build-cache',NODE_PATH=str(R/'stage3/api/node_modules'))
lock=(W/'workflow.lock').open('w');fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
statefile=W/'state.json';state=json.loads(statefile.read_text()) if statefile.exists() else {'main':MAIN,'c02':C02,'comparisons':{}}
scout=json.loads(Path('/workspace/family5-analysis/sites.json').read_text())
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

def target_sites(folder,groups):
 reasons={s['reason'] for s in scout if s['group'] in groups};sites=set()
 for p in folder.glob('*/latent.jsonl'):
  for f in read_stream(p):
   if f['reason'] in reasons:sites.add((p.parent.name,re.sub(r'^.*?/src/','src/',f['where']),f['reason']))
 return sites

def census(label,tree,file):
 out=W/label/file.removesuffix('.ts');env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TAIL='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(tree/'src/compiler'/file),LATENT_TIMES='/dev/stderr')
 inputs={'types_sha256':digest(tree/'src/compiler/types.ts'),'file_sha256':digest(tree/'src/compiler'/file),'tool_sha256':digest('/tmp/family5-tools/entry-census')}
 run(label+'/'+file.removesuffix('.ts'),['/tmp/family5-tools/entry-census',tree/'src/tsc/tsc.ts',out/'latent.jsonl'],env,inputs=inputs)
 read_stream(out/'latent.jsonl')

def stock_and_public(name,base,tree):
 run(name+'/stock',['node','/tmp/family5-stock.cjs',tree,W/name/'stock.json'],inputs={'types_sha256':digest(tree/'src/compiler/types.ts')})
 run(name+'/stock-strict',['node','/tmp/family5-stock.cjs',tree,W/name/'stock-strict.json'],dict(os.environ,FAMILY5_STOCK_STRICT='1'),inputs={'types_sha256':digest(tree/'src/compiler/types.ts')})
 roots=json.loads((W/name/'stock.json').read_text())['roots'];assert len(roots)==79
 run(name+'/adamic-types',['/tmp/family5-tools/adamic','types',*[tree/f for f in roots]],inputs={'types_sha256':digest(tree/'src/compiler/types.ts')})
 # The local build uses the locked upstream development install, not the parser install.
 p=W/name/'public-build/status.json'
 if p.exists():
  run(name+'/public-build',['node',R/'stage3/typescript/node_modules/hereby/bin/hereby.js','local'],cwd=tree)
 else:
  link=tree/'node_modules';assert not link.exists();link.symlink_to(R/'stage3/typescript/node_modules',target_is_directory=True)
  try:run(name+'/public-build',['node',R/'stage3/typescript/node_modules/hereby/bin/hereby.js','local'],cwd=tree)
  finally:link.unlink()
 run(name+'/public-compare',['cmp',base/'built/local/typescript.d.ts',tree/'built/local/typescript.d.ts'])
 (W/name/'public.json').write_text(json.dumps({'byte_identical':True,'before_sha256':digest(base/'built/local/typescript.d.ts'),'after_sha256':digest(tree/'built/local/typescript.d.ts')},indent=2)+'\n')

def comparison(name,before,after,groups):
 version=digest(W/'compare.py')[:12];out=W/name/('comparison-'+version);path=out/'comparison.json'
 run(name+'/comparison-'+version,[sys.executable,W/'compare.py',before,after,path,*groups])
 data=json.loads(path.read_text());assert all(f['target_after_distinct_sites']==0 and f['after_target_mentions_all_kinds']==0 for f in data['files'].values())
 state['comparisons'][name]=str(path.relative_to(W));save()

def paired(name,base,tree,files,groups,reuse=None):
 before=W/(name+'-before');after=W/(name+'-after')
 for file in files:
  if reuse:
   dest=before/file.removesuffix('.ts')
   if not dest.exists():shutil.copytree(reuse/file.removesuffix('.ts'),dest)
   (before/'reused-before.json').write_text(json.dumps({'source':str(reuse),'base_source_sha256':digest(base/'src/compiler/types.ts'),'reason':'The completed sequential C02 after-checker run is exactly C03\'s unchanged base input, with the same main-built tool and flags.'},indent=2)+'\n')
  else:census(name+'-before',base,file)
  census(name+'-after',tree,file)
 # Confirm that every file still exposes at least its scout count before measuring clearance.
 observed=target_sites(before,set(groups));expected=collections.Counter(Path(x['where'].rsplit(':',2)[0]).stem for x in scout if x['group'] in groups);actual=collections.Counter(x[0] for x in observed)
 coverage={'scout_sites_per_file':dict(expected),'full_before_sites_per_file':dict(actual),'full_before_distinct_sites':len(observed)}
 (W/name/'coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
 if any(actual[k]<v for k,v in expected.items()):raise Stop(name+' full baseline did not expose all scout sites: '+json.dumps(coverage))
 comparison(name,before,after,groups)

def mutant(name,adapter,base,tree,files,groups):
 summary=W/name/'mutant-result.json'
 if summary.exists() and json.loads(summary.read_text()).get('target_sites_restored'):return
 original=(tree/'src/compiler/types.ts').read_bytes()
 try:
  run(name+'/mutant-restore',['node',adapter/'mutant.cjs',tree],force=True)
  run(name+'/mutant-source-compare',['cmp',base/'src/compiler/types.ts',tree/'src/compiler/types.ts'],force=True)
  for file in files:census(name+'-mutant',tree,file)
  a=target_sites(W/(name+'-before'),groups);b=target_sites(W/(name+'-mutant'),groups)
  result={'before_target_sites':len(a),'mutant_target_sites':len(b),'target_sites_restored':a==b and bool(a),'missing':sorted(a-b),'extra':sorted(b-a)}
  if not result['target_sites_restored']:raise Stop(name+' mutant did not restore every baseline target site: '+json.dumps(result))
 finally:
  run(name+'/mutant-reapply',['node',adapter/'adapt.cjs',tree],force=True)
  assert (tree/'src/compiler/types.ts').read_bytes()==original
 result['source_restored_exactly']=True;summary.write_text(json.dumps(result,indent=2)+'\n');print('PASS '+name+' mutant target sites restored: '+str(len(a)),flush=True)

def verify_landing(sha,name,parent):
 assert git('rev-parse',sha+'^')==parent
 files=git('diff-tree','--no-commit-id','--name-only','-r',sha).splitlines();allowed='stage3/adapt/'+name+'/'
 assert all(f=='stage3/upstream/LEDGER.md' or f.startswith(allowed) for f in files),files
 old=git('show',parent+':stage3/upstream/LEDGER.md')+'\n';new=git('show',sha+':stage3/upstream/LEDGER.md')+'\n';assert new.startswith(old) and len(new[len(old):].splitlines())==1

import os,sys,json,time,subprocess,shutil,hashlib,re,gzip,signal,traceback,datetime,collections,fcntl
from pathlib import Path
R=Path('/tmp/family5-land'); W=Path('/tmp/family5-control'); MAIN='fbe5b72ba115a8f3060d4b5229fd578574f4505f'; C02='28a2e22015d0cf51667fcefa7acb246cab553817'; API='unittests:: Public APIs for typescript.d.ts should be acknowledged when they change'
lock=(W/'workflow.lock').open('w');fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
os.chdir(R);os.environ.update(ADAMIC_TYPESCRIPT_SOURCE='/workspace/mapset-upstream',ADAMIC_BUILD_CACHE_DIR='/workspace/family5-build-cache',NODE_PATH=str(R/'stage3/api/node_modules'))
statefile=W/'state.json';state=json.loads(statefile.read_text()) if statefile.exists() else {'main':MAIN,'c02':C02,'lanes':{},'pushes':{},'comparisons':{}};state.pop('stop',None)
def save():statefile.write_text(json.dumps(state,indent=2)+'\n')
class Stop(Exception):pass
def git(*args):return subprocess.check_output(['git',*args],cwd=R,text=True).strip()
def checkout(sha):subprocess.run(['git','switch','--detach',sha],cwd=R,check=True)
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
scout=json.loads(Path('/workspace/family5-analysis/sites.json').read_text())
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

def timeout_lane(label,folder,exitcode):
 report=json.loads((folder/'report.json').read_text());oracle=json.loads((folder/'oracle/report.json').read_text());failures=oracle.get('failures',[])
 extra=[f for f in failures if f['title']!=API];timed=bool(extra) and all(f.get('cause')=='timeout' for f in extra)
 result={'folder':str(folder),'exit_code':exitcode,'counts':report.get('counts'),'timeouts_only_beside_sanction':timed,'timeout_titles':[f['title'] for f in extra if f.get('cause')=='timeout'],'failures':[{k:f.get(k) for k in ['title','cause','elapsed_ms','timeout_ms','message']} for f in failures],'test_exit':oracle.get('phases',{}).get('tests',{}).get('exit')}
 state['lanes'][label]=result;save()
 return result

def require_box_pair(candidate_label,main_label):
 c=state['lanes'][candidate_label];m=state['lanes'][main_label]
 if m['exit_code']==0:raise Stop(main_label+' is green where '+candidate_label+' is red; user instructed a stop')
 if not (c['timeouts_only_beside_sanction'] and m['timeouts_only_beside_sanction']):raise Stop('Lane pair has failures other than timeouts beside the sanctioned API failure')
 c['accepted_box_red_with']=main_label;save();print('BOX-RED ACCEPTED '+candidate_label+' real_child_exit='+str(c['exit_code'])+' beside '+main_label+' real_child_exit='+str(m['exit_code']),flush=True)

def cleanup():
 if state.get('cleanup_done'):return
 paths=[Path('/workspace/family5-c02-lane/adapted-tree')]+list(Path('/workspace/family5-build-cache/local').glob('.building-*'))
 for p in paths:
  if p.exists():shutil.rmtree(p)
 state['cleanup_done']=True;save()

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

def push_landing(name,sha,branch):
 run(name+'/push',['git','push','origin',sha+':refs/heads/'+branch]);state['pushes'][branch]=sha;save()

def setup_c03():
 if state.get('c03'):checkout(state['c03']);return state['c03']
 checkout(C02);branch='typescript/type-constraint-1'
 found=subprocess.run(['git','rev-parse','--verify',branch],stdout=subprocess.PIPE,stderr=subprocess.DEVNULL,text=True)
 if found.returncode==0:
  sha=found.stdout.strip();verify_landing(sha,'99-correctness-type-constraint',C02);checkout(sha);state['c03']=sha;save();return sha
 subprocess.run(['git','switch','-c',branch,C02],check=True)
 pinlog=W/'c03/pin-search.txt';pinlog.parent.mkdir(parents=True,exist_ok=True)
 with pinlog.open('w') as f:subprocess.run(['rg','-n',r'\bType\b|types\.ts','stage3/adapt','--glob','adapt.cjs','--glob','owners.json','--glob','targets.json'],stdout=f,check=False)
 d=R/'stage3/adapt/99-correctness-type-constraint';assert not d.exists();shutil.copytree('/tmp/family5-c03-draft',d)
 p=R/'stage3/upstream/LEDGER.md';p.write_text(p.read_text()+'- **Type-only (#pncjpqy, C03):** `99-correctness-type-constraint` declares Type\'s internal optional `constraint?: Type`, matching all compiler constraint writes; AST TypeParameterDeclaration.constraint is separate. Adaptation 72\'s compiled-program certificate; JavaScript and public stripped declarations unchanged.\n')
 subprocess.run(['git','add',str(d),'stage3/upstream/LEDGER.md'],check=True);subprocess.run(['git','commit','-m','stage3: declare the internal Type constraint slot'],check=True)
 sha=git('rev-parse','HEAD');verify_landing(sha,'99-correctness-type-constraint',C02);state['c03']=sha;save();return sha

def product(name):run(name+'/product',['go','test','./stage3/lane/adaptedtree','-run','^TestProduct_Stage3AdaptedTree$','-count=1','-timeout','45m','-v'])

def copy_lane_records(name,result):
 folder=Path(result['folder']);dest=W/'lane-records'/name
 if dest.exists():return
 for p in folder.rglob('*'):
  if not p.is_file():continue
  rel=p.relative_to(folder)
  if rel.parts[0]=='adapted-tree' or 'node_modules' in rel.parts:continue
  if p.suffix not in ['.json','.jsonl','.log','.md','.txt','.diff']:continue
  target=dest/rel;target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,target)

def publish():
 # Called after the active proof child has ended, including on failure.
 for name,result in state['lanes'].items():copy_lane_records(name,result)
 for sha,label in [(C02,'c02'),(state.get('c03'),'c03')]:
  if sha:(W/(label+'-candidate.patch')).write_text(git('show','--format=fuller',sha)+'\n')
 shutil.copyfile('/tmp/family5-stock.cjs',W/'stock.cjs');shutil.copyfile('/tmp/family5-contract.cjs',W/'contract.cjs');shutil.copyfile('/tmp/family5-evidence/setup/tool-provenance.json',W/'tool-provenance.json')
 lines=['# Family 5 C02/C03: main-control continuation','', ('Stopped: '+state['stop']) if state.get('stop') else 'All requested gates completed; authorized landings pushed.','',f'Main `{MAIN}`; C02 `{C02}`; C03 `{state.get("c03","not created")}`. TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`. All proof children in this continuation ran one at a time on the same four-core/16 GiB box.','', '## Main controls and candidate lanes','', '| Run | Real exit | Passing | Failing | Pending | Classification |','|---|---:|---:|---:|---:|---|']
 for name,v in state['lanes'].items():
  c=v.get('counts') or {};classification='box-red beside '+v['accepted_box_red_with'] if v.get('accepted_box_red_with') else 'green' if v['exit_code']==0 else 'red control or unresolved red'
  lines.append(f'| {name} | {v["exit_code"]} | {c.get("passing")} | {c.get("failing")} | {c.get("pending")} | {classification} |')
  for f in v['failures']:
   if f.get('cause')=='timeout':lines.extend(['',name+': `'+f['title']+'`; '+str(f.get('message'))+'; elapsed '+str(f.get('elapsed_ms'))+' ms.'])
 lines+=['', 'Box-red acceptance follows the user\'s explicit instruction: a candidate\'s timeout-red lane is accepted only beside a timeout-red main control on this same box. No timeout or worker-count override was made. A green main control stops the candidate. Ordinary lane and npm child exits remain recorded as their real nonzero values.','', '## Every actual proof child','', '| Proof receipt | Result | Real child exit |','|---|---|---:|']
 for p in sorted(W.rglob('status.json')):
  if 'lane-records' in p.parts:continue
  s=json.loads(p.read_text());lines.append(f'| {p.parent.relative_to(W)} | {"PASS" if s["exit_code"]==0 else "FAIL / see lane classification"} | {s["exit_code"]} |')
 lines+=['', 'Stock checks enforce 79 roots and zero diagnostics in each configuration. Public compares use cmp on independently built declarations. Product commands use -timeout 45m. Full censuses use LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1, CommonJSBundle, exact absolute LATENT_ONLY_FILE, and the tsc.ts entry. Complete streams and stderr are retained.','', '## Census outcomes and every non-gone finding','']
 for label,path in state['comparisons'].items():
  d=json.loads((W/path).read_text());lines+=['',label+':','```json',json.dumps(d['files'],indent=2),'```']
  for change in d['changes']:
   if change['grade']=='gone':continue
   f=change.get('after',{});lines+=['', '- **'+change['grade']+'** at `'+str(f.get('where'))+'`, unit `'+str(f.get('unit'))+'`, phase `'+str(f.get('phase'))+'`: '+str(f.get('reason')),'', '  '+str(f.get('text'))]
 for p in sorted(W.rglob('comparison.json')):
  if str(p.relative_to(W)) in state['comparisons'].values():continue
  lines+=['', 'Unaccepted comparison retained at `'+str(p.relative_to(W))+'`:']
  for change in json.loads(p.read_text()).get('changes',[]):
   if change['grade']=='gone':continue
   f=change.get('after',{});lines+=['', '- **'+change['grade']+'** at `'+str(f.get('where'))+'`, unit `'+str(f.get('unit'))+'`, phase `'+str(f.get('phase'))+'`: '+str(f.get('reason')),'', '  '+str(f.get('text'))]
 for name in ['c02','c03']:
  p=W/name/'mutant-result.json'
  if p.exists():lines+=['',name+' mutant:','```json',p.read_text(),'```']
 lines+=['', '## Publication','',json.dumps(state.get('pushes',{}),indent=2),'', 'Each landing commit is verified to contain only its adaptation directory and one appended ledger line, with C03 directly parented by C02. No internal/ change. The prior five missing results remain accounted for in the earlier report: six test bodies were skipped after one failed before-all hook, which itself counted as one failure; net loss five.','', 'A required proof absent from this receipt table was not run or is not complete; no success is implied. If a census comparison did not complete, target clearance and zero new findings remain unestablished. Previous reports/evidence remain in Git history and previous-control-report.md.']
 (W/'REPORT.md').write_text('\n'.join(lines)+'\n');save()
 # Freeze a final current-state marker, and losslessly compress only after all timed work has stopped.
 (W/'current.json').write_text(json.dumps({'finished':True,'stop':state.get('stop'),'pushes':state['pushes']},indent=2)+'\n')
 for p in list(W.rglob('*')):
  if p.is_file() and p.suffix in ['.log','.jsonl','.diff']:
   with p.open('rb') as src,gzip.GzipFile(filename=str(p)+'.gz',mode='wb',mtime=0) as dst:shutil.copyfileobj(src,dst)
   p.unlink()
 subprocess.run(['git','switch','typescript/family5-c02c03-evidence'],check=True);subprocess.run(['git','pull','--ff-only','origin','typescript/family5-c02c03-evidence'],check=True)
 root=R/'stage3/census/family5-c02c03';dest=root/'main-control'
 # Every continuation commit retains its previous frozen evidence rather than overwriting it.
 if dest.exists():dest=root/('main-control-'+str(time.time_ns()))
 shutil.copytree(W,dest,ignore=shutil.ignore_patterns('workflow.lock'))
 (root/'previous-control-report.md').write_text((root/'REPORT.md').read_text());(root/'REPORT.md').write_text((W/'REPORT.md').read_text().replace('main-control/','main-control/'))
 # Root report links point to this continuation's evidence directory.
 prefix=dest.relative_to(root).as_posix();text=(root/'REPORT.md').read_text();text=text.replace('## Every actual proof child','Complete continuation evidence: ['+prefix+']('+prefix+'/REPORT.md).\n\n## Every actual proof child');(root/'REPORT.md').write_text(text)
 (root/'summary.json').write_text(json.dumps(state,indent=2)+'\n')
 subprocess.run(['git','add',str(root)],check=True);subprocess.run(['git','commit','-m','stage3: record main controls and sequential C02/C03 proof gates'],check=True);subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/family5-c02c03-evidence'],check=True)
 final={'evidence':git('rev-parse','HEAD'),'main':MAIN,'c02':C02,'c03':state.get('c03'),'pushes':state['pushes'],'stop':state.get('stop')};(Path('/tmp/family5-control-final.json')).write_text(json.dumps(final,indent=2)+'\n');print('FINAL '+json.dumps(final),flush=True)

try:
 # Wait for the already-started lone main control; no other proof starts before its actual child exits.
 p=W/'main/lane-alone/status.json';n=0
 while not p.exists():
  time.sleep(5);n+=1
  if n%10==0:print('RUNNING initial main control; waiting for its real exit',flush=True)
 m=json.loads(p.read_text());timeout_lane('main-control-c02',Path('/workspace/family5-main-lane-alone'),m['exit_code'])
 c=json.loads(Path('/tmp/family5-sequential/c02/lane-alone/status.json').read_text());timeout_lane('c02',Path('/workspace/family5-c02-lane-alone'),c['exit_code']);require_box_pair('c02','main-control-c02')
 cleanup();checkout(C02);verify_landing(C02,'99-correctness-line-and-character',MAIN)
 provenance=json.loads(Path('/tmp/family5-evidence/setup/tool-provenance.json').read_text());assert provenance['source_sha']==MAIN
 for name,h in provenance['tools'].items():assert digest('/tmp/family5-tools/'+name)==h
 base=Path('/tmp/family5-base-tree');tree=Path('/workspace/family5-c02-tree');files=['utilities.ts','parser.ts','program.ts','checker.ts']
 product('c02');stock_and_public('c02',base,tree)
 # Fresh whole apply receipt; its source is checked against the canonical census input.
 if not (W/'c02/whole-apply-source.json').exists():
  out=Path('/workspace/family5-c02-apply-check');run('c02/whole-apply',['bash','stage3/apply.sh',out]);assert source_manifest(out)==source_manifest(tree)
  (W/'c02/whole-apply-source.json').write_text(json.dumps({'all_source_files_identical':True,'source_files':len(source_manifest(tree))},indent=2)+'\n');shutil.rmtree(out)
 paired('c02',base,tree,files,['G02']);mutant('c02',R/'stage3/adapt/99-correctness-line-and-character',base,tree,files,{'G02'})
 state['c02_all_gates_accepted']=True;save();push_landing('c02',C02,'typescript/line-and-character-1')
 sha=setup_c03();run('c03/adapter-contract',['node','/tmp/family5-contract.cjs',R/'stage3/adapt/99-correctness-type-constraint',tree/'src/compiler/types.ts','/tmp/family5-control-contract-c03']);tree3=Path('/workspace/family5-c03-tree');run('c03/whole-apply',['bash','stage3/apply.sh',tree3])
 stock_and_public('c03',tree,tree3)
 code=run('c03/lane-alone',['bash','stage3/lane/run.sh','/workspace/family5-c03-lane-alone'],allow=True);candidate=timeout_lane('c03',Path('/workspace/family5-c03-lane-alone'),code)
 if code:
  checkout(MAIN);control=Path('/workspace/family5-main-control-for-c03');rc=run('main/control-for-c03',['bash','stage3/lane/run.sh',control],allow=True);timeout_lane('main-control-c03',control,rc);require_box_pair('c03','main-control-c03');checkout(sha)
 product('c03');paired('c03',tree,tree3,['checker.ts'],['G03','G12'],reuse=W/'c02-after');mutant('c03',R/'stage3/adapt/99-correctness-type-constraint',tree,tree3,['checker.ts'],{'G03','G12'})
 state['c03_all_gates_accepted']=True;save();push_landing('c03',sha,'typescript/type-constraint-1')
except Exception as error:
 state['stop']=type(error).__name__+': '+str(error);save();(W/'STOP.txt').write_text(str(error)+'\n'+traceback.format_exc());print('STOP '+str(error),flush=True)
try:publish()
except Exception:
 (W/'PUBLISH-ERROR.txt').write_text(traceback.format_exc());traceback.print_exc();sys.exit(1)

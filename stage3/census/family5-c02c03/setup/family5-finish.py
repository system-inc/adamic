import os,sys,time,json,subprocess,shutil,gzip,hashlib,signal,re,traceback
from pathlib import Path
R=Path('/tmp/family5-land');E=Path('/tmp/family5-evidence');MAIN='fbe5b72ba115a8f3060d4b5229fd578574f4505f';C02=subprocess.check_output(['git','rev-parse','HEAD'],cwd=R,text=True).strip();C03=None
os.chdir(R);os.environ.update(ADAMIC_TYPESCRIPT_SOURCE='/workspace/mapset-upstream',ADAMIC_BUILD_CACHE_DIR='/workspace/family5-build-cache',NODE_PATH=str(R/'stage3/api/node_modules'))
class Stop(Exception):pass
def command(cmd,cwd=R):
 return subprocess.check_output(cmd,cwd=cwd,text=True).strip()
def proof(name,cmd,env=None,cwd=R):
 code=subprocess.call([sys.executable,'/tmp/family5-proof.py',str(E/name)]+list(map(str,cmd)),cwd=cwd,env=env)
 if code:raise Stop(f'{name}: child exit {code}; see stdout.log and stderr.log')
def wait_status(paths):
 paths=[E/p/'status.json' for p in paths]
 while True:
  for p in paths:
   if p.exists() and json.loads(p.read_text())['exit_code']!=0:raise Stop(str(p.relative_to(E))+': '+str(json.loads(p.read_text())['exit_code']))
  if all(p.exists() for p in paths):return
  time.sleep(5)
def census(phase,tree,files):
 env=dict(os.environ,FAMILY5_TREE=str(tree));proof(phase+'/driver',[sys.executable,'/tmp/family5-census.py',phase,*files],env)
def targets(folder,groups):
 reasons={s['reason'] for s in json.loads((E/'setup/scout-sites.json').read_text()) if s['group'] in groups};result=set()
 for f in folder.glob('*/latent.jsonl'):
  rows=[json.loads(l) for l in f.read_text().splitlines()];assert rows[-1].get('status')=='relation_report'
  for row in rows:
   for x in row.get('findings',[]):
    if x['reason'] in reasons:result.add((f.parent.name,re.sub(r'^.*?/src/','src/',x['where']),x['reason']))
 return result
def mutant(name,adapter,tree,files,base,groups):
 old=(tree/'src/compiler/types.ts').read_bytes()
 proof(name+'/mutant-restore',['node',adapter/'mutant.cjs',tree]);census(name+'-mutant',tree,files)
 before=targets(base,groups);after=targets(E/(name+'-mutant'),groups)
 receipt={'before_target_sites':len(before),'mutant_target_sites':len(after),'target_sites_exactly_restored':before==after,'missing':sorted(before-after),'extra':sorted(after-before)}
 (E/name/'mutant-comparison.json').write_text(json.dumps(receipt,indent=2)+'\n')
 print(('PASS' if before==after else 'FAIL')+' '+name+' mutant target sites restored',flush=True)
 if not before or before!=after:raise Stop(name+' mutant did not exactly restore target sites')
 proof(name+'/reapply',['node',adapter/'adapt.cjs',tree]);assert (tree/'src/compiler/types.ts').read_bytes()==old

def lane_receipt(name):
 folder=Path('/workspace/family5-'+name+'-lane');dest=E/name/'lane-receipts';dest.mkdir(exist_ok=True)
 for file in folder.rglob('*'):
  if not file.is_file():continue
  rel=file.relative_to(folder)
  if rel.parts[0]=='adapted-tree' or 'node_modules' in rel.parts:continue
  if file.suffix not in ['.json','.log','.md','.txt']:continue
  target=dest/rel;target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(file,target)
 report=json.loads((folder/'report.json').read_text());assert report['status']=='pass',report
 # The lane checker enforces the exact platform counts and sanctioned failure.
 return report

def after_checks(name,tree,before_api):
 d=E/name
 proof(name+'/stock',['node','/tmp/family5-stock.cjs',tree,d/'stock.json'])
 proof(name+'/stock-strict',['node','/tmp/family5-stock.cjs',tree,d/'stock-strict.json'],dict(os.environ,FAMILY5_STOCK_STRICT='1'))
 roots=json.loads((d/'stock.json').read_text())['roots'];proof(name+'/adamic-types',['/tmp/family5-tools/adamic','types',*[tree/p for p in roots]])
 link=tree/'node_modules';assert not link.exists();link.symlink_to(R/'stage3/typescript/node_modules',target_is_directory=True)
 try:proof(name+'/public-build',['node',R/'stage3/typescript/node_modules/hereby/bin/hereby.js','local'],cwd=tree)
 finally:link.unlink()
 after=(tree/'built/local/typescript.d.ts').read_bytes();before=gzip.decompress(before_api.read_bytes());(d/'public-after.d.ts.gz').write_bytes(gzip.compress(after,mtime=0));b=d/'public-before.d.ts';b.write_bytes(before)
 proof(name+'/public-byte-comparison',['cmp',b,tree/'built/local/typescript.d.ts']);b.unlink()
 (d/'public-comparison.json').write_text(json.dumps({'byte_identical':before==after,'before_sha256':hashlib.sha256(before).hexdigest(),'after_sha256':hashlib.sha256(after).hexdigest()},indent=2)+'\n')

def compare(name,base,head,groups):
 proof(name+'/census-comparison',[sys.executable,'/tmp/family5-compare.py',base,head,E/name/'comparison.json',*groups])

def kill_current():
 rows=[]
 for line in command(['ps','-eo','pid,ppid,args']).splitlines()[1:]:
  parts=line.strip().split(None,2)
  if len(parts)==3:rows.append((int(parts[0]),int(parts[1]),parts[2]))
 roots={pid for pid,parent,args in rows if '/tmp/family5-proof.py /tmp/family5-evidence/' in args and 'ps -eo' not in args and args.startswith(('/opt/','python'))}
 descendants=set(roots)
 while True:
  added={pid for pid,parent,args in rows if parent in descendants}-descendants
  if not added:break
  descendants |= added
 for pid,parent,args in reversed(rows):
  if pid in descendants-roots:
   try:os.kill(pid,signal.SIGTERM)
   except ProcessLookupError:pass
 time.sleep(3)

def package(reason):
 global C03
 kill_current()
 shutil.copyfile('/tmp/family5-finish.py',E/'setup/family5-finish.py')
 # Preserve all candidates even when their landing is intentionally withheld.
 for sha,name in [(C02,'c02'),(C03,'c03')]:
  if sha:(E/(name+'-candidate.patch')).write_text(command(['git','show','--format=fuller',sha])+'\n')
 for name in ['c02','c03']:
  folder=Path('/workspace/family5-'+name+'-lane')
  if (folder/'report.json').exists():
   try:lane_receipt(name)
   except Exception:pass
 lines=['# Family 5 C02/C03 landing proofs','',f'Origin/main at start: `{MAIN}`. TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. Submodules and tool hashes are in setup/.','',f'C02 candidate: `{C02}`. C03 candidate: `{C03 or "not created"}`. C03 is directly parented by C02 when created.','',('Stopped: '+reason) if reason else 'All required proof gates passed. Landings pushed without force.','', 'Only each adaptation directory and one appended ledger line occur in its landing commit. No internal/ or compiler-lowering change. The complete owner pins, slot/heritage ownership distinction, premises and escapes are in the candidate README and pin-search evidence.','', '## Proof receipts','', '| Proof | Result | Real child exit |','|---|---|---:|']
 for f in sorted(E.rglob('status.json')):
  s=json.loads(f.read_text());lines.append(f'| {f.parent.relative_to(E)} | {"PASS" if s["exit_code"]==0 else "FAIL"} | {s["exit_code"]} |')
 lines+=['','Both stock configurations enforce 79 roots and zero diagnostics. The public comparison uses cmp. Census environment: LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1, CommonJSBundle, exact absolute LATENT_ONLY_FILE, tsc.ts entry. Whole-apply and product test are separate receipts; product uses -timeout 45m. The unfiltered lane checker enforces 106,366 passing and one sanctioned API acknowledgement failure (the oracle child exits 1, while the validated lane exits 0).','', '## Census changes and reached stops','']
 for name in ['c02','c03']:
  f=E/name/'comparison.json'
  if not f.exists():lines +=[name+': comparison not established.'];continue
  c=json.loads(f.read_text());lines +=[name+':','', '```json',json.dumps(c['files'],indent=2),'```','']
  for x in c['changes']:
   if x['grade']=='gone':continue
   v=x.get('after',{});lines +=[f'- **{x["grade"]}**, `{v.get("where")}`, owning unit `{v.get("unit")}`, phase `{v.get("phase")}`: '+v.get('reason',''),'', '  '+v.get('text',''),'']
 for name in ['c02','c03']:
  f=E/name/'lane-receipts/report.json'
  if f.exists():lines+=['',name+' complete lane verdict:', '```json',f.read_text(),'```']
 lines+=['','Large streams and logs are gzip-compressed losslessly. Each status.json records the exact command, environment, cwd, wall time and real child exit. Baseline early stock was taken during adaptation and is explicitly preliminary; baseline-stock-settled is the valid baseline. The debugger attempt did not attach because libpython3.10 was missing.']
 (E/'REPORT.md').write_text('\n'.join(lines)+'\n')
 # Losslessly compress logs/streams; leave status and summary JSON directly readable.
 for f in list(E.rglob('*')):
  if f.is_file() and f.suffix in ['.log','.jsonl']:
   with f.open('rb') as src,gzip.GzipFile(filename=str(f)+'.gz',mode='wb',mtime=0) as dst:shutil.copyfileobj(src,dst)
   f.unlink()
 command(['git','switch','-c','typescript/family5-c02c03-evidence',MAIN])
 dest=R/'stage3/census/family5-c02c03';shutil.copytree(E,dest)
 command(['git','add',str(dest)]);print(command(['git','commit','-m','stage3: retain C02/C03 landing proof evidence']),flush=True)
 print(command(['git','push','origin','HEAD:refs/heads/typescript/family5-c02c03-evidence']),flush=True)
 result={'main':MAIN,'c02':C02,'c03':C03,'evidence':command(['git','rev-parse','HEAD']),'stopped':reason};Path('/tmp/family5-final.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)

reason=None
try:
 files=['program.ts','utilities.ts','parser.ts','checker.ts']
 statuses=['c02/adapted-tree-product','c02/whole-lane']+[f'{phase}/{f[:-3]}' for phase in ['c02-before','c02-after'] for f in files]
 wait_status(statuses);lane_receipt('c02');compare('c02',E/'c02-before',E/'c02-after',['G02'])
 mutant('c02',R/'stage3/adapt/99-correctness-line-and-character',Path('/workspace/family5-c02-tree'),files,E/'c02-before',{'G02'})
 print('PASS all C02 proof gates',flush=True)
 command(['git','switch','-c','typescript/type-constraint-1',C02]);d=R/'stage3/adapt/99-correctness-type-constraint';shutil.copytree('/tmp/family5-c03-draft',d)
 ledger=R/'stage3/upstream/LEDGER.md';ledger.write_text(ledger.read_text()+'- **Type-only (#pncjpqy, C03):** `99-correctness-type-constraint` declares Type\'s internal optional `constraint?: Type`, matching all compiler constraint writes; AST TypeParameterDeclaration.constraint is separate. Adaptation 72\'s compiled-program certificate; JavaScript and public stripped declarations unchanged.\n')
 command(['git','add',str(d),'stage3/upstream/LEDGER.md']);command(['git','commit','-m','stage3: declare the internal Type constraint slot']);C03=command(['git','rev-parse','HEAD']);tree=Path('/workspace/family5-c02-tree')
 proof('c03/adapter-contract',['node','/tmp/family5-contract.cjs',d,tree/'src/compiler/types.ts','/tmp/family5-c03-contract-final'])
 proof('c03/apply-live',['node',d/'adapt.cjs',tree]);after_checks('c03',tree,E/'c02/public-after.d.ts.gz')
 before=E/'c03-before';before.mkdir();shutil.copytree(E/'c02-after/checker',before/'checker')
 census('c03-after',tree,['checker.ts']);compare('c03',before,E/'c03-after',['G03','G12'])
 mutant('c03',d,tree,['checker.ts'],before,{'G03','G12'})
 proof('c03/whole-apply',['bash','stage3/apply.sh','/workspace/family5-c03-tree'])
 proof('c03/adapted-tree-product',['go','test','./stage3/lane/adaptedtree','-run','^TestProduct_Stage3AdaptedTree$','-count=1','-timeout','45m','-v'])
 proof('c03/whole-lane',['bash','stage3/lane/run.sh','/workspace/family5-c03-lane']);lane_receipt('c03')
 for sha,branch in [(C02,'typescript/line-and-character-1'),(C03,'typescript/type-constraint-1')]:print(command(['git','push','origin',sha+':refs/heads/'+branch]),flush=True)
except Exception as err:
 reason=str(err);print('STOP '+reason,flush=True);(E/'STOP.txt').write_text(reason+'\n'+traceback.format_exc())
try:package(reason)
except Exception:
 (E/'PACKAGE-ERROR.txt').write_text(traceback.format_exc());traceback.print_exc();sys.exit(1)

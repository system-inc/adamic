import pathlib,subprocess,os,json,time,signal,shutil
R=pathlib.Path('/workspace/scratch/core-wave');env=dict(os.environ,NODE_PATH='/workspace/adamic/stage3/api/node_modules');bins=R/'bin'
def run(label,args,cwd=None,env=env,limit=1700):
 begin=time.monotonic()
 with (R/(label+'.log')).open('w') as log:
  p=subprocess.Popen(args,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=limit)
  except subprocess.TimeoutExpired:code=124
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 record={'command':list(map(str,args)),'cwd':str(cwd),'exit':code,'seconds':time.monotonic()-begin}
 (R/(label+'-result.json')).write_text(json.dumps(record,indent=2)+'\n');print(label,code,flush=True)
 return code
mode=__import__('sys').argv[1]
selected=__import__('sys').argv[2:] or ['before','clone','groupby']
if mode=='types':
 for name in selected:
  tree=pathlib.Path('/workspace/core-'+name);roots=sorted(map(str,(tree/'src/compiler').rglob('*.ts')));assert len(roots)==79,len(roots)
  (R/('roots-'+name+'.json')).write_text(json.dumps(roots,indent=2)+'\n');assert run('types-'+name,[str(bins/'adamic'),'types',*roots])==0
elif mode=='census':
 for name in selected:
  tree=pathlib.Path('/workspace/core-'+name);e=dict(env,LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(R/('census-'+name+'-times.jsonl')))
  assert run('census-'+name,[str(bins/'latent'),str(tree/'src/compiler'),str(R/('census-'+name+'.jsonl'))],env=e)==0
 for name,repo in [(n,r) for n,r in [('clone','/workspace/adamic'),('groupby','/workspace/adamic-groupby')] if n in selected]:
  source=pathlib.Path('/workspace/core-'+name);tree=R/('mutant-'+name);shutil.copytree(source/'src',tree/'src');pin=json.loads((pathlib.Path(repo)/('stage3/adapt/99-correctness-core-'+name)/'owners.json').read_text())[0]
  f=tree/'src/compiler/core.ts';text=f.read_bytes().decode();spell=lambda s:s.replace('\n','\r\n') if '\r\n' in text else s;assert text.count(spell(pin['after']))==1;f.write_bytes(text.replace(spell(pin['after']),spell(pin['before'])).encode())
  e=dict(env,LATENT_FULL='1',LATENT_ONLY_FILE=str(f),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(R/('census-mutant-'+name+'-times.jsonl')))
  assert run('census-mutant-'+name,[str(bins/'latent'),str(tree/'src/compiler'),str(R/('census-mutant-'+name+'.jsonl'))],env=e)==0
elif mode=='build':
 # A shared lockfile install is safe because every compiler checkout is the same pin.
 packages=R/'packages';packages.mkdir(exist_ok=True)
 for f in ['package.json','package-lock.json']:shutil.copyfile('/workspace/typescript-source/'+f,packages/f)
 if not (packages/'node_modules').exists():assert run('npm-compiler',['npm','ci','--no-audit','--no-fund'],cwd=packages)==0
 for name in selected:
  tree=pathlib.Path('/workspace/core-'+name);(tree/'node_modules').symlink_to(packages/'node_modules',target_is_directory=True)
  assert run('build-'+name,['npm','run','build:compiler'],cwd=tree)==0
 for name,repo in [(n,r) for n,r in [('clone','/workspace/adamic'),('groupby','/workspace/adamic-groupby')] if n in selected]:
  directory=repo+'/stage3/adapt/99-correctness-core-'+name
  assert run('compare-'+name,['node',directory+'/compare.cjs','/workspace/core-before','/workspace/core-'+name,'/workspace/core-'+name])==0
elif mode=='contracts':
 for name,repo in [(n,r) for n,r in [('clone','/workspace/adamic'),('groupby','/workspace/adamic-groupby')] if n in selected]:
  directory=repo+'/stage3/adapt/99-correctness-core-'+name
  assert run('contract-'+name,['node',directory+'/contract.cjs','/workspace/core-before'])==0

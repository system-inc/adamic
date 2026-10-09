import pathlib,subprocess,json,time,os,concurrent.futures
out=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-css-print'); roots=json.loads((out/'port-validation-worktrees.json').read_text())
for r in roots:
 p=pathlib.Path(r['root'])/'stage3/api/node_modules'
 if not p.exists(): p.symlink_to('/workspace/adamic/stage3/api/node_modules',target_is_directory=True)
def check(r):
 ident=r['id']; env=os.environ.copy(); env.pop('ADAMIC_MUTANT',None); env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u079/cache/standalone-'+ident
 log=out/(ident+'-standalone-native.log'); cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run','^TestProduct_CSSPrinterSanitizedAndLowered$']; start=time.time()
 with log.open('w') as f: code=subprocess.run(cmd,cwd=r['root'],env=env,stdout=f,stderr=subprocess.STDOUT).returncode
 return dict(id=ident,command=cmd,root=r['root'],code=code,wall=time.time()-start,log=str(log.relative_to('/workspace/adamic')))
records=[]
# Two builds at a time keep each validation within its 90-second binary budget.
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for record in pool.map(check,roots):
  records.append(record); (out/'port-native-validations.json').write_text(json.dumps(records,indent=2))

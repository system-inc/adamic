import json,os,subprocess,time,sys
from pathlib import Path
root=Path.cwd();out=Path(__file__).resolve().parent;cache=Path(os.environ['GOCACHE']);os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4])
rows=json.loads((out/'mutant-catalog.json.txt').read_text())
rows += [dict(name='primitive-missing-nominal-typeerror',changes=[['internal/native/runtime/exceptions.c','adamic_object *error = adamic_error_new_kind(text, "TypeError");','adamic_object *error = adamic_error_new(text);']],selection='^TestOrdinaryToPrimitive$',catcher='C differs from Node',package='internal/native',replace_all=False),dict(name='spread-undefined-pending-lost',changes=[['internal/native/runtime/object.c','if (adamic_exception_pending) {','if (adamic_thrown != NULL) {']],selection='^TestSpreadUndefinedThrowStopsBeforeNextGetter$',catcher='spread:',package='internal/oracle',replace_all=False)]
results=json.loads((out/'mutant-results.json.txt').read_text()) if (out/'mutant-results.json.txt').exists() else []
for m in rows:
 if any(r['name']==m['name'] for r in results):continue
 if len(sys.argv)>1 and m['name'] not in sys.argv[1:]:continue
 directory=out/'mutants'/m['name'];directory.mkdir(parents=True,exist_ok=True);replace={};external={}
 for relative,old,new in m['changes']:
  path=root/relative;original=path.read_text();count=original.count(old)
  assert count and (count==1 or m['replace_all']),(m['name'],relative,count)
  changed=original.replace(old,new) if m['replace_all'] else original.replace(old,new,1)
  if m['name']=='native-drops-line':changed=changed.replace("\tfputc('\\n', stderr);","\tfputc('\\n', stderr);\n\t}",1)
  target=directory/(path.name+'.txt');target.write_text(changed);replace[str(path)]=str(target)
  if relative.endswith('.mjs'):external[path]=(original,changed)
 overlay=directory/'overlay.json.txt';overlay.write_text(json.dumps({'Replace':replace},indent=2)+'\n')
 command=['go','test','-p','1','-overlay',str(overlay),'./'+m['package'],'-run',m['selection'],'-count=1','-v','-parallel=4','-timeout','90s']
 environment=dict(os.environ,ADAMIC_GATE_UNCACHED='1')
 if m['name']=='header-forces-runtime-features':
  environment.update(ADAMIC_BUILD_OUTPUT_DIR=str(directory/'products'),ADAMIC_BUILD_CACHE_DIR=str(directory/'empty-read-cache'))
 before=set(cache.glob('*/*-d'));start=time.monotonic()
 try:
  for path,(old,new) in external.items():path.write_text(new)
  with (directory/'test.log').open('w') as log:
   try:result=subprocess.run(['timeout','110',*command],stdout=log,stderr=subprocess.STDOUT,env=environment,timeout=115);code=result.returncode
   except subprocess.TimeoutExpired:code=124
 finally:
  for path,(old,new) in external.items():path.write_text(old)
 seconds=round(time.monotonic()-start,3);text=(directory/'test.log').read_text()
 caught=code!=0 and code!=124 and m['catcher'] in text and '[build failed]' not in text and 'no space left on device' not in text and 'panic: test timed out' not in text
 if m['name'] not in ['lint-profile-default-runtime','split-runtime-default-flags']:
  caught=caught and 'clang failed' not in text
 row=dict(name=m['name'],caught=caught,exit=code,seconds=seconds,catcher=m['catcher'],command=command)
 results.append(row);(out/'mutant-results.json.txt').write_text(json.dumps(results,indent=2)+'\n')
 # These are only new, unshared compiler archives from this finished overlay.
 # The unmutated cache and all immutable product artifacts remain available.
 reclaimed=0
 for path in set(cache.glob('*/*-d'))-before:
  if os.environ.get('SLICE_EVICT_MUTANT_CACHE') != '1':break
  st=path.stat()
  if st.st_size>10*1024**2 and st.st_nlink==1:reclaimed+=st.st_size;path.unlink()
 print(m['name'],caught,code,seconds,'cache reclaimed',reclaimed,flush=True)
 if not caught:sys.exit(1)

import pathlib,subprocess,os,json,time
r=pathlib.Path('/workspace/adamic');p=pathlib.Path('/tmp/def-regexp');src=r/'internal/regexp/parser.go'; original=src.read_text(); names=[x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')];result=[]
try:
 subprocess.run(['git','apply',str(p/'D08.diff')],cwd=r,check=True)
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/def-regexp/cache/D08'
 for n in names:
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/regexp/','-run','^'+n+'$']; start=time.monotonic()
  with (p/f'D08-{n}.log').open('w') as f:q=subprocess.run(cmd,cwd=r,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for line in (p/f'D08-{n}.log').read_text().splitlines():
   try:events.append(json.loads(line))
   except:pass
  outputs=[e.get('Output','').strip() for e in events if e.get('Test')==n and ('parser_test.go:'in e.get('Output','') or 'panic:' in e.get('Output',''))]
  result.append({'test':n,'exit':q.returncode,'wall_seconds':time.monotonic()-start,'evidence':outputs,'command':'ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd)+f' > D08-{n}.log 2>&1'})
  (p/'D08-isolated.json').write_text(json.dumps(result,indent=2));print(n,q.returncode,flush=True)
finally:src.write_text(original)

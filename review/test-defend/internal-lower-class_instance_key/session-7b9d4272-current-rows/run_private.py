import subprocess,pathlib,difflib,json,os,time
p=pathlib.Path(__file__).parent
mutants=[('D4','internal/lower/class_static.go','l.signature(function, member, l.thisLocal(function))','l.signature(function, member, -1)'),('D5','internal/lower/class_static.go','meta.Methods[slot] = function','meta.Methods[slot] = function - 1')]
(p/'private-plan.json').write_text(json.dumps(mutants,indent=2))
for ident,path,old,new in mutants:
 f=pathlib.Path(path); base=subprocess.check_output(['git','show','HEAD:'+path]).decode(); assert base.count(old)==1
 changed=base.replace(old,new);(p/(ident+'.diff')).write_text(''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/'+path,tofile='b/'+path)))
 try:
  f.write_text(changed)
  with (p/(ident+'-vet.log')).open('w') as log: vet=subprocess.run(['go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT)
  if vet.returncode: raise Exception('vet failed '+ident)
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/class-key-defense/cache/'+ident
  begin=time.monotonic()
  with (p/(ident+'.log')).open('w') as log: result=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],env=env,stdout=log,stderr=subprocess.STDOUT)
  (p/(ident+'-run.json')).write_text(json.dumps({'command':'ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .','returncode':result.returncode,'wall_seconds':time.monotonic()-begin}))
 finally: f.write_text(base)
 print(ident+' complete',flush=True)

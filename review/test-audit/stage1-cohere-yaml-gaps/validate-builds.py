import pathlib,subprocess,time,os,json
p=pathlib.Path('/tmp/u152');root=pathlib.Path('/workspace/adamic');helper=root/'audit_u152_build.go';assert not helper.exists();records=[]
helper.write_text((p/'build-port.go.txt').read_text())
try:
 for id,path,entry in [('M2','internal/native/runtime/string_build_impl.h','gaps/valueConjunction.ts'),('M3','internal/native/runtime/string_append.c','gaps/sharedSliceAppend.ts'),('M4','stage1/cohere/yaml/lexer.ts','lex_main.ts')]:
  f=root/path;old=f.read_text()
  try:
   subprocess.run(['git','apply',str(p/(id+'.diff'))],cwd=root,check=True)
   env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/(id+'-build-validation'))
   cmd=['timeout','90','go','run',str(helper),str(root/'stage1/cohere/yaml'/entry),str(p/(id+'-compiled-product'))];start=time.monotonic()
   with (p/(id+'-build.log')).open('w') as out:r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
   rec={'id':id,'command':' '.join(cmd),'wall':round(time.monotonic()-start,3),'exit':r.returncode,'timing':(p/(id+'-build.log')).read_text().strip()};records.append(rec);(p/'builds.json').write_text(json.dumps(records,indent=2));print(rec,flush=True);assert r.returncode==0
  finally:f.write_text(old)
finally:helper.unlink()

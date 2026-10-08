import pathlib,subprocess,json,os,sys
sys.path.insert(0,'/tmp/regex-runtime-python')
import brotli
root=pathlib.Path('/tmp/regex-runtime-postmerge-sizes');root.mkdir(exist_ok=True)
rows=[]
for version,repo,driver in [('before','/workspace/adamic','/tmp/regex-runtime-adamic-postmerge-disabled'),('after','/workspace/adamic','/tmp/regex-runtime-adamic-postmerge')]:
 out=root/version;out.mkdir(exist_ok=True)
 for name,fixture in [('hello','internal/load/testdata/0.1/compile/01_hello.ts'),('request','cmd/adamic/testdata/wasi/request.a')]+([('dynamic_gap','internal/oracle/testdata/regexp_dynamic/dynamic_gap.a')] if version=='after' else []):
  for target in ['native','wasm32-wasi']:
   binary=out/(name+('.wasm' if target!='native' else ''))
   args=[driver,'build']+(['--target',target] if target!='native' else [])+[repo+'/'+fixture,'-o',str(binary)]
   with open(out/(name+'.'+target+'.build.log'),'wb') as log:subprocess.run(args,cwd=repo,stdout=log,stderr=log,check=True)
   data=binary.read_bytes();packed=brotli.compress(data,quality=11)
   rows.append(dict(version=version,name=name,target=target,raw=len(data),brotli=len(packed)))
for name in ['hello','request']:
 a=(root/'before'/(name+'.wasm')).read_bytes();b=(root/'after'/(name+'.wasm')).read_bytes();assert a==b,(name,'WASI bytes differ')
(root/'sizes.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows,indent=2))

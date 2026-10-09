from pathlib import Path
import subprocess,json,time
out=Path('review/test-audit/stage1-cohere-markdownblocks-leaf_composition_products');res=[]
with (out/'cli-build.log').open('w') as log:r=subprocess.run(['go','build','-o','/tmp/u127-adamic','./cmd/adamic'],stdout=log,stderr=subprocess.STDOUT)
assert r.returncode==0
for mid,entry in [('M1','list_probe.ts'),('M2','list_probe.ts'),('M3','list_probe.ts'),('M4','mdast_probe.ts'),('P1','list_probe.ts'),('P2','mdast_probe.ts')]:
 diff=out/f'{mid}.diff';cmd=['git','apply',str(diff)]
 assert subprocess.run(['git','apply','--check',str(diff)]).returncode==0
 subprocess.run(cmd,check=True)
 try:
  start=time.monotonic()
  with (out/f'{mid}-native-build.log').open('w') as log:r=subprocess.run(['timeout','90','/tmp/u127-adamic','build','stage1/cohere/markdownblocks/testdata/'+entry,'-o','/tmp/u127-'+mid,'--sanitize'],stdout=log,stderr=subprocess.STDOUT)
  res.append({'id':mid,'command':['timeout','90','/tmp/u127-adamic','build','stage1/cohere/markdownblocks/testdata/'+entry,'-o','/tmp/u127-'+mid,'--sanitize'],'exit':r.returncode,'seconds':time.monotonic()-start});(out/'standalone-builds.json').write_text(json.dumps(res,indent=2))
 finally:subprocess.run(['git','apply','-R',str(diff)],check=True)

import sys
sys.dont_write_bytecode=True
from run import *
import concurrent.futures,shutil,hashlib,difflib
names=json.loads(pathlib.Path('/workspace/scratch/sys-timer-stack/main-js.json').read_text());assert len(names)==9
unit=next(u for u in json.loads((ROOT/'stage3/lane/shard-plan.json').read_text())['units'] if u['name']=='stage3-lane-byte-comparisons');(S/'api-tasks.json').write_text(json.dumps(unit['tasks']))
def build(label):
 tree=S/label;code=run(label+'-build-compiler',['npm','run','build:compiler','--','--lkg'],tree)
 assert (code==0)==(label=='H'),(label,code)
 missing=[n for n in names if not (tree/'built/local'/n).exists()]
 if missing:
  assert missing==['tsserverlibrary.js'],missing
  assert run(label+'-finish-shim',['node',S/'finish-shim.mjs',tree],tree)==0
 assert run(label+'-bundle-declarations',['node','scripts/dtsBundler.mjs','--entrypoint','./built/local/typescript/typescript.d.ts','--output','./built/local/typescript.d.ts'],tree)==0
 # Use the same fixed, already built Head 2 public-API harness for both artifact readers.
 for n in ['run.js','run.js.map']:shutil.copyfile(pathlib.Path('/workspace/scratch/sys-timer-stack/head2/built/local')/n,tree/'built/local'/n)
 code=run(label+'-api-baseline',['node',ROOT/'stage3/oracle/run-tasks.cjs',tree,S/'api-tasks.json',S/(label+'-api-worker'),'1'],tree)
 assert code==1,code
 report=json.loads((S/(label+'-api-worker/report.json')).read_text());assert report['counts']['failing']==1,report
 api=tree/'tests/baselines/local/api/typescript.d.ts';assert api.read_text()==(tree/'built/local/typescript.d.ts').read_text()
 return {'typedCompilerBuildExit':json.loads((S/(label+'-build-compiler-status.json')).read_text())['exit'],'normalBuildMissingBeforeShim':missing,'apiWorker':report}
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 jobs={pool.submit(build,label):label for label in ['H','D']};results={}
 for job in concurrent.futures.as_completed(jobs):results[jobs[job]]=job.result()
js=[]
for name in names:
 a=(S/'H/built/local'/name).read_bytes();b=(S/'D/built/local'/name).read_bytes();c=(pathlib.Path('/workspace/scratch/sys-timer-stack/head2/built/local')/name).read_bytes();assert a==c,name
 js.append({'file':name,'head2Sha256':hashlib.sha256(a).hexdigest(),'DSha256':hashlib.sha256(b).hexdigest(),'bytes':len(a),'byteIdentical':a==b})
assert all(x['byteIdentical'] for x in js)
outputs=[]
for relative in ['built/local/typescript.d.ts','built/local/typescript.internal.d.ts','tests/baselines/local/api/typescript.d.ts','tests/baselines/reference/api/typescript.d.ts']:
 a=(S/'H'/relative).read_bytes();b=(S/'D'/relative).read_bytes();name=relative.replace('/','-')+'.diff.txt';diff=''.join(difflib.unified_diff(a.decode().replace('\r\n','\n').splitlines(True),b.decode().replace('\r\n','\n').splitlines(True),fromfile='head2/'+relative,tofile='D/'+relative,n=0));(S/name).write_text(diff);outputs.append({'file':relative,'diffFile':name,'changed':a!=b,'beforeSha256':hashlib.sha256(a).hexdigest(),'afterSha256':hashlib.sha256(b).hexdigest()})
(S/'artifacts.json').write_text(json.dumps({'builds':results,'javascript':js,'javascriptCount':len(js),'allNineByteIdentical':True,'declarations':outputs,'fixedHarnessFrom':'accepted Head 2 built/local/run.js; not a D typecheck substitute'},indent=2)+'\n')
print('all nine fresh JS bundles byte-identical; actual public API baseline artifacts captured',flush=True)

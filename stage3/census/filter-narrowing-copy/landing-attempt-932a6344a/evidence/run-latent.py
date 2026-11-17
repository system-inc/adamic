import os,json,subprocess,sys
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor,as_completed
p=Path('/workspace/scratch/ev14ch9-landing-proof');out=p/'census';out.mkdir(exist_ok=True)
files=['src/compiler/core.ts','src/compiler/commandLineParser.ts','src/compiler/factory/utilities.ts','src/compiler/transformers/ts.ts','src/services/callHierarchy.ts','src/services/codefixes/fixAddMissingParam.ts','src/services/services.ts','src/compiler/checker.ts']
def run(case):
 tree,file=case;key=tree+'-'+file.replace('/','_');root=p/tree;source=root/file;scope=root/('src/services' if '/services/' in file else 'src/compiler');env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(source),CENSUS_MODULE_FORMAT='CommonJSBundle')
 with (out/(key+'.log')).open('w') as log:r=subprocess.run([str(p/'latent'),str(scope),str(out/(key+'.jsonl'))],env=env,stdout=log,stderr=subprocess.STDOUT)
 (out/(key+'.exit')).write_text(str(r.returncode)+'\n');print(('PASS' if r.returncode==0 else 'FAIL')+' census child '+key+' exit '+str(r.returncode),flush=True)
 return r.returncode
with ThreadPoolExecutor(max_workers=2) as pool:
 results=list(pool.map(run,[(tree,file) for file in files for tree in ('before','after')]))
sys.exit(next((r for r in results if r),0))

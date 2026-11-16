from pathlib import Path
import subprocess,os,json,concurrent.futures,time
root=Path('/workspace/scratch/cast-tags');e=root/'evidence'
tasks=[(side,file) for side in ['main','c'] for file in ['checker.ts','transformers/generators.ts']]
def run(spec):
 side,file=spec; tag=side+'-'+Path(file).stem; tree=root/(side+'-tree');out=e/(tag+'.jsonl');env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(tree/'src/compiler'/file),LATENT_PROGRESS_FILE=str(e/(tag+'-progress.json')))
 begin=time.time(); print('START',tag,flush=True)
 with (e/(tag+'.log')).open('w') as log:r=subprocess.run([str(root/'latent'),str(tree/'src/compiler'),str(out)],env=env,stdout=log,stderr=subprocess.STDOUT)
 result={'case':tag,'exit':r.returncode,'seconds':time.time()-begin};(e/(tag+'-status.json')).write_text(json.dumps(result)+'\n');print('END',result,flush=True);return result
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 results=list(pool.map(run,tasks))
(e/'census-status.json').write_text(json.dumps(results,indent=2)+'\n')

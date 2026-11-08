import sys
sys.dont_write_bytecode=True
from run import *
import concurrent.futures
for label in ['H','D']:
 tree=S/label
 assert run(label+'-git-init',['git','init','-q',tree])==0
 (tree/'.git/objects/info/alternates').write_text('/workspace/typescript-source/.git/objects\n')
 assert run(label+'-git-pin',['git','-C',tree,'update-ref','HEAD','050880ce59e30b356b686bd3144efe24f875ebc8'])==0
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 jobs={pool.submit(run,label+'-stock',['node','--max-old-space-size=8192',S/'check.cjs',S/label,ROOT,S/(label+'-stock.json')]):label for label in ['H','D']}
 for job in concurrent.futures.as_completed(jobs):assert job.result()==0
for label in ['H','D']:
 r=json.loads((S/(label+'-stock.json')).read_text());print(label+' adamic='+str(len(r['adamic']))+' upstream='+str(len(r['solution']))+' solutionExit='+str(r['solutionExit']),flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 jobs={pool.submit(run,label+'-latent',['/workspace/scratch/sys-timer-stack/bin/census',S/label/'src/compiler',S/(label+'-latent.jsonl')],extra={'LATENT_FULL':'1','LATENT_ONLY_FILE':str(S/label/'src/compiler/sys.ts'),'LATENT_ASSERT_NO_OUTPUT':'1','LATENT_TIMES':str(S/(label+'-times.jsonl'))}):label for label in ['H','D']}
 for job in concurrent.futures.as_completed(jobs):print(jobs[job]+' census child_exit='+str(job.result()),flush=True)

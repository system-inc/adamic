import pathlib,subprocess,sys,concurrent.futures
runtime=pathlib.Path('/workspace/scratch/namedeval-kind-runtime');out=pathlib.Path('/workspace/scratch/namedeval-kind-report/stage3/census/namedeval-kind');tree=runtime/'adapted-B'
cases=[(c,'latent',f) for f in ['namedEvaluation','esDecorators'] for c in ['B','S1','S2']]+[(c,'scan','namedEvaluation') for c in ['B','S1','S2']]
def run(job):
 case,mode,file=job;root=pathlib.Path('/workspace/scratch/namedeval-kind-'+case);name=case+'-'+mode+'-'+file
 if mode=='latent':cmd=['env','-u','LATENT_SPECULATIVE','LATENT_FULL=1','LATENT_ASSERT_NO_OUTPUT=1','CENSUS_MODULE_FORMAT=CommonJSBundle','LATENT_ONLY_FILE='+str(tree/'src/compiler/transformers'/(file+'.ts')),str(runtime/case/'latent'),str(tree/'src/compiler'),str(out/(name+'.jsonl'))]
 else:cmd=[str(runtime/case/'scan-all'),str(tree/'src/compiler/transformers/namedEvaluation.ts')]
 return subprocess.run([sys.executable,str(runtime/'run.py'),name,*cmd],cwd=root).returncode
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:codes=list(pool.map(run,cases))
sys.exit(next((c for c in codes if c),0))

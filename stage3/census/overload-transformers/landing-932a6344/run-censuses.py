import pathlib,subprocess,sys,concurrent.futures
r=pathlib.Path('/workspace/scratch/get-properties-runtime');e=pathlib.Path('/workspace/scratch/overload-transformers/stage3/census/overload-transformers/landing-932a6344')
cases=[(tree,file) for tree in ['base','candidate'] for file in ['utilities','ts','classFields']]+[('mutant','utilities')]
def run(case):
 tree,file=case;root=r/(tree+'-adapted');name='latent-'+tree+'-'+file
 cmd=['env','-u','LATENT_SPECULATIVE','LATENT_FULL=1','LATENT_ASSERT_NO_OUTPUT=1','CENSUS_MODULE_FORMAT=CommonJSBundle','LATENT_ONLY_FILE='+str(root/'src/compiler/transformers'/ (file+'.ts')),str(r/'latent'),str(root/'src/compiler'),str(e/(name+'.jsonl'))]
 return subprocess.run([sys.executable,str(r/'run.py'),name,*cmd]).returncode
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
 codes=list(pool.map(run,cases))
sys.exit(next((c for c in codes if c),0))

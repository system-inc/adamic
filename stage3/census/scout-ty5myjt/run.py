import pathlib,subprocess,os,json,time
root=pathlib.Path('/workspace/scratch/scout-ty5myjt');repo=root/'adamic'
env=os.environ.copy();env.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',TMPDIR=str(root/'tmp'))
def run(args,out,extra={}):
 start=time.time()
 with open(out,'w') as log:p=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,env=env|extra)
 print(json.dumps({'args':list(map(str,args)),'status':p.returncode,'seconds':time.time()-start}),flush=True)
 if p.returncode:raise RuntimeError(str(args))
run(['node',str(root/'measure.cjs'),str(root/'adapted'),str(root/'results/baseline')],root/'stock-baseline.log')
for f in ['program.ts','resolutionCache.ts']:
 run([str(root/'latent'),str(root/'adapted/src/compiler'),str(root/'results/baseline'/f'{f}.jsonl')],root/f'baseline-{f}.log',{'LATENT_ONLY_FILE':str(root/'adapted/src/compiler'/f)})
run(['python3',str(root/'variants.py')],root/'variants.log')
for name in ['diag-intersection','diag-readonly','diag-exact','diag-overloads','resolution-readonly','resolution-exact','resolution-result-view']:
 f='program.ts' if name.startswith('diag') else 'resolutionCache.ts';tree=root/name;out=root/'results'/name
 run(['node',str(root/'measure.cjs'),str(tree),str(out)],root/f'{name}-stock.log')
 run([str(root/'latent'),str(tree/'src/compiler'),str(out/f'{f}.jsonl')],root/f'{name}-census.log',{'LATENT_ONLY_FILE':str(tree/'src/compiler'/f)})

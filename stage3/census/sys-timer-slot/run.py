"""Run commands with separate logs and their actual child status; no pipelines."""
import subprocess, pathlib, json, time, os, sys, signal
repo=pathlib.Path('/workspace/adamic'); scratch=pathlib.Path('/workspace/scratch/sys-timer-slot'); evidence=repo/'stage3/census/sys-timer-slot'
env=dict(os.environ, NODE_PATH=str(repo/'stage3/api/node_modules'), ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source')
records=[]
def run(label,args,cwd=repo,extra=None):
    start=time.time()
    with (scratch/(label+'.log')).open('w') as log:
        p=subprocess.Popen(args,cwd=cwd,env={**env,**(extra or {})},stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
        # A bounded wall timeout is stricter than the requested no-output limit.
        try: code=p.wait(timeout=1800)
        except subprocess.TimeoutExpired:
            os.killpg(p.pid,signal.SIGTERM)
            try: code=p.wait(timeout=10)
            except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);code=p.wait()
    r=dict(label=label,args=list(map(str,args)),cwd=str(cwd),exit=code,seconds=time.time()-start)
    records.append(r);(scratch/('commands-'+(sys.argv[1] if len(sys.argv)>1 else pathlib.Path(sys.argv[0]).stem)+'.json')).write_text(json.dumps(records,indent=2)+'\n')
    print(json.dumps(r),flush=True)
    return code
if __name__=='__main__':
    mode=sys.argv[1]
    if mode=='measure':
        run('references',['node',str(evidence/'measure.cjs'),str(scratch/'main'),str(scratch/'references.json')])
    elif mode=='checks':
        for label in ['main','A','B']:
            run(label+'-stock',['node','--max-old-space-size=8192',str(evidence/'check.cjs'),str(scratch/label),str(repo),str(scratch/(label+'-stock.json'))])
    elif mode=='builds':
        for label in ['main','A','B']:
            run(label+'-build-compiler',['npm','run','build:compiler','--','--lkg'],scratch/label)
            js={str(p.relative_to(scratch/label/'built/local')):__import__('hashlib').sha256(p.read_bytes()).hexdigest() for p in (scratch/label/'built/local').rglob('*.js')}
            (scratch/(label+'-compiler-js.json')).write_text(json.dumps(js,indent=2)+'\n')
            run(label+'-build-tests',['npm','run','build:tests','--','--lkg'],scratch/label)
    elif mode=='latent-build':
        if run('overlay',['python3','stage3/census/latent/make_overlay.py',str(repo),str(scratch/'overlay')])==0 and run('overlay-txt',['python3',str(evidence/'overlay.py')])==0:
            run('latent-build',['go','build','-buildvcs=false','-overlay='+str(scratch/'overlay/overlay.json'),'-o',str(scratch/'census'),'./stage3/census/latent/tool'])
    elif mode=='latent':
        for label in ['main','A','B']:
            run(label+'-latent',[str(scratch/'census'),str(scratch/label/'src/compiler'),str(scratch/(label+'-latent.jsonl'))],extra={'LATENT_FULL':'1','LATENT_ONLY_FILE':str(scratch/label/'src/compiler/sys.ts'),'LATENT_ASSERT_NO_OUTPUT':'1','LATENT_TIMES':str(scratch/(label+'-latent-times.jsonl'))})

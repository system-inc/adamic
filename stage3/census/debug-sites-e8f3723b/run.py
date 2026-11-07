import os,sys,time,json,subprocess,signal,pathlib
repo=pathlib.Path('/workspace/adamic'); out=repo/'stage3/census/debug-sites-e8f3723b'; scratch=pathlib.Path('/workspace/scratch/debug-sites'); tree=scratch/'adapted'
def run(name,args,env=None,watch=()):
    start=time.monotonic(); last=start; sizes=None
    with (out/(name+'.log')).open('w') as log:
        p=subprocess.Popen(args,cwd=repo,env=env,start_new_session=True,stdout=log,stderr=subprocess.STDOUT)
        try:
            while p.poll() is None:
                current=tuple((path.stat().st_size if path.exists() else -1) for path in [out/(name+'.log'),*watch])
                if current!=sizes: last=time.monotonic(); sizes=current
                if time.monotonic()-last>1800:
                    os.killpg(p.pid,signal.SIGTERM); p.wait(timeout=20); break
                time.sleep(1)
        finally:
            if p.poll() is None:
                os.killpg(p.pid,signal.SIGTERM)
                try:p.wait(timeout=20)
                except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
    result={'args':args,'exit_code':p.returncode,'elapsed_seconds':time.monotonic()-start}
    (out/(name+'-result.json')).write_text(json.dumps(result,indent=2)+'\n'); print(name,json.dumps(result),flush=True)
    if p.returncode:sys.exit(p.returncode if p.returncode>0 else 1)
    return result
if __name__=='__main__':
    mode=sys.argv[1]
    env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE=str(scratch/'TypeScript'))
    if mode=='apply':run('apply',['bash','stage3/apply.sh',str(tree)],env)
    elif mode=='build':
        import re
        source=repo/'stage3/census/latent/make_overlay.py'; text=source.read_text()
        text=text.replace("scratch / name.replace('/', '_')", "scratch / (name.replace('/', '_') + '.txt')")
        text=re.sub(r"((?:scratch|output) / '[^']+\.go)'",r"\1.txt'",text)
        sys.argv=[str(source),str(repo),str(scratch/'overlay')]
        exec(compile(text,str(source),'exec'),{'__file__':str(source),'__name__':'__main__'})
        run('build',['go','build','-buildvcs=false','-overlay='+str(scratch/'overlay/overlay.json'),'-o',str(scratch/'latent'),'./stage3/census/latent/tool'])
    else:
        env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
        env.update(LATENT_ONLY_FILE=str(tree/'src/compiler/debug.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/(mode+'-times.jsonl')))
        if mode=='full':env['LATENT_FULL']='1'
        run(mode,[str(scratch/'latent'),str(tree/'src/compiler'),str(out/(mode+'.jsonl'))],env,[out/(mode+'.jsonl'),out/(mode+'-times.jsonl')])

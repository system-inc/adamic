import os,sys,subprocess,time,signal,json
from pathlib import Path
root=Path('/workspace/adamic'); work=Path('/workspace/scratch/abrq61p'); work.mkdir(exist_ok=True)

def run(name,cmd,env=None):
    log=work/(name+'.log')
    with log.open('w') as out:
        p=subprocess.Popen(cmd,cwd=root,env=dict(os.environ,**(env or {})),stdout=out,stderr=subprocess.STDOUT,start_new_session=True)
        last=time.monotonic(); size=0
        try:
            while p.poll() is None:
                now=(log.stat().st_size, Path(env['LATENT_TIMES']).stat().st_size if env and env.get('LATENT_TIMES') and Path(env['LATENT_TIMES']).exists() else 0)
                if now!=size: size=now;last=time.monotonic()
                if time.monotonic()-last>1800:
                    os.killpg(p.pid,signal.SIGTERM)
                    try:p.wait(timeout=10)
                    except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
                    print(name,'stopped after 1800 seconds silence',flush=True);break
                time.sleep(1)
        finally:
            if p.poll() is None:os.killpg(p.pid,signal.SIGKILL);p.wait()
        (work/(name+'.exit')).write_text(str(p.returncode)+'\n')
        print(name,'child exit',p.returncode,flush=True)
        if p.returncode:sys.exit(1)
    return p.returncode

if len(sys.argv)>1 and sys.argv[1]=='baseline':
    run('base-overlay',['python3','stage3/census/latent/make_overlay.py',str(root),str(work/'base-overlay')])
    run('base-gofmt',['bash','-c','gofmt -w /workspace/scratch/abrq61p/base-overlay/*.go'])
    run('base-build',['go','build','-buildvcs=false','-overlay='+str(work/'base-overlay/overlay.json'),'-o',str(work/'base-census'),'./stage3/census/latent/tool'])
    run('base-entry-overlay',['python3','stage3/meter/entry_overlay.py',str(work/'base-overlay'),str(work/'base-entry-overlay')])
    run('base-entry-build',['go','build','-buildvcs=false','-overlay='+str(work/'base-entry-overlay/overlay.json'),'-o',str(work/'base-entry'),'./stage3/census/latent/tool'])
    run('base-directory',[str(work/'base-census'),'/workspace/scratch/abrq61p-adapted/src/compiler',str(work/'base.jsonl')],dict(LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(work/'base-times.jsonl')))
    run('base-entry',[str(work/'base-entry'),'/workspace/scratch/abrq61p-adapted/src/tsc/tsc.ts',str(work/'base-entry.jsonl')],dict(LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(work/'base-entry-times.jsonl')))

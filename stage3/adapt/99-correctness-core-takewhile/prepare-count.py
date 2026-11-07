import hashlib,json,os,shutil,signal,subprocess,time
from pathlib import Path
repo=Path('/workspace/adamic');root=Path('/workspace/scratch/takewhile');raw=root/'raw';target=root/'counted';out=repo/'stage3/adapt/99-correctness-core-takewhile'
shutil.copytree(root/'before',target,symlinks=True,ignore=shutil.ignore_patterns('node_modules'))
(target/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),NODE_OPTIONS='--require='+str(out/'count-preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1')
for name,args in [('instrument',['node',str(out/'instrument.cjs'),str(target)]),('count-build',['npm','run','build'])]:
    begin=time.monotonic()
    with (raw/(name+'.log')).open('w') as log:
        child=subprocess.Popen(args,cwd=target,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
        try: code=child.wait(timeout=1400)
        finally:
            try: os.killpg(child.pid,signal.SIGKILL)
            except ProcessLookupError: pass
            child.wait()
    assert code==0,(name,code)
    print(name,code,round(time.monotonic()-begin,2),flush=True)
(target/'node_modules').unlink()
manifest={str(p.relative_to(target)):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((target/base).rglob('*')) if p.is_file()}
(raw/'counted-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')

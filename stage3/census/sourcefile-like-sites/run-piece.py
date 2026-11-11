import os, sys, json, pathlib, subprocess, time, signal, gzip
out=pathlib.Path(__file__).resolve().parent
roots=json.loads((out/'roots.json').read_text())
piece=int(sys.argv[1]); selected=roots[(piece-1)*10:piece*10]
assert selected
root=pathlib.Path('/workspace/scratch/baobab-adapted/src/compiler')
def measure(name):
    stem=name.replace('/','__').removesuffix('.ts'); raw=out/'raw'/stem
    def artifact(suffix): return pathlib.Path(str(raw)+suffix)
    status=artifact('.status.json')
    assert not status.exists(), str(status)
    command=['/workspace/scratch/baobab-latent',str(root),str(artifact('.jsonl'))]
    env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(root/name),LATENT_TIMES=str(artifact('.times.jsonl')))
    for key in ['LATENT_SPECULATIVE','LATENT_RUN_ONLY','LATENT_RETRY_FILES']: env.pop(key,None)
    start=time.monotonic(); timed_out=False
    with artifact('.log').open('wb') as log:
        child=subprocess.Popen(command,stdout=log,stderr=subprocess.STDOUT,env=env,start_new_session=True)
        try: child.wait(timeout=300)
        except subprocess.TimeoutExpired:
            timed_out=True; os.killpg(child.pid,signal.SIGTERM)
            try: child.wait(timeout=10)
            except subprocess.TimeoutExpired: os.killpg(child.pid,signal.SIGKILL); child.wait()
    record=dict(file=name,command=command,child_exit=child.returncode,wall_seconds=time.monotonic()-start,over_five_minutes=timed_out,piece=piece,max_concurrent_files=(1 if piece==1 else 2))
    status.write_text(json.dumps(record,indent=2)+'\n')
    for suffix in ['.jsonl','.times.jsonl','.log']:
        path=artifact(suffix)
        if path.exists() and path.stat().st_size>1000000:
            with path.open('rb') as src,gzip.open(str(path)+'.gz','wb') as dest:
                import shutil; shutil.copyfileobj(src,dest)
            path.unlink()
    print(json.dumps(record),flush=True)

if piece==1:
    for name in selected: measure(name)
else:
    from concurrent.futures import ThreadPoolExecutor
    with ThreadPoolExecutor(max_workers=2) as pool: list(pool.map(measure,selected))

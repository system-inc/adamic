"""Run one disjoint full-mode file group, committing each side on completion.

Usage: run-piece.py NAME FILE... (paths relative to adapted/src/compiler)
The complete compiler program is loaded on each side; selection only limits
measured files. Piece records are compared only to their matching selection.
"""
from pathlib import Path
import gzip, hashlib, json, os, signal, subprocess, sys, time
repo=Path('/workspace/adamic')
root=Path('/workspace/scratch/20pb346-2')
evidence=Path(__file__).resolve().parent
name,*files=sys.argv[1:]
assert files and len(files)==len(set(files))
source=root/'adapted/src/compiler'
selected=[str((source/f).resolve()) for f in files]
assert all(Path(f).is_file() and Path(f).is_relative_to(source) for f in selected)
out=evidence/name
out.mkdir()
(out/'selection.json').write_text(json.dumps(dict(mode='full',source=str(source),files=files,source_sha256={f:hashlib.sha256((source/f).read_bytes()).hexdigest() for f in files},compiler_base=subprocess.check_output(['git','rev-parse','origin/main'],cwd=repo,text=True).strip(),budget_seconds=1500,GOMEMLIMIT='8GiB'),indent=2)+'\n')
raw={}
for side in ('before','after'):
    work=root/'piece-runs'/name/side
    work.mkdir(parents=True)
    binary=root/('pieces-'+side)
    env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
    env.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE='\n'.join(selected),LATENT_TIMES=str(work/'latent-times.jsonl'),GOMEMLIMIT='8GiB')
    start=time.monotonic()
    with (work/'output.txt').open('w') as log:
        p=subprocess.Popen([str(binary),str(source),str(work/'latent.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
        try:
            code=p.wait(timeout=1500)
        except BaseException:
            os.killpg(p.pid,signal.SIGTERM)
            try:p.wait(timeout=10)
            except subprocess.TimeoutExpired:
                os.killpg(p.pid,signal.SIGKILL);p.wait()
            raise
    wall=time.monotonic()-start
    data=(work/'latent.jsonl').read_bytes()
    rows=[json.loads(line) for line in data.splitlines()]
    result=dict(side=side,exit_code=code,wall_seconds=wall,bytes=len(data),sha256=hashlib.sha256(data).hexdigest(),binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),files=[r['file'] for r in rows[1:]],units=sum(len(r['units']) for r in rows[1:]))
    assert code==0,result
    assert set(result['files'])==set(selected),result
    for filename in ('latent.jsonl','latent-times.jsonl','output.txt'):
        with (out/(side+'-'+filename+'.gz')).open('wb') as target:
            with gzip.GzipFile(fileobj=target,mode='wb',mtime=0) as writer:writer.write((work/filename).read_bytes())
    (out/(side+'-result.json')).write_text(json.dumps(result,indent=2)+'\n')
    raw[side]=data
    if side=='after':
        assert raw['before']==raw['after'],'piece byte identity failed'
        (out/'identity.json').write_text(json.dumps(dict(identical=True,bytes=len(data),sha256=result['sha256'],files=len(files),units=result['units']),indent=2)+'\n')
    subprocess.run(['git','add',str(evidence)],cwd=repo,check=True)
    subprocess.run(['git','commit','-m','meter: save '+name+' '+side+' piece evidence'],cwd=repo,check=True,stdout=subprocess.DEVNULL)
    print(json.dumps(result),flush=True)

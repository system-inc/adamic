"""Sequential foreground pieces, committing and pushing each completed child before continuing."""
import argparse
import datetime
import gzip
import json
import os
from pathlib import Path
import resource
import shutil
import signal
import subprocess
import sys
import time
from pieces_common import OUT,E,S,TREE,digest,dump,load,norm,validate_piece,verify_partition

parser=argparse.ArgumentParser()
parser.add_argument('--piece')
parser.add_argument('--label',choices=['base','head'])
args=parser.parse_args()
repo=OUT.parents[2]
plan=json.loads((E/'plan.json').read_text())
builds=json.loads((E/'builds.json').read_text())
with gzip.open(OUT/'evidence/base-latent.jsonl.gz','rt') as f:whole=[json.loads(x) for x in f]
header=whole[0]
all_units={u['where'] for row in whole[1:] for u in row['units']}
verify_partition(plan['pieces'],all_units)
raw=S/'piece-checkpoints'
raw.mkdir(exist_ok=True)
def git(where,*arguments):
    return subprocess.check_output(['git','-C',str(where),*arguments],stderr=subprocess.STDOUT,text=True)
if not (raw/'.git').exists():
    git(raw,'init','-b','measurement-checkpoints')
    for field in ('user.name','user.email'):
        git(raw,'config',field,git(repo,'config',field).strip())
    shutil.copyfile(E/'plan.json',raw/'plan.json')
    git(raw,'add','plan.json'); git(raw,'commit','-m','Pin first-error piece plan')
manifest=json.loads((OUT/'evidence/source-manifest.json').read_text())
assert manifest=={str(p.relative_to(TREE)):digest(p) for p in sorted((TREE/'src').rglob('*')) if p.is_file()}
assert git(repo,'branch','--show-current').strip()=='typescript/any-members-census'
now=lambda:datetime.datetime.now(datetime.timezone.utc).isoformat()
publication=S/'piece-publication.log'

def publish(label,piece,work,metrics):
    target=E/label/piece['id']; target.mkdir(parents=True,exist_ok=True)
    for name in ('latent.jsonl','times.jsonl','census.log','progress.log'):
        path=target/(name+'.gz')
        if not path.exists():
            with gzip.open(path,'wb') as f:f.write((work/name).read_bytes())
        with gzip.open(path,'rb') as f:assert f.read()==(work/name).read_bytes()
    dump(target/'metrics.json',metrics)
    checkpoint=git(raw,'log','-1','--format=%H','--',str(work.relative_to(raw))).strip()
    assert checkpoint
    checkpoint_file=target/'checkpoint.json'
    if checkpoint_file.exists():
        saved=json.loads(checkpoint_file.read_text())
        assert saved['latent_sha256']==digest(work/'latent.jsonl') and saved['times_sha256']==digest(work/'times.jsonl')
    else:
        dump(checkpoint_file,dict(raw_git_commit=checkpoint,latent_sha256=digest(work/'latent.jsonl'),times_sha256=digest(work/'times.jsonl')))
    for name in ('started.json','memory-before.txt','memory-after.txt'):
        if (work/name).exists() and not (target/name).exists():shutil.copyfile(work/name,target/name)
    index={side:{p.parent.name:json.loads(p.read_text()) for p in sorted((E/side).glob('*/metrics.json'))} for side in ('base','head')}
    dump(E/'INDEX.json',index)
    changed=git(repo,'status','--porcelain','--',str(target),str(E/'INDEX.json')).strip()
    if changed:
        with publication.open('a') as log:
            subprocess.run(['git','-C',str(repo),'add','-f','--',str(target),str(E/'INDEX.json')],check=True,stdout=log,stderr=subprocess.STDOUT)
            subprocess.run(['git','-C',str(repo),'diff','--cached','--check'],check=True,stdout=log,stderr=subprocess.STDOUT)
            staged=git(repo,'diff','--cached','--name-only').splitlines()
            assert staged and all(p.startswith('stage3/census/any-members/') for p in staged)
            subprocess.run(['git','-C',str(repo),'commit','-m',f'stage3/census: checkpoint {label} first-error {piece["id"]}'],check=True,stdout=log,stderr=subprocess.STDOUT)
    with publication.open('a') as log:
        subprocess.run(['git','-C',str(repo),'push','origin','HEAD:refs/heads/typescript/any-members-census'],check=True,stdout=log,stderr=subprocess.STDOUT)
    print(now(), 'PUSHED',label,piece['id'],git(repo,'rev-parse','HEAD').strip(),flush=True)

def run(label,piece):
    binary=S/f'piece-{label}-bin/latent-census'
    assert digest(binary)==builds[label]['binary_sha256']
    work=raw/label/piece['id']
    metrics_path=work/'metrics.json'
    archive=E/label/piece['id']
    if not metrics_path.exists() and (archive/'metrics.json').exists():
        work.mkdir(parents=True,exist_ok=True)
        for name in ('latent.jsonl','times.jsonl','census.log','progress.log'):
            with gzip.open(archive/(name+'.gz'),'rb') as f:(work/name).write_bytes(f.read())
        shutil.copyfile(archive/'metrics.json',metrics_path)
        for name in ('started.json','memory-before.txt','memory-after.txt'):
            if (archive/name).exists():shutil.copyfile(archive/name,work/name)
    identity=dict(schema=1,label=label,piece=piece['id'],plan_sha256=digest(E/'plan.json'),selection_sha256=digest(E/f'{piece["id"]}.selection.json'),binary_sha256=digest(binary),source_manifest_sha256=digest(OUT/'evidence/source-manifest.json'))
    if metrics_path.exists() and json.loads(metrics_path.read_text()).get('returncode')==0:
        metrics=json.loads(metrics_path.read_text())
        assert metrics['identity']==identity
        assert metrics['latent_sha256']==digest(work/'latent.jsonl') and metrics['times_sha256']==digest(work/'times.jsonl')
        validate_piece(piece,load(work/'latent.jsonl'),load(work/'times.jsonl'),header)
        # Recover a completed observation that survived before either Git commit.
        if git(raw,'status','--porcelain','--',str(work.relative_to(raw))).strip():
            git(raw,'add','-f','--',str(work.relative_to(raw)));git(raw,'commit','-m',f'Complete {label} {piece["id"]}')
        print(now(),'RESUME',label,piece['id'],flush=True)
        publish(label,piece,work,metrics);return
    if work.exists():
        work.rename(work.with_name(work.name+'-interrupted-'+str(int(time.time()))))
    work.mkdir(parents=True)
    env=dict(os.environ,LATENT_FULL='0',LATENT_SPECULATIVE='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(work/'times.jsonl'),ANY_MEMBERS_CENSUS_PIECE=str(E/f'{piece["id"]}.selection.json'),ANY_MEMBERS_CENSUS_ROOT=str(TREE),TMPDIR=str(S/'tmp'))
    for key in list(env):
        if key.startswith('LATENT_MUTANT_') or key in ('LATENT_PIECE_PLAN','LATENT_PIECE_ID','LATENT_ONLY_FILE'):env.pop(key)
    start=time.monotonic();cpu=resource.getrusage(resource.RUSAGE_CHILDREN)
    cpu_before=cpu.ru_utime+cpu.ru_stime
    metadata=dict(identity=identity,start_utc=now(),expected_units=len(piece['expected_units']),expected_timed_units=len(piece['expected_timed_units']),estimated_seconds=piece['estimated_seconds'])
    dump(work/'started.json',metadata)
    (work/'memory-before.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
    progress=(work/'progress.log').open('w',buffering=1)
    def say(message):
        line=now()+' '+label+' '+piece['id']+' '+message
        progress.write(line+'\n');print(line,flush=True)
    say('START files='+','.join(f['file'] for f in piece['files']))
    failure=None
    with (work/'census.log').open('w') as log:
        child=subprocess.Popen([str(binary),str(TREE/'src/compiler'),str(work/'latent.pending.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT)
        while True:
            try:code=child.wait(timeout=20);break
            except subprocess.TimeoutExpired:
                timer=work/'times.jsonl'
                lines=timer.read_text().splitlines() if timer.exists() else []
                records=[]
                for line in lines:
                    try:records.append(json.loads(line))
                    except json.JSONDecodeError:break
                age=time.time()-timer.stat().st_mtime if timer.exists() else time.monotonic()-start
                last=norm(records[-1]['where']) if records else 'prepass'
                say(f'completed={len(records)}/{len(piece["expected_timed_units"])} last={last} age={age:.1f}s elapsed={time.monotonic()-start:.1f}s')
                if age>300 or time.monotonic()-start>plan['timeout_seconds']:
                    failure='no timer progress for 300s' if age>300 else 'piece exceeded 1200s budget'
                    say('STOP '+failure+'; capture SIGQUIT goroutine dump before any retry')
                    child.send_signal(signal.SIGQUIT)
                    try:code=child.wait(timeout=60)
                    except subprocess.TimeoutExpired:child.kill();code=child.wait()
                    break
    cpu=resource.getrusage(resource.RUSAGE_CHILDREN)
    metadata.update(returncode=code,end_utc=now(),process_wall_seconds=time.monotonic()-start,process_cpu_seconds=cpu.ru_utime+cpu.ru_stime-cpu_before,failure=failure)
    (work/'memory-after.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
    say(f'EXIT returncode={code}')
    progress.close()
    if code!=0:
        dump(work/'failed.json',metadata)
        raise RuntimeError(f'{label} {piece["id"]} failed: {failure}; see {work}/census.log')
    rows,timers=validate_piece(piece,load(work/'latent.pending.jsonl'),load(work/'times.jsonl'),header)
    (work/'latent.pending.jsonl').replace(work/'latent.jsonl')
    metadata.update(latent_sha256=digest(work/'latent.jsonl'),times_sha256=digest(work/'times.jsonl'),unit_wall_seconds=sum(t['wall_seconds'] for t in timers),unit_cpu_seconds=sum(t['cpu_seconds'] for t in timers))
    dump(metrics_path,metadata)
    # Persist raw latent.jsonl to a local scratch Git repository immediately.
    git(raw,'add','-f','--',str(work.relative_to(raw)))
    git(raw,'commit','-m',f'Complete {label} {piece["id"]}')
    print(now(),'SCRATCH_COMMITTED',label,piece['id'],git(raw,'rev-parse','HEAD').strip(),flush=True)
    publish(label,piece,work,metadata)

selected=[p for p in plan['pieces'] if not args.piece or p['id']==args.piece]
assert selected,'unknown piece'
for piece in selected:
    for label in ([args.label] if args.label else ['base','head']):run(label,piece)
print(now(),'ALL REQUESTED PIECES SAVED AND PUSHED',flush=True)

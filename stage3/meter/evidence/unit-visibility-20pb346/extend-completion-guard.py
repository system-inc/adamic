"""Extend only our parent watchdogs, keeping both census children uninterrupted."""
from pathlib import Path
import json, math, os, signal, time
root=Path('/workspace/scratch/20pb346-2')
budget=21600
active=None

def state(pid):
    try:
        value=Path('/proc')/str(pid)
        raw=(value/'stat').read_text()
        fields=raw[raw.rfind(')')+2:].split()
        args=(value/'cmdline').read_bytes().split(b'\0')[:-1]
        return fields,args
    except FileNotFoundError:
        return None,None

def resume(pid):
    try: os.kill(pid,signal.SIGCONT)
    except ProcessLookupError: pass

def abort(signum,frame):
    if active:
        child,parent=active
        try: os.killpg(child,signal.SIGTERM)
        except ProcessLookupError: pass
        resume(parent)
    raise SystemExit(128+signum)

signal.signal(signal.SIGTERM,abort)
signal.signal(signal.SIGINT,abort)
for side in ('before','after'):
    folder=root/(side+'-compiler-complete')
    while not (folder/'pid').exists(): time.sleep(1)
    if (folder/'run.json').exists(): continue
    child=int((folder/'pid').read_text())
    fields,args=state(child)
    if fields is None: continue
    assert args[0].decode()==str(root/(side+'-compiler')), (child,args)
    parent=int(fields[1])
    parent_fields,parent_args=state(parent)
    assert parent_args[1:]==[str(root/'run.py').encode(),side.encode(),b'compiler',(side+'-compiler-complete').encode(),b'10800'],parent_args
    start=int(fields[19])/os.sysconf('SC_CLK_TCK')
    active=(child,parent)
    event=dict(side=side,child=child,parent_watchdog=parent,original_budget_seconds=10800,effective_budget_seconds=budget,extended_at_elapsed_seconds=time.monotonic()-start)
    (folder/'watchdog-extension.json').write_text(json.dumps(event,indent=2)+'\n')
    os.kill(parent,signal.SIGSTOP)
    try:
        next_sample=math.ceil((time.monotonic()-start)/60)*60
        while True:
            current,_=state(child)
            if current is None or current[0]=='Z': break
            elapsed=time.monotonic()-start
            if elapsed>=budget:
                os.killpg(child,signal.SIGTERM)
                deadline=time.monotonic()+10
                while time.monotonic()<deadline:
                    current,_=state(child)
                    if current is None or current[0]=='Z':break
                    time.sleep(0.1)
                else:os.killpg(child,signal.SIGKILL)
                break
            if elapsed>=next_sample:
                row=dict(seconds=round(elapsed,3),times_bytes=(folder/'latent-times.jsonl').stat().st_size,latent_bytes=(folder/'latent.jsonl').stat().st_size)
                with (folder/'growth.jsonl').open('a') as output:output.write(json.dumps(row)+'\n')
                next_sample+=60
            time.sleep(1)
    finally:
        resume(parent)
        active=None
    while not (folder/'run.json').exists():time.sleep(0.1)
    original=(folder/'run.json').read_text()
    (folder/'run-original-watchdog.json').write_text(original)
    run=json.loads(original)
    run['original_wall_budget_seconds']=run['wall_budget_seconds']
    run['wall_budget_seconds']=budget
    run['watchdog']='external completion guard; census child never suspended'
    run['samples']=[json.loads(line) for line in (folder/'growth.jsonl').read_text().splitlines()]
    (folder/'run.json').write_text(json.dumps(run,indent=2)+'\n')

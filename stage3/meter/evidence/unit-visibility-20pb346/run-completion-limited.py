"""Continue an orphaned census and supervise remaining runs across exec resets."""
from pathlib import Path
import json, os, select, signal, subprocess, time
root=Path('/workspace/scratch/20pb346-2')
budget=21600
active=None

def proc(pid):
    folder=Path('/proc')/str(pid)
    try:
        text=(folder/'stat').read_text()
        fields=text[text.rfind(')')+2:].split()
        args=(folder/'cmdline').read_bytes().split(b'\0')[:-1]
        return fields,args
    except FileNotFoundError:return None,None

def stop(signum,frame):
    if active:
        try:os.killpg(active,signal.SIGTERM)
        except ProcessLookupError:pass
    raise SystemExit(128+signum)

signal.signal(signal.SIGTERM,stop)
signal.signal(signal.SIGINT,stop)
for side in ('before','after'):
    folder=root/(side+'-compiler-complete')
    child_object=None
    log=None
    folder.mkdir()
    log=(folder/'latent.log').open('w')
    env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(folder/'latent-times.jsonl'),GOMEMLIMIT='8GiB')
    child_object=subprocess.Popen([str(root/(side+'-compiler')),str(root/'adapted/src/compiler'),str(folder/'latent.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
    pid=child_object.pid
    (folder/'pid').write_text(str(pid)+'\n')
    fields,args=proc(pid)
    active=pid
    start=int(fields[19])/os.sysconf('SC_CLK_TCK')
    event=dict(side=side,pid=pid,budget_seconds=budget,monitor_resumed_at_seconds=time.monotonic()-start,note='controlled retry after default-memory baseline SIGKILL',GOMEMLIMIT='8GiB')
    (folder/'supervision.json').write_text(json.dumps(event,indent=2)+'\n')
    poller=select.poll()
    pidfd=os.pidfd_open(pid)
    poller.register(pidfd,select.POLLIN)
    next_sample=(int((time.monotonic()-start)/60)+1)*60
    try:
        while True:
            remaining=budget-(time.monotonic()-start)
            if remaining<=0:
                os.killpg(pid,signal.SIGTERM)
                if not poller.poll(10000):os.killpg(pid,signal.SIGKILL)
                poller.poll(10000)
                break
            if poller.poll(min(1000,max(1,int(remaining*1000)))):break
            elapsed=time.monotonic()-start
            if elapsed>=next_sample:
                row=dict(seconds=round(elapsed,3),times_bytes=(folder/'latent-times.jsonl').stat().st_size,latent_bytes=(folder/'latent.jsonl').stat().st_size)
                with (folder/'growth.jsonl').open('a') as output:output.write(json.dumps(row)+'\n')
                next_sample+=60
        elapsed=time.monotonic()-start
        final,_=proc(pid)
        assert final is not None and final[0]=='Z', 'cannot read authoritative kernel exit status'
        raw_status=int(final[49])
        code=os.waitstatus_to_exitcode(raw_status)
        if child_object is not None:assert child_object.wait()==code
        samples=[json.loads(line) for line in (folder/'growth.jsonl').read_text().splitlines()]
        result=dict(side=side,mode='compiler',wall_budget_seconds=budget,wall_seconds=elapsed,exit_code=code,kernel_wait_status=raw_status,wall_clock_method='CLOCK_MONOTONIC minus /proc PID stat start ticks, resolution 1/CLK_TCK',supervision='detached pidfd monitor',GOMEMLIMIT='8GiB',samples=samples)
        (folder/'run.json').write_text(json.dumps(result,indent=2)+'\n')
    finally:
        os.close(pidfd)
        if log is not None:log.close()
        active=None
(root/'completion-supervisor.done').write_text('both full passes exited\n')

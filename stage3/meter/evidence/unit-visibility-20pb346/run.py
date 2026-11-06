from pathlib import Path
import json, os, signal, subprocess, sys, time
root=Path('/workspace/scratch/20pb346-2')
side, mode, label, budget = sys.argv[1:5]
budget=int(budget)
folder=root/label
folder.mkdir()
source=Path(sys.argv[5]) if len(sys.argv)>5 else root/'adapted/src'/('tsc/tsc.ts' if mode=='entry' else 'compiler')
env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(folder/'latent-times.jsonl'))
if mode=='entry': env['LATENT_CPU_PROFILE']=str(folder/'cpu.pprof')
process=None

def cleanup(*args):
    if process is not None and process.poll() is None:
        os.killpg(process.pid,signal.SIGTERM)
        try: process.wait(timeout=10)
        except subprocess.TimeoutExpired:
            os.killpg(process.pid,signal.SIGKILL)
            process.wait()
    if args: raise SystemExit(128+args[0])

signal.signal(signal.SIGTERM,cleanup)
signal.signal(signal.SIGINT,cleanup)
start=time.monotonic()
samples=[]
try:
    with (folder/'latent.log').open('w') as log:
        process=subprocess.Popen([str(root/(side+'-'+mode)),str(source),str(folder/'latent.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
        (folder/'pid').write_text(str(process.pid)+'\n')
        for second in range(60,budget+1,60):
            wait=max(0,start+second-time.monotonic())
            # Event.wait keeps signals responsive; each wait is at most 60 s.
            try: process.wait(timeout=wait)
            except subprocess.TimeoutExpired: pass
            if process.poll() is not None: break
            sample={'seconds':second,'times_bytes':(folder/'latent-times.jsonl').stat().st_size if (folder/'latent-times.jsonl').exists() else 0,'latent_bytes':(folder/'latent.jsonl').stat().st_size if (folder/'latent.jsonl').exists() else 0}
            samples.append(sample)
            with (folder/'growth.jsonl').open('a') as output: output.write(json.dumps(sample)+'\n')
        cleanup()
finally:
    cleanup()
    result=dict(side=side,mode=mode,wall_budget_seconds=budget,wall_seconds=time.monotonic()-start,exit_code=process.returncode if process is not None else None,samples=samples)
    (folder/'run.json').write_text(json.dumps(result,indent=2)+'\n')

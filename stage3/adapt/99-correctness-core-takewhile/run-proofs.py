import json,os,signal,subprocess,time
from pathlib import Path
repo=Path('/workspace/adamic');root=Path('/workspace/scratch/takewhile');raw=root/'raw';out=repo/'stage3/adapt/99-correctness-core-takewhile';begin=time.monotonic();ready=raw/'prepare-proof-driver.log'
while not ready.exists() or not ready.read_text().rstrip().endswith('READY'):
    if time.monotonic()-begin>1400:raise SystemExit('Proof preparation did not complete within bound')
    time.sleep(2)
commands={
    'types-census':['python3','-c',"import subprocess; p='"+str(out)+"/'; subprocess.run(['python3',p+'check-types.py','after'],check=True); subprocess.run(['python3',p+'census.py','after','mutant'],check=True); subprocess.run(['python3',p+'analyze-census.py'],check=True)"],
    'replay':['python3',str(out/'replay.py')],
    'after-lane':['python3',str(out/'after-lane.py')],
}
children={};logs={};started={};results={}
for name,command in commands.items():
    logs[name]=(raw/(name+'-driver.log')).open('w');children[name]=subprocess.Popen(command,cwd=repo,stdout=logs[name],stderr=subprocess.STDOUT,start_new_session=True);started[name]=time.monotonic();print('START',name,children[name].pid,flush=True)
def killtree(pid):
    rows=[r.split(None,2) for r in subprocess.check_output(['ps','-eo','pid,ppid,comm'],text=True).splitlines()[1:]];ids={pid}
    while True:
        more=ids|{int(p) for p,parent,_ in rows if int(parent) in ids}
        if more==ids:break
        ids=more
    for number in sorted(ids,reverse=True):
        try:os.kill(number,signal.SIGKILL)
        except ProcessLookupError:pass
try:
    while len(results)<len(children):
        for name,child in children.items():
            if name in results:continue
            code=child.poll()
            if code is None and time.monotonic()-started[name]>1400:killtree(child.pid);code=child.wait()
            if code is not None:
                results[name]={'exit':code,'seconds':time.monotonic()-started[name]};logs[name].close();(raw/'proof-jobs.json').write_text(json.dumps(results,indent=2)+'\n');print(name,json.dumps(results[name]),flush=True)
        time.sleep(1)
finally:
    for name,child in children.items():
        killtree(child.pid);child.wait()
        if not logs[name].closed:logs[name].close()
assert all(r['exit']==0 for r in results.values()),results
print('ALL PROOF JOBS PASS',flush=True)

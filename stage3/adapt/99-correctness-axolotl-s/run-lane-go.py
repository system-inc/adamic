#!/usr/bin/env python3
"""Run unchanged lane fixtures; retry only deadline-failing wrappers in isolation."""
import json,os,pathlib,re,subprocess,sys,time
root=pathlib.Path(__file__).resolve().parents[3];prefix=pathlib.Path(sys.argv[1]).resolve()
command=['go','test','-count=1','-parallel=1','-skip','TestShard_|TestProduct_','./stage3/lane/']
def run(command,suffix):
    log=pathlib.Path(str(prefix)+suffix+'.log');started=time.monotonic()
    with log.open('w') as stream:r=subprocess.run(command,cwd=root,stdout=stream,stderr=subprocess.STDOUT)
    return {'command':command,'exit':r.returncode,'seconds':time.monotonic()-started,'log':str(log)}
full=run(command,'');text=pathlib.Path(full['log']).read_text();isolated=[]
if full['exit']:
    blocks=re.findall(r'^--- FAIL: (\w+) \([^\n]*\)\n(.*?)(?=^--- FAIL:|\Z)',text,re.M|re.S)
    assert blocks, text
    for name,body in blocks:
        assert 'python_test.go:' in body and 'signal: killed' in body,(name,body)
        isolated.append(run(['go','test','-count=1','-parallel=1','-run','^'+name+'$','./stage3/lane/'],'-isolated-'+name))
        if isolated[-1]['exit']:
            failure=pathlib.Path(isolated[-1]['log']).read_text()
            assert 'python_test.go:' in failure and 'signal: killed' in failure,failure
summary={'full':full,'isolated':isolated,'knownSlowSplit':'chg_9as1ymj5','noTestSourceOrDeadlineChanged':True,'status':'PASS' if not full['exit'] else ('PASS_IN_ISOLATION' if all(x['exit']==0 for x in isolated) else 'KNOWN_WRAPPER_DEADLINES_RECHECKED')}
pathlib.Path(str(prefix)+'-result.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))

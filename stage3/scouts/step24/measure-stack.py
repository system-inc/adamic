"""Independent process thresholds; native calibration is not parser stack proof."""
from pathlib import Path
import json
import os
import resource
import subprocess
import tempfile

here=Path(__file__).resolve().parent
compiler=os.environ.get('STEP24_COMPILER','/tmp/step24-adamic')
source=here/'fixtures/recursion.a'
node_script="const ts=require(process.env.PARSER_TYPESCRIPT); const n=Number(process.argv[1]); try { ts.createSourceFile('deep.ts','const x='+ '('.repeat(n)+'1'+')'.repeat(n)+';',ts.ScriptTarget.Latest,false,ts.ScriptKind.TS); console.log('parsed'); } catch(e) { console.error(String(e));process.exit(1); }"
def limits():
    resource.setrlimit(resource.RLIMIT_STACK,(8388608,8388608))
    resource.setrlimit(resource.RLIMIT_CORE,(0,0))

def threshold(command):
    trials=[]
    def good(n):
        p=subprocess.run(command+[str(n)],capture_output=True,preexec_fn=limits)
        trials.append({'depth':n,'exit':p.returncode,'stdout':p.stdout.decode(),'stderr':p.stderr.decode()})
        return p.returncode==0
    lo,hi=0,1
    while good(hi):
        lo,hi=hi,hi*2
        assert hi<=1048576,'probe bound exceeded'
    while hi-lo>1:
        mid=(lo+hi)//2
        if good(mid):lo=mid
        else:hi=mid
    return {'last_success':lo,'first_failure':hi,'trials':trials}
with tempfile.TemporaryDirectory(prefix='step24-stack-') as tmp:
    binary=Path(tmp)/'recursion'
    p=subprocess.run([compiler,'build',str(source),'-o',str(binary)],capture_output=True)
    assert p.returncode==0,p.stderr
    native=threshold([str(binary)])
    node=threshold(['node','-e',node_script])
    emitted=subprocess.run([compiler,'c',str(source)],capture_output=True)
    assert emitted.returncode==0,emitted.stderr
    c=Path(tmp)/'recursion.c';c.write_bytes(emitted.stdout)
    object_file=Path(tmp)/'recursion.o'
    p=subprocess.run(['clang','-O2','-fstack-usage','-I',str(here.parents[2]/'internal/native/runtime'),'-c',str(c),'-o',str(object_file)],capture_output=True)
    assert p.returncode==0,p.stderr
    frames=[l for l in object_file.with_suffix('.su').read_text().splitlines() if 'descend' in l]
    assert len(frames)==1,frames
    report={'stack_bytes':8388608,'node':node,'native_calibration':native,'native_frame_usage':frames,
            'limit':'Calibration descends a string-building function. Actual parser C/frame sizes remain unavailable.'}
(here/'evidence/stack.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps({k:v for k,v in report.items() if k not in ['node','native_calibration']},indent=2))
print('Node parentheses',node['last_success'],node['first_failure'])
print('Native calibration',native['last_success'],native['first_failure'])

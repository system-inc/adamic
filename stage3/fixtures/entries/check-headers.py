#!/usr/bin/env python3
"""Run the unchanged fetched Gate.aCheck on these six paths and prove its header predicate fails."""
import argparse
import importlib.util
import json
import subprocess
from pathlib import Path

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[2]
p=argparse.ArgumentParser();p.add_argument('gate',type=Path);p.add_argument('--scratch',type=Path,required=True);a=p.parse_args()
a.scratch.mkdir(parents=True,exist_ok=True)
spec=importlib.util.spec_from_file_location('step12_fastgate',a.gate);module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
paths=[str(f.relative_to(ROOT)) for f in sorted(HERE.glob('*.a'))]
observations={}
class Process:
    def __init__(self,command):self.command=command;self.process=subprocess.Popen(command,cwd=ROOT,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
    def communicate(self):
        out,err=self.process.communicate(timeout=180);self.returncode=self.process.returncode
        observations[self.command[-1]]={'stdout':out,'stderr':err,'exit':self.returncode}
        name=Path(self.command[-1]).stem
        (a.scratch/(name+'.c.stdout')).write_text(out);(a.scratch/(name+'.c.stderr')).write_text(err)
        return out,err
class Harness:
    def __init__(self):self.arguments=type('Args',(),{'tree':str(ROOT)})();self.result={};self.steps={};self.exits={};self.failure=None
    def step(self,name,command):
        with (a.scratch/'acheck-build.log').open('wb') as log:code=subprocess.run(command,cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,timeout=180).returncode
        assert code==0;return True
    def spawn(self,command,stdout,stderr):return Process(command)
    def fail(self,name,detail):self.failure=detail
h=Harness();module.Gate.aCheck(h,paths)
assert h.exits['a-check']==0,h.failure
assert len(h.result['a_check'])==6
(a.scratch/'a-check.json').write_text(json.dumps(h.result['a_check'],indent=2)+'\n')
print('PASS unchanged Gate.aCheck:',len(paths),'accepted')
# Replay actual compiler diagnostics to test only the gate's header predicate.
# The sole input mutation is a first-line comment; bodies and line offsets are unchanged.
class Captured:
    def __init__(self,row):self.row=row;self.returncode=row['exit']
    def communicate(self):return self.row['stdout'],self.row['stderr']
class Replay(Harness):
    def step(self,name,command):return True
    def spawn(self,command,stdout,stderr):return Captured(self.captured)
mutants=[]
for path in paths:
    original=(ROOT/path).read_text();body=original.split('\n',1)[1]
    scratch=a.scratch/Path(path).name
    scratch.write_text('// a-check: refused deliberately unmatched step12 reason\n'+body)
    assert scratch.read_text().split('\n',1)[1]==body
    test=Replay();test.captured=observations[path]
    module.Gate.aCheck(test,[str(scratch)])
    assert test.exits['a-check']==1,'header predicate accepted wrong reason: '+path
    mutants.append({'file':Path(path).name,'edit':'Replace first-line expected header with a deliberately unmatched refusal reason','caught_by':'unchanged Gate.aCheck over captured real compiler diagnostics','failure':test.failure})
(a.scratch/'header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
print('PASS six header mutants caught by unchanged Gate.aCheck predicate')

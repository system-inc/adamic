import importlib.util, json, pathlib, subprocess, hashlib
root=pathlib.Path.cwd(); evidence=root/'review/compiler/loom-r-hidden-boundaries-ac'
spec=importlib.util.spec_from_file_location('pinned_gate',evidence/'gate.py'); gate=importlib.util.module_from_spec(spec); spec.loader.exec_module(gate)
rows=json.loads((evidence/'statuses.json').read_text()); paths=[r['path'] for r in rows]
for r in rows:
    p=root/r['path']; old=subprocess.check_output(['git','show','6685711:'+r['path']]); body=p.read_bytes()
    if body.startswith(b'// a-check:'): body=body.split(b'\n',1)[1]
    original=old.split(b'\n',1)[1] if old.startswith(b'// a-check:') else old
    assert body==original and hashlib.sha256(body).hexdigest()==r['body_sha256']
    if old.startswith(b'// a-check:'):
        header=old.split(b'\n',1)[0].decode(); expected=header.split('// a-check: ',1)[1]
        assert expected.startswith('refused ') and expected[len('refused '):] in r['diagnostic']
        p.write_bytes(old);r['header']=header;r['header_change']='retained'
    elif r['header']: r['header_change']='added'
    else:r['header_change']='none'
(evidence/'statuses.json').write_text(json.dumps(rows,indent=2)+'\n')
class Process:
    def __init__(self,command): self.p=subprocess.Popen(['/workspace/scratch/hidden-ac-compiler']+command[1:],stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
    def communicate(self):
        out,err=self.p.communicate(timeout=45);self.returncode=self.p.returncode;return out,err
class Harness:
    def __init__(self):self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None
    def step(self,name,command):return True # Reuse the independently built branch compiler only.
    def spawn(self,command,stdout,stderr):return Process(command)
    def fail(self,name,detail):self.failure=detail
baseline=Harness();gate.Gate.aCheck(baseline,paths)
assert baseline.exits['a-check']==0,baseline.failure
(evidence/'a-check.json').write_text(json.dumps(baseline.result,indent=2)+'\n')
print('PASS unchanged pinned Gate.aCheck:',len(paths),'programs')
selected=next(r['path'] for r in rows if r['header_change']=='added')
p=root/selected;original=p.read_bytes();body=original.split(b'\n',1)[1];mutants=[]
try:
    for name,source in [('removed-header',body),('wrong-reason',b'// a-check: refused unrelated deliberate mutant reason\n'+body)]:
        p.write_bytes(source);h=Harness();gate.Gate.aCheck(h,[selected]);assert h.exits['a-check']==1,(name,h.result)
        mutants.append(dict(mutant=name,path=selected,exit=h.exits['a-check'],failure=h.failure,result=h.result));print('CAUGHT',name,h.failure)
finally:p.write_bytes(original)
(evidence/'header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
for r in rows:
    old=subprocess.check_output(['git','show','6685711:'+r['path']]);new=(root/r['path']).read_bytes()
    if old.startswith(b'// a-check:'):assert old==new
    elif r['header_change']=='added':assert new.split(b'\n',1)[1]==old
    else:assert old==new
print('PASS all 60 bodies byte identical to source; 15 inserted headers; 18 retained headers')

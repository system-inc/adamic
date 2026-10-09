import importlib.util,json,pathlib,subprocess,collections
root=pathlib.Path.cwd();folder=pathlib.Path('/tmp/speculative-a-check')
spec=importlib.util.spec_from_file_location('fastgate',folder/'gate.py');module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
class Process:
    def __init__(self,process,path):self.process=process;self.path=path
    def communicate(self):
        output,error=self.process.communicate();self.returncode=self.process.returncode
        diagnostics[self.path]={'exit':self.returncode,'stderr':error,'stdout_bytes':len(output.encode())}
        return output,error
class Harness:
    def __init__(self,captured=None):
        self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None;self.captured=captured
    def step(self,name,cmd):
        if self.captured is not None:return True
        with (folder/'build.log').open('w') as log:
            code=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode
        assert code==0,code
        return True
    def spawn(self,cmd,stdout,stderr):
        if self.captured is None:return Process(subprocess.Popen(cmd,cwd=root,stdout=stdout,stderr=stderr,text=True),cmd[-1])
        capture=self.captured[cmd[-1]]
        class Captured:
            returncode=capture['exit']
            def communicate(self):return '',capture['stderr']
        return Captured()
    def fail(self,name,detail):self.failure=detail
paths=sorted(str(p.relative_to(root)) for p in (root/'stage3/census/speculative').rglob('*.a'));diagnostics={};baseline=Harness();module.Gate.aCheck(baseline,paths)
assert baseline.exits['a-check']==0,baseline.failure
results=baseline.result['a_check']
for path,row in results.items():row.update(diagnostics[path])
mutants=[]
for path,row in results.items():
    if not row['expected'].startswith('type error'):continue
    target=root/path;original=target.read_bytes();header,body=original.split(b'\n',1)
    try:
        for name,contents in [('removed',body),('wrong-code',b'// a-check: type error TS999999\n'+body)]:
            target.write_bytes(contents);h=Harness(diagnostics);module.Gate.aCheck(h,[path]);assert h.exits['a-check']==1,(path,name)
            mutants.append({'path':path,'mutant':name,'caught':h.failure})
    finally:target.write_bytes(original)
artifact={'gate_sha':(folder/'gate-sha.txt').read_text().strip(),'paths':results,'files':len(paths),'mismatches':0,'outcomes':dict(collections.Counter(r['outcome'] for r in results.values())),'mutants':mutants,'mutant_method':'unchanged Gate.aCheck header predicate against captured real in-place compiler diagnostics; every header restored'}
(folder/'result.json').write_text(json.dumps(artifact,indent=2)+'\n')
print(json.dumps({k:v for k,v in artifact.items() if k not in ('paths','mutants')},indent=2))
print('PASS: all speculative .a files checked in place; zero mismatches; '+str(len(mutants))+' missing/wrong-code header mutants caught')

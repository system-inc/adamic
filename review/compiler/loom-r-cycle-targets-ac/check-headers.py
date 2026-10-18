import collections, hashlib, importlib.util, json, pathlib, subprocess, sys, time
root=pathlib.Path(__file__).resolve().parents[3]
evidence=pathlib.Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('pinned_fast_gate',evidence/'pinned-gate.py')
module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
paths=subprocess.check_output(['git','diff','--name-only','dbdf977f5b','bc53e520','--','*.a'],cwd=root,text=True).splitlines()
(evidence/'paths.json').write_text(json.dumps(paths,indent=2)+'\n')
class Process:
    def __init__(self,p,path,records):self.p=p;self.path=path;self.records=records
    def communicate(self):
        out,err=self.p.communicate(timeout=90);self.returncode=self.p.returncode
        self.records[self.path]={'exit':self.returncode,'stdout':out,'stderr':err}
        return out,err
class Harness:
    def __init__(self,label,replay=None):
        self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None;self.label=label;self.records={};self.replay=replay
    def step(self,name,cmd):
        if self.replay is not None:return True
        pathlib.Path(cmd[3]).parent.mkdir(parents=True,exist_ok=True)
        with (evidence/(self.label+'-build.log')).open('w') as log:
            p=subprocess.run(['timeout','180',*cmd],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=190)
        assert p.returncode==0,p.returncode
        return True
    def spawn(self,cmd,stdout,stderr):
        if self.replay is not None:
            row=self.replay[cmd[-1]]
            class Recorded:
                returncode=row['exit']
                def communicate(self):return row['stdout'],row['stderr']
            return Recorded()
        return Process(subprocess.Popen(['timeout','90',*cmd],cwd=root,stdout=stdout,stderr=stderr,text=True),cmd[-1],self.records)
    def fail(self,name,detail):self.failure=detail

def run(label,selected,replay=None):
    h=Harness(label,replay);module.Gate.aCheck(h,selected)
    result={'predicate_exit':h.exits['a-check'],'failure':h.failure,'seconds':h.steps['a-check'],'results':h.result['a_check'],'diagnostics':h.records}
    (evidence/(label+'.json')).write_text(json.dumps(result,indent=2)+'\n')
    print(label,'exit',h.exits['a-check'],dict(collections.Counter(x['outcome'] for x in h.result['a_check'].values())),flush=True)
    return result
mode=sys.argv[1]
if mode=='before':
    before=run('before',paths)
    assert before['predicate_exit']==1
    records=[]
    for path in paths:
        data=(root/path).read_bytes();diag=before['diagnostics'][path]
        first=diag['stderr'].splitlines()[0]
        if diag['exit']==0:status,reason='checked',''
        elif "can't lower" in diag['stderr'] and ' yet' in diag['stderr']:status,reason='NotYet',first
        elif 'Adamic 0.1 refuses ' in first:
            status='refused';reason=first.split('Adamic 0.1 refuses ',1)[1].split(';',1)[0].split(', and the write at ',1)[0]
        elif ' error TS' in diag['stderr']:
            import re
            status='type error';reason=re.search(r'error (TS[0-9]+)',diag['stderr']).group(1)
        else:raise RuntimeError(first)
        records.append({'path':path,'status':status,'reason':reason,'body_sha256':hashlib.sha256(data).hexdigest(),'body_bytes':len(data)})
    (evidence/'statuses.json').write_text(json.dumps(records,indent=2)+'\n')
    print(json.dumps(records,indent=2))
elif mode=='after':
    after=run('after',paths);assert after['predicate_exit']==0
    records=json.loads((evidence/'statuses.json').read_text())
    for row in records:
        data=(root/row['path']).read_bytes();header,body=data.split(b'\n',1)
        assert hashlib.sha256(body).hexdigest()==row['body_sha256']
        assert row['reason'].encode() in header
    path=paths[0];original=(root/path).read_bytes();mutants=[]
    for name,body in [('removed-header',original.split(b'\n',1)[1]),('wrong-reason',b'// a-check: refused a deliberately incorrect closure cycle reason\n'+original.split(b'\n',1)[1])]:
        try:
            (root/path).write_bytes(body)
            result=run(name,[path],after['diagnostics'])
            assert result['predicate_exit']==1 and 'expected' in result['failure']
            mutants.append({'mutant':name,'path':path,'caught':True,'predicate_exit':1,'compiler_diagnostics':'captured from the real after-header compiler run, as in the README'})
        finally:(root/path).write_bytes(original)
    (evidence/'header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
    final=run('final',paths,after['diagnostics']);assert final['predicate_exit']==0
else:raise RuntimeError(mode)

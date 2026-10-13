import json, os, pathlib, re, subprocess, time
root=pathlib.Path.cwd(); evidence=root/'review/compiler/loom-hidden-boundaries'
def run(label,command):
    started=time.monotonic()
    with (evidence/(label+'.log')).open('w') as log:
        completed=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT)
    events=[]
    for line in (evidence/(label+'.log')).read_text().splitlines():
        try:
            event=json.loads(line)
            if isinstance(event,dict): events.append(event)
        except ValueError: pass
    result={'label':label,'command':command,'exit':completed.returncode,'seconds':round(time.monotonic()-started,3),'tests':[{'name':e['Test'],'action':e['Action'],'seconds':e.get('Elapsed',0)} for e in events if e.get('Test') and e.get('Action') in ('pass','fail','skip')]}
    (evidence/(label+'.json')).write_text(json.dumps(result,indent=2)+'\n'); print(label,completed.returncode,result['seconds'],flush=True)
    return result
if __name__=='__main__':
    changed=subprocess.check_output(['git','diff','--cached','--name-only'],text=True).splitlines()
    groups={}
    for file in changed:
        if file.endswith('_test.go'):
            package=str(pathlib.Path(file).parent)
            names=re.findall(r'^func (Test\w+)\(t \*testing.T\)',pathlib.Path(file).read_text(),re.M)
            if package=='internal/flow': names=[n for n in names if n.startswith(('TestFlowGetters','TestFlowTraceSnapshots','TestFlowCorpusHidden'))]
            if package=='internal/ir': names=['TestCallTargetReaders']
            groups.setdefault(package,[]).extend(names)
    groups['internal/flow']+=['TestFlowCorpusRemainder','TestFlowCorpusSetupIsShared']
    for package,names in groups.items():
        recorded=evidence/(package.replace('/','-')+'.json')
        if recorded.exists() and json.loads(recorded.read_text()).get('exit')==0: continue
        if package=='internal/ir':
            warmed=run('call-target-export-warm',['timeout','180','go','list','-trimpath','-buildvcs=false','-export','-deps','-test','-json','./internal/native','./internal/lower','./internal/fresh','./internal/flow'])
            if warmed['exit']: raise SystemExit(1)
        result=run(package.replace('/','-'),['timeout','180','go','test','./'+package,'-run','^('+'|'.join(sorted(set(names)))+')$','-count=1','-timeout','90s','-parallel','4','-json'])
        if result['exit']: raise SystemExit(1)
    pattern='(hidden_boundary_|hidden_wave2_method_destructuring|overload_|census_overload_binder|getters_census_)'
    result=run('fixtures',['timeout','180','go','test','./internal/oracle','-run','^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/'+pattern,'-count=1','-timeout','90s','-parallel','4','-json'])
    if result['exit']: raise SystemExit(1)
    result=run('vet',['timeout','180','go','vet']+['./'+p for p in sorted(groups)])
    if result['exit']: raise SystemExit(1)

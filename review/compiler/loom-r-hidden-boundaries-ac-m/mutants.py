import json,pathlib,re,subprocess
from mutant_verify import root,evidence,run
previous=root/'review/compiler/hidden-boundaries-next/getters'
saved=evidence/'mutants-results.json'
results=json.loads(saved.read_text()) if saved.exists() else []
completed={r['label'] for r in results if r.get('caught')}
for record in json.loads((previous/'compiler-mutants.json').read_text()):
    name=record['mutant']
    if 'mutant-'+name in completed: continue
    patch=(previous/'patches'/(name+'.patch')).read_text(); lines=patch.splitlines(keepends=True)
    file=lines[0][4:].strip(); text=(root/file).read_text().splitlines(keepends=True); delta=0; index=2
    while index<len(lines):
        match=re.match(r'@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@',lines[index]); assert match,lines[index]
        start=int(match[1]); count=int(match[2]) if match[2] is not None else 1; index+=1; old=[]; new=[]
        while index<len(lines) and not lines[index].startswith('@@'):
            line=lines[index]; index+=1
            if line.startswith('-'): old.append(line[1:])
            elif line.startswith('+'): new.append(line[1:])
            elif line.startswith(' '): old.append(line[1:]);new.append(line[1:])
        at=(start-1 if count else start)+delta
        if text[at:at+len(old)]!=old:
            matches=[i for i in range(len(text)-len(old)+1) if text[i:i+len(old)]==old]
            assert len(matches)==1,(name,file,start,matches)
            at=matches[0]
        text[at:at+len(old)]=new; delta+=len(new)-len(old)
    copied=evidence/(name+'.go.txt'); copied.write_text(''.join(text)); overlay=evidence/(name+'-overlay.json'); overlay.write_text(json.dumps({'Replace':{str(root/file):str(copied)}}))
    command=record['command'][:]; command[command.index('-overlay')+1]=str(overlay)
    position=command.index('-run')+1; command[position]=command[position].replace('^TestNativeAgreesWithNode$/','^TestNativeAgreesWithNode_[0-9]+$/')
    command+=['-parallel','4']
    result=run('mutant-'+name,command); failed=[t['name'] for t in result['tests'] if t['action']=='fail']; result['caught']=result['exit']==1 and bool(failed)
    results.append(result); (evidence/'mutants-results.json').write_text(json.dumps(results,indent=2)+'\n')
    if not result['caught']: raise SystemExit('uncaught '+name)
for name,file,before,after,test,marker in [
('invoke-field-getters','internal/flow/trace_test.go',"'value' in descriptor ? adamicPrint(descriptor.value, seen) : '<accessor>'","adamicPrint(value[key], seen)",'TestFlowTraceSnapshotsDoNotInvokeGetters','snapshot invoked value getter'),
('invoke-code-getter','internal/flow/trace_test.go',"Object.getOwnPropertyDescriptor(value, 'code')?.value","value.code",'TestFlowTraceSnapshotsDoNotInvokeGetters','snapshot invoked code getter'),
('omit-data-mutation','internal/flow/trace_test.go',"adamicPrint(descriptor.value, seen)","'<data>'",'TestFlowTraceSnapshotsDoNotInvokeGetters','snapshot lost nested data mutation'),
('admit-refused-getters','internal/flow/flow_test.go',' || refusedGettersCensusFixture(path)','', 'TestFlowGettersCensusRefusalClassification','entered the runnable flow corpus')]:
    if 'mutant-'+name in completed: continue
    source=(root/file).read_text(); assert source.count(before)==1
    copied=evidence/(name+'.go.txt'); copied.write_text(source.replace(before,after)); overlay=evidence/(name+'-overlay.json'); overlay.write_text(json.dumps({'Replace':{str(root/file):str(copied)}}))
    result=run('mutant-'+name,['timeout','180','go','test','-overlay',str(overlay),'./internal/flow','-run','^'+test+'$','-count=1','-timeout','90s','-parallel','4','-json']); result['caught']=result['exit']==1 and marker in (evidence/('mutant-'+name+'.log')).read_text()
    results.append(result); (evidence/'mutants-results.json').write_text(json.dumps(results,indent=2)+'\n')
    if not result['caught']: raise SystemExit('uncaught '+name)

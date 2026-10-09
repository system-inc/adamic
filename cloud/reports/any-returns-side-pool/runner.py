import json, os, subprocess, sys
from pathlib import Path
cases = [
('skipcensus','./internal/skipcensus','^TestCensus$'),
('flow','./internal/flow','^TestFlowCorpusUnitsCoverEveryProgram$'),
('fresh','./internal/fresh','^TestFreshCorpusUnitsCoverEveryProgram$'),
('fuzz','./internal/fuzz','^TestFindingReproducesAlone$/^leak_runner_death$'),
('ir','./internal/ir','^TestCallTargetReaders$'),
('tnode','./internal/lower','^TestHiddenTNodeConstraintMutationRemainsNotYet$'),
('representation','./internal/lower','^TestRepresentationClockSourceCheckedTypes$'),
('signal','./internal/native','^TestRuntimeStaticsSignalAndExit$/^handler_buffer_mutant$'),
('uint16','./internal/oracle','^TestNewExpressionUint16RemainsNotYet$'),
('optional','./internal/oracle','^TestOptionalFieldCheckedViewPending$'),
]
for backend in ['Native','WASI']:
    for key, fixture in [('uint16','internal/oracle/testdata/new_expression_uint16.a'),('uint16-notyet','internal/oracle/testdata/new_expression_uint16_notyet.a'),('tnode-mutation','internal/oracle/testdata/notyet/hidden_boundary_generic_tnode_mutation.a'),('binder','stage3/fixtures/taste/17_binder_flow.a')]:
        cases.append((backend.lower()+'-'+key,'./internal/oracle','^Test'+backend+'AgreesWithNode$/'+ '/'.join('^'+part.replace('.','[.]')+'$' for part in fixture.split('/'))))
label, worktree = sys.argv[1:3]
root=Path('/workspace/scratch/any-classify')/label; root.mkdir(exist_ok=True)
env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',ADAMIC_ORACLE_WASI='1',WASI_SYSROOT='/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot',GOTMPDIR='/workspace/scratch/hidden-go-build')
results=[]
for key, package, selection in cases:
    command=['go','test',package,'-run',selection,'-timeout','90s','-count=1','-json']
    with (root/(key+'.jsonl')).open('w') as log:
        result=subprocess.run(command,cwd=worktree,env=env,stdout=log,stderr=subprocess.STDOUT)
    events=[]
    for line in (root/(key+'.jsonl')).read_text().splitlines():
        try: events.append(json.loads(line))
        except json.JSONDecodeError: pass
    tests=[{'test':event['Test'],'action':event['Action']} for event in events if 'Test' in event and event.get('Action') in ['pass','fail','skip']]
    record=dict(key=key,command=command,exit=result.returncode,tests=tests)
    results.append(record)
    (root/'results.json').write_text(json.dumps(results,indent=2)+'\n')
    print(label,key,result.returncode,tests,flush=True)

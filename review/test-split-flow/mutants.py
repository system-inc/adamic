#!/usr/bin/env python3
"""Use Go overlays to plant program failures and verify exact unit ownership."""
import json
import os
from pathlib import Path
import subprocess

root=Path(__file__).resolve().parents[2]
logs=Path('/tmp/test-split-flow/mutants'); logs.mkdir(parents=True,exist_ok=True)
env={**os.environ,'ADAMIC_GATE_UNCACHED':'1','GOMAXPROCS':'4'}
units=json.loads((root/'review/test-split-flow/units.json').read_text())
results=[]

def unit(package,path,family=None):
    return next(r['test'] for r in units if r['package']==package and r['path']==path and (family is None or r['family']==family))

def run(name,relative,change,package,selector,expected,catcher,timeout='10m'):
    original=(root/relative).read_text()
    mutated=change(original)
    assert mutated!=original,name
    source=logs/(name+'.go'); source.write_text(mutated)
    overlay=logs/(name+'-overlay.json')
    overlay.write_text(json.dumps({'Replace':{str(root/relative):str(source)}}))
    command=['go','test','-overlay='+str(overlay),'./internal/'+package,'-json','-run',selector,'-count=1','-parallel=4','-timeout='+timeout]
    log_path=logs/(name+'.jsonl')
    with log_path.open('w') as log:
        process=subprocess.run(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
    output=log_path.read_text(); events=[]
    for line in output.splitlines():
        try:events.append(json.loads(line))
        except json.JSONDecodeError:pass
    failed=sorted({r['Test'] for r in events if r['Action']=='fail' and r.get('Test')})
    passed=[r['Test'] for r in events if r['Action']=='pass' and r.get('Test') and '/' not in r['Test']]
    assert process.returncode==1 and failed==[expected] and catcher in output and '[build failed]' not in output,(name,failed,output[-8000:])
    results.append({'name':name,'command':command,'exit':process.returncode,'failed_units':failed,'passed_units':len(passed),'catcher':catcher,'log':str(log_path)})
    print(name+': only '+expected+' failed; '+str(len(passed))+' other units passed',flush=True)
    (root/'review/test-split-flow/mutants.json').write_text(json.dumps(results,indent=2)+'\n')

flow_owner=unit('flow','testdata/joins.a')
def ghost_read(text):
    needle='\tprogram := lowered(t, path)\n\tfor function := -1;'
    assert text.count(needle)==1
    return text.replace(needle,'''\tprogram := lowered(t, path)
    if path == "testdata/joins.a" {
        copy := *program
        copy.Locals = append(slices.Clone(program.Locals), ir.Local{Name: "planted_undefined", Type: ir.Number, Function: -1})
        copy.Main = append(slices.Clone(program.Main), ir.Evaluate{Value: ir.Read{Local: len(program.Locals), Of: ir.Number}})
        program = &copy
    }
    for function := -1;''')
run('uninitialized-join-read','internal/flow/flow_test.go',ghost_read,'flow','^TestFlow(Program|SingleAssignment|MutationRanges|GraphPaths|Liveness)_',flow_owner,'nothing defines')

fresh_owner=unit('fresh','../../dedication/dedication.a')
def missing_write(text):
    text=text.replace('"github.com/system-inc/adamic/internal/fresh"','"github.com/system-inc/adamic/internal/fresh"\n "github.com/system-inc/adamic/internal/ir"')
    needle='\tfor _, write := range fresh.ProveWrites(lowered) {'
    return text.replace(needle,'''    if path == "../../dedication/dedication.a" {
        lowered.Main = append(lowered.Main, ir.SetProperty{Object: ir.ObjectLiteral{}, Name: "planted", Value: ir.ObjectLiteral{}, Site: 0})
    }
'''+needle)
run('unrecorded-dedication-write','internal/fresh/fresh_test.go',missing_write,'fresh','^TestFreshWrites_',fresh_owner,"a write lowering didn't record")

for package,path,owner in [('flow','testdata/joins.a','TestFlowCorpusUnitsCoverEveryProgram'),('fresh','../flow/testdata/joins.a','TestFreshCorpusUnitsCoverEveryProgram')]:
    def drop(text,path=path):
        needle='\t"'+path+'",\n'; assert text.count(needle)==1
        return text.replace(needle,'')
    run(package+'-missing-program','internal/'+package+'/corpus_units_test.go',drop,package,'^'+owner+'$',owner,'corpus count')

run('missing-liveness-check','internal/flow/corpus_coverage_test.go',lambda s:s.replace('checkGraphPathsProgram, checkLivenessProgram,','checkGraphPathsProgram,'),'flow','^TestFlowCorpusUnitsCoverEveryProgram$','TestFlowCorpusUnitsCoverEveryProgram','corpus count')
run('rebuild-lowering-setup','internal/flow/flow_test.go',lambda s:s.replace('value, _ := loweredPrograms.LoadOrStore(absolute, &loweredProgram{})','value := any(&loweredProgram{})'),'flow','^TestFlowCorpusSetupIsShared$','TestFlowCorpusSetupIsShared','lowered program setup was rebuilt')
run('rebuild-trace-setup','internal/flow/trace_test.go',lambda s:s.replace('value, _ := preparedTraces.LoadOrStore(path, &preparedTrace{})','value := any(&preparedTrace{})'),'flow','^TestFlowCorpusSetupIsShared$','TestFlowCorpusSetupIsShared','trace setup was rebuilt')
for package,owner in [('flow',unit('flow','../../dedication/dedication.a')),('fresh',fresh_owner)]:
    run(package+'-over-budget','internal/'+package+'/corpus_coverage_test.go',lambda s:s.replace('began := time.Now()','began := time.Now().Add(-31 * time.Second)'),package,'^'+owner+'$',owner,'test unit exceeded 30 seconds',timeout='30s')

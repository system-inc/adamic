#!/usr/bin/env python3
"""Prove private bridge facts and release guards without shared source edits."""
import argparse, json, os, pathlib, subprocess
p=argparse.ArgumentParser();p.add_argument('artifacts',type=pathlib.Path);args=p.parse_args()
root=pathlib.Path(__file__).resolve().parents[4];unit=pathlib.Path(__file__).resolve().parent
out=args.artifacts.resolve();out.mkdir(parents=True,exist_ok=True)
records=[]
def run(name,cmd,expected=0):
    with (out/(name+'.stdout')).open('wb') as stdout,(out/(name+'.stderr')).open('wb') as stderr:
        result=subprocess.run([str(x) for x in cmd],cwd=root,stdout=stdout,stderr=stderr)
    records.append(dict(name=name,command=[str(x) for x in cmd],exit=result.returncode))
    (out/'commands.json').write_text(json.dumps(records,indent=2)+'\n')
    if result.returncode!=expected:raise RuntimeError(f'{name}: exit {result.returncode}, expected {expected}')
    return (out/(name+'.stdout')).read_bytes(),(out/(name+'.stderr')).read_bytes()
facts=root/'bridge/tsgo/checker/facts.go';side=out/'facts.go'
source=facts.read_text().replace('switch mode {','switch mode {\ncase "wave07-symbol-context": return p.wave07SymbolContext(c,node,question)\ncase "wave07-control-flow": return p.wave07ControlFlow(node,question)\ncase "wave07-program-modules": return p.wave07ProgramModules(node,question)')
side.write_text(source)
def archive(name,changes):
    replacements={str(facts):str(side)}
    for relative,old,new in changes:
        file=root/relative;text=file.read_text()
        if text.count(old)!=1:raise RuntimeError('nonunique mutant '+name)
        changed=out/(name+'-'+file.name);changed.write_text(text.replace(old,new));replacements[str(file)]=str(changed)
    overlay=out/(name+'.json');overlay.write_text(json.dumps({'Replace':replacements}))
    binary=out/(name+'.a');run(name+'-archive',['go','build','-overlay',overlay,'-buildmode=c-archive','-o',binary,'./bridge/tsgo/archive'])
    return binary
stage0=pathlib.Path('/workspace/wave-07-timer/adamic')
for name,file,old,new,entry,controls in [
 ('symbol-value','bridge/tsgo/checker/wave07_symbol_context.go','out.number(p.symbolID(value))','out.number(p.symbolID(value) * 0)','timer_suite.a',pathlib.Path('/workspace/wave-07-timer')),
 ('flow-edge','bridge/tsgo/checker/wave07_control_flow.go','out.ids(successors)','out.ids(nil)','process_suite.a',pathlib.Path('/workspace/wave-07-process')),
 ('module-target','bridge/tsgo/checker/wave07_program_modules.go','target = next.FileName()','target = ""','streams_suite.a',pathlib.Path('/workspace/wave-07-streams-final')),
]:
    library=archive(name,[(file,old,new)])
    binary=out/name;run(name+'-build',[stage0,'build',unit/entry,'-o',binary,'--tsgo',library])
    changed,stderr=run(name,[binary,controls/'tsconfig.json',controls/'controls.manifest'])
    truth=(controls/'controls-go.stdout').read_bytes()
    if stderr or truth==changed:raise RuntimeError(name+' mutant survived or failed outside byte comparison')
    offset=next((i for i,(a,b) in enumerate(zip(truth,changed)) if a!=b),min(len(truth),len(changed)))
    print(f'{name}: exit 0, empty stderr; production Go bytes catch offset {offset}',flush=True)
config=out/'tsconfig.json';config.write_text('{"compilerOptions":{"target":"ES2022","strict":true},"files":["ambient.d.ts"]}')
(out/'ambient.d.ts').write_text('declare function f():void;')
input=out/'input.a';input.write_text('f();')
probe=out/'released.a';probe.write_text("""import {panic,programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??panic('missing input');const question=args[2]??panic('missing question');
const program=tsgoProgram(args[0]??panic('missing config'),[file]);
if(args[3]==='released'){tsgoRelease(program);}
console.log(tsgoInspect(program,file,0,question==='wave07-symbol-context'?3:4,question==='wave07-symbol-context'?'CallExpression':'SourceFile',question));
if(args[3]!=='released'){tsgoRelease(program);}
""")
normal=pathlib.Path('/workspace/wave-07-streams-final/checker.a')
run('release-build',[stage0,'build',probe,'-o',out/'released','--tsgo',normal])
for question in ['wave07-symbol-context','wave07-control-flow','wave07-program-modules']:
    run(question+'-live',[out/'released',config,input,question,'live'])
    stdout,stderr=run(question+'-released',[out/'released',config,input,question,'released'],70)
    if stdout or stderr!=b'adamic: panic: invalid or released checker handle\n':raise RuntimeError('unexpected release refusal')
    print(question+': live success; released panic 70',flush=True)
mutant=archive('release-registry',[('bridge/tsgo/archive/main.go','delete(programs.live, uint64(handle))','// mutant retains released program')])
run('released-mutant-build',[stage0,'build',probe,'-o',out/'released-mutant','--tsgo',mutant])
run('released-mutant',[out/'released-mutant',config,input,'wave07-symbol-context','released'])
print('release-registry: mutant exits 0, caught by required panic 70',flush=True)
print('PASS question mutants and release checks',flush=True)

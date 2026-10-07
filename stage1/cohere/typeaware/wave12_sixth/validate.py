"""Prepared reporting comparison only; never a native JSX source oracle."""
import gzip
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

repository = Path(__file__).resolve().parents[4]
owned = Path(__file__).resolve().parent
scratch = Path(sys.argv[1]); scratch.mkdir(parents=True, exist_ok=True)
commands = []

def run(name, args, directory=repository, expected=0):
    out = scratch / (name + '.stdout'); err = scratch / (name + '.stderr')
    with out.open('wb') as stdout, err.open('wb') as stderr:
        result = subprocess.run([str(a) for a in args], cwd=directory, stdout=stdout, stderr=stderr)
    commands.append(dict(name=name, args=[str(a) for a in args], code=result.returncode))
    assert result.returncode == expected, (name, result.returncode, err.read_text())
    return out.read_bytes(), err.read_bytes()

cases = [
 ('fragment', 'const x = <React.Fragment />;', -1, ''),
 ('undef', 'const x = <Missing />;', -1, ''),
 ('object', 'function C() { return <Ctx.Provider value={{a:1}}/>; }', 0, ''),
 ('array', 'function C() { return <Ctx.Provider value={[]}/>; }', 1, ''),
 ('function', 'function C() { return <Ctx.Provider value={() => {}}/>; }', 2, ''),
 ('declaration', 'function C() { function v() {} return <Ctx.Provider value={v}/>; }', 3, 'v'),
 ('variable', 'function C() { const v = {a:1}; return <Ctx.Provider value={v}/>; }', 0, 'v'),
 ('class', 'function C() { return <Ctx.Provider value={class {}}/>; }', 4, ''),
 ('new', 'function C() { return <Ctx.Provider value={new Object()}/>; }', 5, ''),
 ('regex', 'function C() { return <Ctx.Provider value={/a/}/>; }', 6, ''),
 ('jsx', 'function C() { return <Ctx.Provider value={<span/>}/>; }', 7, ''),
 ('jsxfragment', 'function C() { return <Ctx.Provider value={<></>}/>; }', 8, ''),
 ('assignment', 'function C() { let v: any; return <Ctx.Provider value={v = {}}/>; }', 9, ''),
 ('memo', 'function C() { const v = useMemo(() => ({a:1})); return <Ctx.Provider value={v}/>; }', -1, ''),
 ('unstable', 'function C() { const dep = {}; const v = useMemo(() => ({a:1}), [dep]); return <Ctx.Provider value={v}/>; }', -1, ''),
 ('korean', 'function C() { const 테스트 = {}; return <Ctx.Provider value={테스트}/>; }', 0, '테스트'),
 ('greek', 'function C() { function Ω() {} return <Ctx.Provider value={Ω}/>; }', 3, 'Ω'),
 ('joiner', 'function C() { const a\u200db = {}; return <Ctx.Provider value={a\u200db}/>; }', 0, 'a\u200db'),
]
paths=[]
for name, source, _, _ in cases:
    path=scratch/(name+'.tsx')
    path.write_text('export {}; declare const React: any; declare const Ctx: any; declare function useMemo<T>(callback: () => T, deps?: any[]): T;\n'+source+'\n')
    paths.append(path)
config=scratch/'tsconfig.json'; config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','jsx':'preserve','noEmit':True},'files':[str(p) for p in paths]}))
manifest=scratch/'manifest';manifest.write_text('\n'.join(map(str,paths))+'\n')
virtual=repository/'cohere/adamic_wave12_sixth_oracle.go'
overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(owned/'oracle.go')}}))
oracle=scratch/'oracle';run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repository/'cohere')
expected,_=run('production-go',[oracle,config,manifest]);(scratch/'expected').write_bytes(expected)
records=expected.decode().splitlines(); generated=[]; case_index=-1; counts={}; missing=[]
for row in records:
    if row.startswith('file\t'):
        case_index+=1; generated.append('console.log('+json.dumps(row)+');');continue
    if row.startswith('findings '):generated.append('console.log('+json.dumps(row)+');');continue
    fields=row.split('\t'); start,end=int(fields[0]),int(fields[1]);rule,id_=fields[2:4]
    name,source,construction,variable=cases[case_index]
    counts[rule]=counts.get(rule,0)+1
    supplied=f'new SuppliedNode({start}, {end})'
    if rule=='react/jsx-fragments':call=f'fragment({supplied}, false)'
    elif rule=='react/jsx-no-undef':call=f'undef({supplied})'
    elif id_=='memoWithoutDependenciesMsg':call=f'memo({supplied}, "useMemo", 2)'
    elif id_=='unstableDependencyMsg':
        # Prepared message arguments for this single known source, not native stability discovery.
        # Go's reason is verified from the production stream, not supplied message text.
        assert 'is a new object built during render' in fields[4], fields[4]
        call=f'unstable({supplied}, "dep", "useMemo", 2, "is a new object built during render", 2)'
    else:
        assert construction>=0, (name,id_)
        call=f'context({supplied}, {construction}, 2, {2 if variable or construction == 9 else 0}, {json.dumps(variable)})'
    generated.append('console.log('+call+'.written());')
assert set(counts)=={'react/jsx-fragments','react/jsx-no-undef','react/jsx-no-constructed-context-values'},counts
imports='''import { SuppliedNode } from './node.a';
import { report as fragment } from './jsx_fragments/messages.a';
import { report as undef } from './jsx_no_undef/messages.a';
import { report as context, memoWithoutDependencies as memo, unstableDependency as unstable } from './jsx_no_constructed_context_values/messages.a';
'''
# Private copies retain relative imports into the unchanged shared Diagnostic.
copy=scratch/'source';copy.mkdir(exist_ok=True)
for file in owned.rglob('*.a'):
    destination=copy/file.relative_to(owned);destination.parent.mkdir(parents=True,exist_ok=True)
    text=file.read_text().replace("'../../diagnostic.ts'",repr(str(repository/'stage1/cohere/typeaware/diagnostic.ts')))
    destination.write_text(text)
entry=copy/'main.a';entry.write_text(imports+'\n'.join(generated)+'\n')
stage0=scratch/'adamic';run('stage0',['go','build','-o',stage0,'./cmd/adamic'])
for sanitize in [False,True]:
    binary=scratch/('native-asan' if sanitize else 'native')
    args=[stage0,'build',entry,'-o',binary]+(['--sanitize'] if sanitize else [])
    run(binary.name+'-build',args)
    got,err=run(binary.name+'-run',[binary]);assert not err;assert got==expected, binary.name
source_node,source_err=run('source-node',['node','--disable-warning=ExperimentalWarning',repository/'oracle/node.mjs',entry]);assert not source_err;assert source_node==expected
emitted,_=run('emitted-build',[stage0,'js',entry])
emitted_file=scratch/'emitted.mjs';emitted_file.write_bytes(emitted)
emitted_node,emitted_err=run('emitted-node',['node','--disable-warning=ExperimentalWarning',repository/'oracle/node.mjs',emitted_file]);assert not emitted_err;assert emitted_node==expected
mutants=[]
for folder,original in [('jsx_fragments',"'jsx-fragments'"),('jsx_no_undef',"'jsx-no-undef'"),('jsx_no_constructed_context_values',"'jsx-no-constructed-context-values'")]:
    file=copy/folder/'messages.a';text=file.read_text();assert original in text
    file.write_text(text.replace(original,original[:-1]+'-mutant\''))
    binary=scratch/(folder+'-mutant');run(binary.name+'-build',[stage0,'build',entry,'-o',binary]);got,err=run(binary.name+'-run',[binary]);assert not err;assert got!=expected
    mutants.append(dict(rule=folder,first_difference=next(i for i,(a,b) in enumerate(zip(got,expected)) if a!=b)))
    file.write_text(text)
quote_file=copy/'jsx_no_constructed_context_values/quote.a'
quote_text=quote_file.read_text();assert quote_text.count("result += character;")==1
quote_file.write_text(quote_text.replace("result += character;", "result += 'X';", 1))
mutated=scratch/'unicode-quote-mutant';run(mutated.name+'-build',[stage0,'build',entry,'-o',mutated]);got,err=run(mutated.name+'-run',[mutated]);assert not err and got!=expected
mutants.append(dict(rule='constructed-context-unicode-quote',first_difference=next(i for i,(a,b) in enumerate(zip(got,expected)) if a!=b)))
quote_file.write_text(quote_text)
guard_results=[]
context_file=copy/'jsx_no_constructed_context_values/messages.a'
original_context=context_file.read_text()
for name,call,before,after in [
    ('construction-range', 'context(new SuppliedNode(0,1), 100, 1, 0, "")', "names[construction]", "names[0]"),
]:
    entry.write_text(imports+'console.log('+call+'.written());\n')
    normal=scratch/(name+'-guard');run(normal.name+'-build',[stage0,'build',entry,'-o',normal])
    got,err=run(normal.name+'-run',[normal],expected=70);assert not got;assert err == b'adamic: panic: unknown prepared construction kind\n'
    assert original_context.count(before)==1
    context_file.write_text(original_context.replace(before,after,1))
    mutated=scratch/(name+'-guard-mutant');run(mutated.name+'-build',[stage0,'build',entry,'-o',mutated]);got,err=run(mutated.name+'-run',[mutated]);assert got and not err
    guard_results.append(dict(guard=name,normal_exit=70,mutant_exit=0))
    context_file.write_text(original_context)
entry.write_text(imports+'\n'.join(generated)+'\n')
from check_kinds import validate_kinds
metadata_mutants=validate_kinds(scratch, run)
parser=scratch/'shared-parser';run('shared-parser-build',[stage0,'build',repository/'stage1/typescript/parser/main.ts','-o',parser])
for path in paths:
    got,err=run('shared-parser-'+path.stem,[parser,path,'--whole']);assert got and not err
    assert b'Jsx' in got, path
# The shared registry's source-Node and emitted-JavaScript paths cannot load the C checker.
probe=owned/'checker_gap.a'
got,err=run('checker-gap-node',['node','--disable-warning=ExperimentalWarning',repository/'oracle/node.mjs',probe],expected=70)
assert not got and b"does not provide an export named 'tsgoProgram'" in err
got,err=run('checker-gap-emitted',[stage0,'js',probe],expected=1)
assert not got and b'refuses an unlinked typescript-go library call' in err
bypass=scratch/'checker-gap-bypass.a';bypass.write_text("console.log('checker probe bypassed');\n")
got,err=run('checker-gap-bypass-node',['node','--disable-warning=ExperimentalWarning',repository/'oracle/node.mjs',bypass]);assert got and not err
emitted,err=run('checker-gap-bypass-emitted',[stage0,'js',bypass]);assert emitted and not err
from check_quote import validate_quote
quote_summary=validate_quote(scratch,run,stage0,owned,repository)
(scratch/'summary.json').write_text(json.dumps(dict(reporting_only=True,cases=len(cases),findings=counts,bytes=len(expected),mutants=mutants,guards=guard_results,metadata_mutants=metadata_mutants,checker_gap=dict(source_node_exit=70,emitted_js_exit=1,bypass_mutant_exit=0),quote=quote_summary,commands=commands),indent=2)+'\n')
print('PARTIAL REPORTING PASS',counts,len(expected),'bytes; three ID mutants caught only by Go bytes')

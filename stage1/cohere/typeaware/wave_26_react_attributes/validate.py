"""Owned three-rule production Go byte oracle and successful native mutants."""
import json, pathlib, subprocess, time, os, re
ROOT=pathlib.Path(__file__).resolve().parents[4]
OWN=pathlib.Path(__file__).resolve().parent
S=pathlib.Path('/workspace/wave-26-attributes')
records=[]
def run(name, command, cwd=ROOT, expected=0):
    started=time.monotonic()
    with (S/(name+'.stdout')).open('wb') as out, (S/(name+'.stderr')).open('wb') as err:
        result=subprocess.run([str(x) for x in command],cwd=cwd,stdout=out,stderr=err)
    records.append(dict(name=name,command=[str(x) for x in command],exit=result.returncode,seconds=time.monotonic()-started))
    (S/'runs.json').write_text(json.dumps(records,indent=2)+'\n')
    assert result.returncode==expected,(name,result.returncode,(S/(name+'.stderr')).read_text())
    return (S/(name+'.stdout')).read_bytes()
virtual=ROOT/'cohere/adamic_wave26_attributes_oracle.go'
overlay=S/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(OWN/'testdata/oracle.go')}}))
oracle=S/'oracle'
run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],ROOT/'cohere')
extract=S/'extract';run('extract-build',['go','build','-o',extract,OWN/'testdata/extract.go'])
inputs=[]
for name in ['button_has_type','checked_requires_onchange_or_readonly','display_name']:
    raw=run(name+'-extract',[extract,ROOT/('cohere/internal/lint/rules/react/'+name+'_test.go')])
    inputs+=json.loads(raw)['Inputs']
inputs += ["const C = React.memo(() => <div/>);", "const b = <button type={c ? 'bad' : 'wrong'} />;", "declare const kind:'button'|'submit';const b=<button type={kind}/>;", "const b=<input checked defaultChecked/>;", "// 世界🌍\r\nexport default function(){return <div/>;}","declare const React:any;const {createElement}=React;createElement('input',{checked:1});", "interface C{}; const C=React.memo(()=> <div/>); C.displayName='C';"]
inputs += ["let ß;ß=()=> <div/>;", "const a=<button type={'\\u0085\\u00a0\\u200b\\u{10ffff}'}/>;", "const a=<button type={'世界🌍'}/>;"]
files=[]
for i, text in enumerate(dict.fromkeys(inputs)):
    file=S/('control-%03d.tsx'%i);file.write_text(text+'\nexport {};\n');files.append(file)
config=S/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'ESNext','jsx':'preserve'},'files':[str(files[0])]}))
manifest=S/'controls.manifest';manifest.write_text(''.join(str(f)+'\n' for f in files))
valid=run('parse-filter',[oracle,config,manifest,'--valid-sources']);manifest.write_bytes(valid)
print('controls: '+str(len(valid.splitlines()))+' parsed, '+str(len(files)-len(valid.splitlines()))+' nonsources excluded',flush=True)
run('native-build',['/workspace/wave-26-rawhir/adamic','build',OWN/'suite.a','-o',S/'native','--tsgo',S/'checker.a'])
run('asan-build',['/workspace/wave-26-rawhir/adamic','build',OWN/'suite.a','-o',S/'native-asan','--tsgo',S/'checker.a','--sanitize'])
truth=None
for name, conf, paths in [('controls',config,manifest),('compiler','/workspace/wave-26-typescript/src/compiler/tsconfig.json','/workspace/wave-26-compiler.manifest'),('repository',ROOT/'tsconfig.json','/workspace/wave-26-repository.manifest')]:
    expected=run(name+'-go',[oracle,conf,paths])
    actual=run(name+'-native',[S/'native',conf,paths]);assert actual==expected,name+' native'
    sanitized=run(name+'-asan',[S/'native-asan',conf,paths]);assert sanitized==expected,name+' ASAN'
    for mode in ['native','asan']:assert (S/(name+'-'+mode+'.stderr')).read_bytes()==b''
    print(name+': '+str(len(expected))+' identical bytes, '+expected.splitlines()[-1].decode(),flush=True)
    if name=='controls':truth=expected
for directory, needle, replacement in [('button_has_type',"if(!proven) {","if(proven) {"),('checked_requires_onchange_or_readonly',"if(!names.has('checked'))","if(names.has('checked'))"),('display_name',"if(named.has(id)) {continue;}","if(!named.has(id)) {continue;}")]:
    target=S/(directory+'-mutant');target.mkdir(exist_ok=True)
    file=OWN/directory/'rule.a';source=file.read_text();assert source.count(needle)==1
    source=source.replace(needle,replacement)
    def absolute(match):return "from '"+str((file.parent/match.group(1)).resolve())+"'"
    source=re.sub(r"from '([^']+)'",lambda m:m.group(0) if m.group(1)=='adamic' else absolute(m),source)
    (target/'rule.a').write_text(source)
    entry=OWN/'suite.a';source=entry.read_text()
    source=source.replace("from './"+directory+"/rule.a'","from '"+str(target/'rule.a')+"'")
    source=re.sub(r"from '([^']+)'",lambda m:m.group(0) if m.group(1)=='adamic' or m.group(1).startswith('/') else "from '"+str((entry.parent/m.group(1)).resolve())+"'",source)
    (target/'suite.a').write_text(source)
    run(directory+'-mutant-build',['/workspace/wave-26-rawhir/adamic','build',target/'suite.a','-o',target/'native','--tsgo',S/'checker.a'])
    assert run(directory+'-mutant-run',[target/'native',config,manifest])!=truth
    assert (S/(directory+'-mutant-run.stderr')).read_bytes()==b''
    print(directory+' mutant: exit 0, empty stderr, complete-byte oracle caught it',flush=True)

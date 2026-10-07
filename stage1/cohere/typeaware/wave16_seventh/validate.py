#!/usr/bin/env python3
"""Owned suite; no shared registration or harness mutations."""
import gzip, hashlib, json, os, pathlib, re, shutil, statistics, subprocess, sys, time
repo=pathlib.Path(__file__).resolve().parents[4]
owned=pathlib.Path(__file__).resolve().parent
art=pathlib.Path(sys.argv[1]).resolve();art.mkdir(parents=True,exist_ok=True)
results=[]
def run(name,args,cwd=repo,env=None,code=0):
    start=time.monotonic()
    with (art/(name+'.stdout')).open('wb') as out,(art/(name+'.stderr')).open('wb') as err:
        p=subprocess.run(list(map(str,args)),cwd=cwd,env=env,stdout=out,stderr=err)
    row={'name':name,'args':list(map(str,args)),'cwd':str(cwd),'exit':p.returncode,'seconds':time.monotonic()-start}
    results.append(row);(art/'validation.json').write_text(json.dumps(results,indent=2))
    assert p.returncode==code,(name,row,(art/(name+'.stderr')).read_text()[-3000:])
    return row
stage=art/'adamic';archive=art/'checker.a';san=art/'checker-asan.a'
# Archive dispatch and question file are added by overlays, never source edits.
source=(repo/'bridge/tsgo/checker/facts.go').read_text();hook='\tout.text(mode)\n'
assert source.count(hook)==1
(art/'facts.go').write_text(source.replace(hook,hook+'\tif answer, handled, err := p.additionalAnswer(out,c,node,question); handled {return answer,err}\n'))
overlay={'Replace':{str(repo/'bridge/tsgo/checker/facts.go'):str(art/'facts.go'),str(repo/'bridge/tsgo/checker/wave16_resolved_signature_declaration.go'):str(owned/'jsx_no_constructed_context_values/resolved_signature_declaration.go')}}
(art/'checker-overlay.json').write_text(json.dumps(overlay))
(art/'oracle-overlay.json').write_text(json.dumps({'Replace':{str(repo/'cohere/adamic_wave16_seventh.go'):str(owned/'oracle.go')}}))
run('stage0-build',['go','build','-o',stage,'./cmd/adamic'])
run('checker-build',['go','build','-tags','adamic_wave16_seventh','-overlay',art/'checker-overlay.json','-buildmode=c-archive','-o',archive,'./bridge/tsgo/archive'])
run('suite-build',[stage,'build',owned/'suite.a','-o',art/'suite','--tsgo',archive])
run('oracle-build',['go','build','-overlay',art/'oracle-overlay.json','-o',art/'oracle','./adamic_wave16_seventh.go'],repo/'cohere')
run('extract',['go','run',owned/'extract_controls.go','--',*[repo/('cohere/internal/lint/rules/react/'+n+'_test.go') for n in ['jsx_fragments','jsx_no_undef','jsx_no_constructed_context_values','jsx_no_constructed_context_values_stability']]])
rows=json.loads((art/'extract.stdout').read_text());paths=[]
for index,row in enumerate(rows):
    f=art/f'control-{index:03}.tsx';f.write_text(row['Source']+'\nexport{};\n');paths.append(str(f))
# Additional controls cover Go %q Unicode, CRLF and the newer memo diagnostic ID.
extra=['/* 世界 🌍 */\r\nfunction Ω(){const Ωvalue={a:1};return <Ctx.Provider value={Ωvalue}/>;}\r\n',
       'function C(){const value=React.useMemo(()=>({a:1}));return <Ctx.Provider value={value}/>;}',
       'function C(){const options={a:1};return <Ctx.Provider value={useMemo(()=>({options}),[options])}/>;}',
       '<Map/>;<Date/>;', 'declare global{var GlobalView:any;}export{};<GlobalView/>;']
for index,text in enumerate(extra):
    f=art/f'extra-{index:03}.tsx';f.write_text(text+'\nexport{};\n');paths.append(str(f))
(art/'controls.manifest').write_text('\n'.join(paths)+'\n')
config=art/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'ESNext','moduleResolution':'Bundler','jsx':'preserve','lib':['ES2022','DOM']},'files':['control-000.tsx']}))
def compare(name,binary,config,manifest,flags=()):
    go=run(name+'-go',[art/'oracle',config,manifest,'--context',*flags])
    native=run(name+'-native',[binary,config,manifest,'--context',*flags])
    truth=(art/(name+'-go.stdout')).read_bytes();actual=(art/(name+'-native.stdout')).read_bytes()
    assert truth==actual,(name,'diagnostic bytes differ')
    assert not (art/(name+'-native.stderr')).read_bytes(),(name,'native stderr')
    native['byte_equal']=True;native['go_seconds']=go['seconds'];native['ratio']=native['seconds']/go['seconds']
    return truth
truth=compare('controls',art/'suite',config,art/'controls.manifest')
for name in ['jsx-fragments','jsx-no-undef','jsx-no-constructed-context-values']:
    assert ('\treact/'+name+'\t').encode() in truth
for profile,flags in [('element',['--element']),('globals',['--allow-globals']),('both',['--element','--allow-globals'])]:
    compare('controls-'+profile,art/'suite',config,art/'controls.manifest',flags)
env=os.environ.copy();env.update(CC='clang',CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all')
run('checker-asan-build',['go','build','-tags','adamic_wave16_seventh','-overlay',art/'checker-overlay.json','-buildmode=c-archive','-o',san,'./bridge/tsgo/archive'],env=env)
run('suite-asan-build',[stage,'build',owned/'suite.a','-o',art/'suite-asan','--tsgo',san,'--sanitize'])
for profile,flags in [('default',[]),('element',['--element']),('globals',['--allow-globals']),('both',['--element','--allow-globals'])]:
    compare('asan-'+profile,art/'suite-asan',config,art/'controls.manifest',flags)
for name,file,old,new in [('fragment-name','jsx_fragments/rule.a',"c.node(right).text === 'Fragment'","c.node(right).text === 'WrongFragment'"),('undef-intrinsic','jsx_no_undef/rule.a','first >= 97 && first <= 122','first >= 65 && first <= 122'),('context-object','jsx_no_constructed_context_values/construction.a',"label = 'object';","label = 'array';")]:
    directory=art/(name+'-source');shutil.copytree(owned,directory,dirs_exist_ok=True)
    for f in directory.rglob('*.a'):
        origin=owned/f.relative_to(directory);text=f.read_text()
        def imports(match):
            value=match.group(1);target=(origin.parent/value).resolve()
            try:target.relative_to(owned);return match.group(0)
            except ValueError:return "from '"+str(target)+"'"
        text=re.sub(r"from '([^']+)'",imports,text)
        if f.relative_to(directory).as_posix()==file:
            assert text.count(old)==1; text=text.replace(old,new)
        f.write_text(text)
    binary=art/name
    run(name+'-build',[stage,'build',directory/'suite.a','-o',binary,'--tsgo',archive])
    row=run(name+'-run',[binary,config,art/'controls.manifest','--context'])
    actual=(art/(name+'-run.stdout')).read_bytes()
    assert not (art/(name+'-run.stderr')).read_bytes() and actual!=truth,(name,'survived')
    row['caught_by']='independent Go diagnostic bytes';row['first_difference']=next(i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b)
corpora=[('compiler',pathlib.Path('/workspace/wave16-corpus/typescript/src/compiler/tsconfig.json'),pathlib.Path('/workspace/wave16-artifacts/compiler.manifest')),('repository',repo/'tsconfig.json',pathlib.Path('/workspace/wave16-artifacts/repository.manifest'))]
for name,conf,manifest in corpora:
    compare(name,art/'suite',conf,manifest)
    compare(name+'-asan',art/'suite-asan',conf,manifest)
    timings=[]
    for sample in range(3):
        timings.append(compare(name+f'-timing-{sample}',art/'suite',conf,manifest))
    pairs=[r for r in results if r['name'].startswith(name+'-timing') and r['name'].endswith('-native')]
    (art/(name+'-timing.json')).write_text(json.dumps({'native_median':statistics.median(r['seconds'] for r in pairs),'go_median':statistics.median(r['go_seconds'] for r in pairs)},indent=2))
# Released handle refusal and its surviving-registry counterexample.
(art/'released.a').write_text("import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';\nconst args=programArguments(),file=args[1]??'';const p=tsgoProgram(args[0]??'',[file]);tsgoRelease(p);console.log(tsgoInspect(p,file,0,1,'Identifier','node-symbol-details'));\n")
(art/'probe.a').write_text('x;\n')
run('released-build',[stage,'build',art/'released.a','-o',art/'released','--tsgo',archive])
run('released-run',[art/'released',config,art/'probe.a'],code=70)
assert (art/'released-run.stderr').read_text()=='adamic: panic: invalid or released checker handle\n'
raw=(repo/'bridge/tsgo/archive/main.go').read_text();needle='delete(programs.live, uint64(handle))';assert raw.count(needle)==1
(art/'registry.go').write_text(raw.replace(needle,'// Mutant keeps the handle.'))
held=json.loads((art/'checker-overlay.json').read_text());held['Replace'][str(repo/'bridge/tsgo/archive/main.go')]=str(art/'registry.go');(art/'registry-overlay.json').write_text(json.dumps(held))
run('released-mutant-archive',['go','build','-tags','adamic_wave16_seventh','-overlay',art/'registry-overlay.json','-buildmode=c-archive','-o',art/'registry.a','./bridge/tsgo/archive'])
run('released-mutant-build',[stage,'build',art/'released.a','-o',art/'released-mutant','--tsgo',art/'registry.a'])
run('released-mutant-run',[art/'released-mutant',config,art/'probe.a'])
# The preserved shared compiler gap must remain an explicit compile refusal.
run('foreign-arena-gap-build',[stage,'build',owned/'gaps/foreign_arena.a','-o',art/'foreign-gap','--tsgo',archive],code=1)
assert b'escaping a constructor' in (art/'foreign-arena-gap-build.stderr').read_bytes()
# A foreign source body reaches the explicit boundary, with Go reporting it normally.
(art/'foreign-helper.a').write_text('export function build(){return {a:1};}\n')
(art/'foreign-helper.ts').symlink_to(art/'foreign-helper.a')
(art/'foreign-client.tsx').write_text("import {build} from './foreign-helper.ts';function Component(){const built=build();const value=useMemo(()=>({built}),[built]);return <Ctx.Provider value={value}/>;}\nexport{};\n")
(art/'foreign.manifest').write_text(str(art/'foreign-client.tsx')+'\n')
run('foreign-go',[art/'oracle',config,art/'foreign.manifest','--context'])
run('foreign-native',[art/'suite',config,art/'foreign.manifest','--context'],code=70)
assert b'cross-file context-value callee arena' in (art/'foreign-native.stderr').read_bytes()
(art/'validation.json').write_text(json.dumps(results,indent=2))
print('PASS: four profiles, three rule mutants, two corpora, sanitizer, released handle; cross-file compiler gap reproduced')

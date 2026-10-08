#!/usr/bin/env python3
"""Isolated validation against the current integrated parser, without editing shared files."""
import gzip, hashlib, json, os, pathlib, shutil, subprocess, sys, time
ROOT=pathlib.Path(__file__).resolve().parents[5]
OWN=pathlib.Path(__file__).resolve().parent
WORK=pathlib.Path('/workspace/wave-22-seventh-work')
WORK.mkdir(exist_ok=True)
JSX=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip()
def run(label,args,cwd=ROOT,ok=True,env=None):
    before=time.perf_counter()
    with (WORK/(label+'.stdout')).open('wb') as out,(WORK/(label+'.stderr')).open('wb') as err:
        p=subprocess.run([str(a) for a in args],cwd=cwd,stdout=out,stderr=err,env=env)
    duration=time.perf_counter()-before
    (WORK/(label+'.json')).write_text(json.dumps({'command':[str(a) for a in args],'exit':p.returncode,'seconds':duration})+'\n')
    if ok and p.returncode:raise RuntimeError(label+': '+(WORK/(label+'.stderr')).read_text()[-4000:])
    print(label,'exit',p.returncode,'seconds',round(duration,3),flush=True)
    return p.returncode
COPY=WORK/'source'
for path in (ROOT/'stage1').rglob('*'):
    if path.is_file() and path.suffix in ('.a','.ts'):
        destination=COPY/path.relative_to(ROOT);destination.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(path,destination)
ENTRY=COPY/OWN.relative_to(ROOT)/'driver.a'
COMPILER='/workspace/wave-22-seventh-adamic';ARCHIVE='/workspace/wave-22-seventh-checker.a'
run('native-build',[COMPILER,'build',ENTRY,'-o',WORK/'native','--tsgo',ARCHIVE])
overlay={'Replace':{str(ROOT/'cohere/adamic_wave22_seventh_oracle.go'):str(OWN/'oracle.go.txt')}}
(WORK/'oracle-overlay.json').write_text(json.dumps(overlay))
run('oracle-build',['go','build','-overlay',WORK/'oracle-overlay.json','-o',WORK/'oracle',ROOT/'cohere/adamic_wave22_seventh_oracle.go'],ROOT/'cohere')
# Read literal source payloads with Go's own parser/unquoter, then let the independent
# production parser reject non-source message strings and malformed inputs explicitly.
extractor=WORK/'extract.go'
extractor.write_text('''package main
import("go/ast";"go/parser";"go/token";"os";"strconv";"encoding/json")
func main(){var values []string;seen:=map[string]bool{};for _,p:=range os.Args[1:]{t,e:=parser.ParseFile(token.NewFileSet(),p,nil,0);if e!=nil{panic(e)};ast.Inspect(t,func(n ast.Node)bool{if v,ok:=n.(*ast.BasicLit);ok&&v.Kind==token.STRING{s,e:=strconv.Unquote(v.Value);if e==nil&&!seen[s]{values=append(values,s);seen[s]=true}};return true})};json.NewEncoder(os.Stdout).Encode(values)}
''')
run('extract-build',['go','build','-o',WORK/'extract',extractor])
tests=[ROOT/'cohere/internal/lint/rules/react'/f'{name}_test.go' for name in ['no_object_type_as_default_prop','no_unstable_nested_components','sort_default_props']]
run('extract',[WORK/'extract',*tests])
sources=json.loads((WORK/'extract.stdout').read_text())
sources += [
 'function C({[ /* keep */ k() ]: x = {},[\ufeffk()]: y=[]}){return <div/>;}',
 '/* 🌍 */ function C({[\u00a0o.p]: x={},[ /* a */ 1+2 ]: y=[]}){return <div/>;}',
 'function C({a={},b=[],c=()=>{},d=function(){},e=class{},f=new Date(),g=<div/>,h=/x/,i=Symbol(),j=<></>,k=Symbol.for(1)}){return <div/>;}',
 'function Parent(){function Child(){return <div/>;}return <Child/>;}',
 'function Parent(){const Child=()=>null;return <div/>;}',
 'class C extends React.Component{static defaultProps={z:1,a:2,b:3,...x,d:1,c:2};render(){return <div/>;}}',
 'const defaults={z:1,a:2}; C.defaultProps=defaults;',
 'function Parent(){return <Thing renderFooter={()=> <div/>} footer={()=> <div/>}/>;}',
 'function Parent(){return xs.map(()=> <div/>);}',
 'function Parent(){return useMemo(()=>{return ()=> <div/>;},[]);}',
 'function ÜParent(){function ÜChild(){return <div/>;}return <ÜChild/>;}',
 'function ßParent(){function Child(){return <div/>;}return <Child/>;}',
 'function _Parent(){function Child(){return <div/>;}return <Child/>;}',
 'class C extends React.Component{static defaultProps={Z:1,a:2,İ:3,i:4,Σ:5,σ:6};}',
 'function C({a:renamed={},"string":s=[],[key]:v={},1:n=/x/}){return <div/>;}',
]

config=WORK/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','jsx':'preserve','lib':['ES2022','DOM'],'allowJs':True,'types':[],'moduleDetection':'auto'}}))
paths=[]
for i,source in enumerate(sources):
    p=WORK/f'control-{i:04d}.tsx';p.write_text(source+'\nexport {};\n');paths.append(str(p))
manifest=WORK/'controls.manifest';manifest.write_text('\n'.join(paths)+'\n')
run('valid-sources',[WORK/'oracle',config,manifest,'--valid-sources'])
accepted=(WORK/'valid-sources.stdout').read_text().splitlines();manifest.write_text('\n'.join(accepted)+'\n')
print('controls accepted',len(accepted),'rejected',len(paths)-len(accepted),flush=True)
def compare(label,binary,configuration,listing,options=()):
    run(label+'-go',[WORK/'oracle',configuration,listing,*options]);run(label+'-native',[binary,configuration,listing,*options])
    a=(WORK/(label+'-go.stdout')).read_bytes();b=(WORK/(label+'-native.stdout')).read_bytes()
    if a!=b:
        position=next((i for i,(x,y) in enumerate(zip(a,b)) if x!=y),min(len(a),len(b)))
        raise RuntimeError(label+' differs at '+str(position)+'\nGO '+repr(a[max(0,position-150):position+400])+'\nNATIVE '+repr(b[max(0,position-150):position+400]))
    if (WORK/(label+'-native.stderr')).read_bytes():raise RuntimeError(label+' native stderr')
    print(label,'identical bytes',len(a),flush=True)
    return a
truth=compare('controls',WORK/'native',config,manifest)
for name in ['no-object-type-as-default-prop','no-unstable-nested-components','sort-default-props']:
    if ('\treact/'+name+'\t').encode() not in truth:raise RuntimeError('No positive '+name)
compare('ignore-case',WORK/'native',config,manifest,['--ignore-case'])
compare('allow-as-props',WORK/'native',config,manifest,['--allow-as-props'])
if '--controls-only' in sys.argv:sys.exit(0)
run('archive-asan',['go','build','-buildmode=c-archive','-asan','-o',WORK/'checker-asan.a','./bridge/tsgo/archive'])
run('native-asan-build',[COMPILER,'build',ENTRY,'-o',WORK/'native-asan','--tsgo',WORK/'checker-asan.a','--sanitize'])
compare('controls-asan',WORK/'native-asan',config,manifest)
for name,relative,before,after in [
 ('default-id','no-object-type-as-default-prop/rule.a',"'forbiddenTypeDefaultParam'","'forbiddenTypeDefaultParamMutated'"),
 ('nested-id','no-unstable-nested-components/rule.a',"'unstableNestedComponent'","'unstableNestedComponentMutated'"),
 ('sort-id','sort-default-props/rule.a',"'propsNotSorted'","'propsNotSortedMutated'")
]:
    target=ENTRY.parent/relative;original=target.read_text();assert before in original;target.write_text(original.replace(before,after,1))
    try:
        run(name+'-build',[COMPILER,'build',ENTRY,'-o',WORK/name,'--tsgo',ARCHIVE]);run(name+'-run',[WORK/name,config,manifest])
        result=(WORK/(name+'-run.stdout')).read_bytes()
        if result==truth or (WORK/(name+'-run.stderr')).read_bytes():raise RuntimeError(name+' survived or crashed')
        print(name,'byte-only mutant caught',flush=True)
    finally:target.write_text(original)
for corpus,corpusroot,corpusconfig in [('repository',ROOT,ROOT/'tsconfig.json'),('compiler',pathlib.Path('/workspace/wave-22-typescript-pinned'),pathlib.Path('/workspace/wave-22-typescript-pinned/src/compiler/tsconfig.json'))]:
    paths=[str(corpusroot/p) for p in (ROOT/f'stage1/cohere/typeaware/validation-coverage/{corpus}.manifest').read_text().splitlines() if p]
    listing=WORK/(corpus+'.manifest');listing.write_text('\n'.join(paths)+'\n')
    compare(corpus,WORK/'native',corpusconfig,listing);compare(corpus+'-asan',WORK/'native-asan',corpusconfig,listing)

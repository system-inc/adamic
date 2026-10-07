"""Private native/production-Go comparison. Every subprocess writes to files."""
from pathlib import Path
import gzip,json,os,re,subprocess,time,statistics
ROOT=Path(__file__).resolve().parents[4];HERE=Path(__file__).resolve().parent
OUT=Path(os.environ.get('WAVE05_CORE_ARTIFACTS','/workspace/wave-05-core-validation'));OUT.mkdir(exist_ok=True)
def run(name,args,cwd=ROOT,env=None,expected=0):
    begin=time.monotonic()
    with (OUT/(name+'.stdout')).open('wb') as stdout,(OUT/(name+'.stderr')).open('wb') as stderr:
        code=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=stdout,stderr=stderr).returncode
    assert code==expected,(name,code,(OUT/(name+'.stderr')).read_text())
    return (OUT/(name+'.stdout')).read_bytes(),(OUT/(name+'.stderr')).read_bytes(),time.monotonic()-begin
stage0=OUT/'adamic';archive=OUT/'checker.a';native=OUT/'native';oracle=OUT/'oracle'
run('stage0',['go','build','-o',stage0,'./cmd/adamic'])
run('archive',['go','build','-buildmode=c-archive','-o',archive,'./bridge/tsgo/archive'])
run('native-build',[stage0,'build',HERE/'main.a','-o',native,'--tsgo',archive])
virtual=ROOT/'cohere/adamic_wave05_core_oracle.go';overlay=OUT/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(HERE/'testdata/oracle.go')}}))
run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],ROOT/'cohere')
controls=[
    'Symbol(); (Symbol)(); Symbol("label"); Symbol(undefined); globalThis.Symbol(); new Symbol();',
    'function f(Symbol:()=>symbol){Symbol();} {let Symbol=()=>0;Symbol();}',
    'declare function Symbol():symbol;Symbol();',
    'const {Symbol}= {Symbol:()=>0};Symbol();',
    'declare const value:unknown;typeof value === "array";typeof value !== "null";typeof value === undefined;typeof value === null;',
    'declare const value:unknown;typeof value === "object"; "strng" === typeof value; typeof value === `date`; typeof value == 0; typeof value === true;',
    'declare const value:unknown;function f(undefined:string){return typeof value===undefined;}',
    'declare const value:unknown;typeof value === -1;typeof value===typeof value;typeof value === /x/;',
    '/* 世界 🌍 */\r\nSymbol();typeof foo===undefined;async function é(){return 1;}\r\n',
]
# All quoted source rows, plus typed raw blocks from the pinned require-await tests.
text=(ROOT/'cohere/internal/lint/rules/core/require_await_test.go').read_text()
for match in re.finditer(r'^\s*\{\s*("(?:[^"\\]|\\.)*")\s*,',text,re.M):
    source=json.loads(match[1])
    if any(x in source for x in ['async','function','class']):controls.append(source)
prelude=re.search(r'prelude := `([^`]*)`',text).group(1)
for source in re.findall(r'`([^`]*)`',text):
    if '\n' in source and any(x in source for x in ['async','function useMemo']):
        if 'useMemo(' in source and 'function useMemo' not in source:source=prelude+source
        controls.append(source)
# Extra binding, promise-contract and generic controls.
controls += [
    'const f:()=>Promise<number>=async()=>1;',
    'const f:()=>number|Promise<number>=async()=>1;',
    'function id<T>(x:T):T{return x;}id(async()=>1);',
    'function id<T>(x:T):T{return x;}id<()=>Promise<number>>(async()=>1);',
    'function require<T extends ()=>Promise<number>>(x:T):T{return x;}require(async()=>1);',
    'function use<T>(x:()=>Promise<T>):void{}use(async()=>1);',
    'interface I{f():Promise<number>;}class C implements I{async f(){return 1;}}',
    'async function f(){for(const x of []){}} async function g(){for await(const x of []){}}',
    'async function f(){const x=1;}async function g(){await using x=resource();}',
    'class A{a:number\nasync [x](){return 1;}}',
    'class A{a:()=>number\nasync [x](){return 1;}}',
    'class A{a:number=0\nasync [x](){return 1;}}',
    'interface I{f():Promise<number>;}const C=class implements I{async f(){return 1;}};',
    'function pick<T>(a:T,b:T):T{return b;}const fn:()=>Promise<number>=async()=>1;pick(fn,async()=>2);',

]
paths=[]
for at,source in enumerate(dict.fromkeys(controls)):
    path=OUT/f'control-{at:03d}.a';path.write_text(source+'\nexport {};\n');paths.append(path)
manifest=OUT/'controls.manifest';manifest.write_text(''.join(str(p)+'\n' for p in paths))
prelude_file=OUT/'prelude.d.ts';prelude_file.write_text('declare const wave05Marker: unique symbol;\n')
config=OUT/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022','DOM']},'files':[str(prelude_file)]}))
valid,_,_=run('parse-filter',[oracle,config,manifest,'--valid-sources'])
manifest.write_bytes(valid);print('controls',len(paths),'parse-valid',len(valid.splitlines()),flush=True)
def compare(name,conf,files,exe):
    truth,_,_=run(name+'-go',[oracle,conf,files]);actual,error,_=run(name+'-native',[exe,conf,files])
    (OUT/(name+'.oracle.gz')).write_bytes(gzip.compress(truth,mtime=0))
    if actual!=truth:(OUT/(name+'.expected')).write_bytes(truth);(OUT/(name+'.actual')).write_bytes(actual)
    assert actual==truth and not error,(name,'byte mismatch',error)
    print(name,len(truth),'identical bytes',truth.splitlines()[-1].decode(),flush=True);return truth
truth=compare('controls',config,manifest,native)
# Compilation, sanitizers, mutants and corpora follow once the controls pass.
if os.environ.get('WAVE05_CORE_CONTROLS_ONLY'):raise SystemExit(0)
asan_archive=OUT/'sanitized.a';san=OUT/'sanitized'
run('asan-archive',['go','build','-buildmode=c-archive','-o',asan_archive,'./bridge/tsgo/archive'],env=dict(os.environ,CC='clang',CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all'))
run('asan-build',[stage0,'build',HERE/'main.a','-o',san,'--tsgo',asan_archive,'--sanitize'])
compare('controls-asan',config,manifest,san)
for name,file,old,new in [
    ('symbol','symbol_description/rule.a','declarations[0] === true','declarations[0] === false'),
    ('typeof','valid_typeof/rule.a','!results.includes(operand.text)','results.includes(operand.text)'),
    ('await','require_await/rule.a','awaiting(context, body, true)','!awaiting(context, body, true)'),
]:
    import shutil
    folder=OUT/('mutant-'+name)
    if folder.exists():
        for p in folder.rglob('*.a'):p.unlink()
    shutil.copytree(HERE,folder,dirs_exist_ok=True)
    p=folder/file;text=p.read_text();assert text.count(old)==1;text=text.replace(old,new);p.write_text(text)
    for p in folder.rglob('*.a'):
        original=HERE/p.relative_to(folder)
        text=p.read_text()
        def dependency(m):
            path=(original.parent/m[1]).resolve()
            return "from '"+str(path)+"'" if m[1].startswith('.') and not path.is_relative_to(HERE) else m[0]
        p.write_text(re.sub(r"from '([^']+)'",dependency,text))
    exe=OUT/('mutant-'+name+'-bin');run(name+'-mutant-build',[stage0,'build',folder/'main.a','-o',exe,'--tsgo',archive])
    wrong,error,_=run(name+'-mutant-run',[exe,config,manifest]);assert not error and wrong!=truth
    print(name,'mutant exits 0, empty stderr; only Go bytes catch it',flush=True)
release=OUT/'released'
run('release-build',[stage0,'build',HERE/'testdata/released.a','-o',release,'--tsgo',archive])
probe=OUT/'released-probe.a';probe.write_text('probe;\n')
for question in ['declared-call-signature','type-projection\n1\nindex']:
    _,error,_=run('released-'+question.split('\n')[0],[release,config,probe,question],expected=70)
    assert error==b'adamic: panic: invalid or released checker handle\n'
    print(question.split('\n')[0],'released handle exact refusal 70',flush=True)
for name,conf,files in [('repository',ROOT/'tsconfig.json',Path('/workspace/wave-05-repository.manifest')),('compiler',Path('/workspace/wave-05-typescript/src/compiler/tsconfig.json'),Path('/workspace/wave-05-compiler.manifest'))]:
    compare(name,conf,files,native);compare(name+'-asan',conf,files,san)
print('PASS three core rules, controls, mutants, corpora and sanitizers',flush=True)

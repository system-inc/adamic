"""Raw-graph agreement only. This does not validate the three lint analyses."""
import argparse
import json
import pathlib
import subprocess
import time

ROOT = pathlib.Path(__file__).resolve().parents[4]
OWN = pathlib.Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('--scratch', required=True)
scratch = pathlib.Path(parser.parse_args().scratch).resolve()
scratch.mkdir(parents=True, exist_ok=True)
records = []


def run(name, command, expected=0, env=None):
    started = time.monotonic()
    with (scratch / (name + '.stdout')).open('wb') as out, (scratch / (name + '.stderr')).open('wb') as err:
        result = subprocess.run([str(x) for x in command], cwd=ROOT, stdout=out, stderr=err, env=env)
    records.append(dict(name=name, command=[str(x) for x in command], exit=result.returncode, seconds=time.monotonic() - started))
    (scratch / 'runs.json').write_text(json.dumps(records, indent=2) + '\n')
    assert result.returncode == expected, (name, result.returncode, (scratch / (name + '.stderr')).read_text())
    return (scratch / (name + '.stdout')).read_bytes()


stage0 = scratch / 'adamic'
run('stage0', ['go', 'build', '-o', stage0, './cmd/adamic'])
archive = scratch / 'checker.a'
run('archive', ['go', 'build', '-buildmode=c-archive', '-o', archive, './bridge/tsgo/archive'])
virtual = ROOT / 'cohere/adamic_wave26_hir_oracle.go'
overlay = scratch / 'oracle-overlay.json'
overlay.write_text(json.dumps({'Replace': {str(virtual): str(OWN / 'testdata/raw_hir_oracle.go')}}))
oracle = scratch / 'oracle'
run('oracle-build', ['go', '-C', ROOT / 'cohere', 'build', '-overlay', overlay, '-o', oracle, virtual])
prelude = 'type Dispatch<S>=(value:S)=>void;interface RefObject<T>{current:T;}declare function useState(v:number):[number,Dispatch<number>];declare function useEffect(fn:()=>void):void;declare function useMemo(fn:()=>number):number;declare function useCallback<T>(fn:T,deps:any[]):T;declare function useRef<T>(v:T):RefObject<T>;'
examples = [
    'function Component(){const[x,setX]=useState(0);useEffect(()=>{setX(1)});return x;}',
    'function Component(){const[x,setX]=useState(0);setX(1);return x;}',
    'function Component(){const Inner = () => null;return <Inner/>;}',
    "function Component(){const Inner=()=> '世界🌍';return <Inner/>;}",
    'function Component(){const[x,setX]=useState(0);useMemo(()=>{setX(1);return 1;});return x;}',
    'function Component(flag:boolean){let C; if(flag){C=make();}else{C=Default;}return <C/>;}',
    'function Component(){const[x,setX]=useState(0);const f=()=>setX(1);const g=()=>f();useEffect(()=>g());return x;}',
    'function Component(){const[x,setX]=useState(0);const f=useCallback(()=>setX(1),[]);useEffect(()=>f());return x;}',
    'function Component(){const[x,setX]=useState(0);const f=useMemo(()=>()=>setX(1));useEffect(()=>f());return x;}',
    'function Component(){const[x,setX]=useState(0);const r=useRef(1);useEffect(()=>{if(r.current){setX(r.current);}});return x;}',
    'function f(x:any){const {a,b:[c,...d],...e}=x;return ()=>a+c+d+e;}',
    'function f(x:any){for(let i=0;i<3;i++){x+=i;if(x)continue;}return x;}',
    'function f(x:any){while(x){x--;if(x===2)break;}do{x++;}while(x<3);return x;}',
    'function f(x:any){for(const k in x){x[k]++;}for(const y of x){foo(y);}return x;}',
    'function f(x:any){try{return x.a();}catch(e){return e;}finally{foo();}}',
    'function f(x:any){switch(x){case 1:x++;break;default:x=2;}return x;}',
    'function f(x:any){return x?.a.b ?? x?.[foo()]?.();}',
    'function f(x:any){return x ? foo(x) : bar(x);}',
    'function f(x:any){return (()=>x+1)();}',
    'function Component(){const f=use\\u0043allback(()=>1,[]);return f();}',
    'function factory(x:any){function Component(){return <x.C/>;}return Component;}',
    'function f(x:any){class C{};return new C();}',
]
files = []
for i, text in enumerate(examples):
    file = scratch / ('control-%03d.tsx' % i)
    file.write_text(prelude + text + '\nexport {};\n')
    files.append(file)
config = scratch / 'tsconfig.json'
config.write_text(json.dumps({'compilerOptions': {'strict': True, 'target': 'ES2022', 'module': 'ESNext', 'jsx': 'preserve'}, 'files': [str(files[0])]}))
manifest = scratch / 'controls.manifest'
manifest.write_text(''.join(str(file) + '\n' for file in files))
truth = run('production-hir', [oracle, config, manifest])
entry = OWN / 'hir_probe.a'
binary = scratch / 'native'
run('native-build', [stage0, 'build', entry, '-o', binary, '--tsgo', archive])
assert run('native-run', [binary, config, manifest]) == truth
assert (scratch / 'native-run.stderr').read_bytes() == b''
print('raw HIR: %d controls, %d exact production bytes' % (len(files), len(truth)), flush=True)
asan = scratch / 'native-asan'
run('asan-build', [stage0, 'build', entry, '-o', asan, '--tsgo', archive, '--sanitize'])
assert run('asan-run', [asan, config, manifest]) == truth
assert (scratch / 'asan-run.stderr').read_bytes() == b''
wire = ROOT / 'stage1/cohere/typeaware/wave_26_react_reporting/rawhir/wire.go'
text = wire.read_text()
needle = 'Construct(f)'
assert text.count(needle) == 2
mutant_source = scratch / 'no-ssa.go'
mutant_source.write_text(text.replace(needle, '// Mutant omits initial SSA construction.', 1))
mutant_overlay = scratch / 'no-ssa.json'
mutant_overlay.write_text(json.dumps({'Replace': {str(wire): str(mutant_source)}}))
mutant_archive = scratch / 'no-ssa.a'
run('ssa-mutant-archive', ['go', 'build', '-overlay', mutant_overlay, '-buildmode=c-archive', '-o', mutant_archive, './bridge/tsgo/archive'])
mutant_binary = scratch / 'no-ssa'
run('ssa-mutant-build', [stage0, 'build', entry, '-o', mutant_binary, '--tsgo', mutant_archive])
assert run('ssa-mutant-run', [mutant_binary, config, manifest]) != truth
assert (scratch / 'ssa-mutant-run.stderr').read_bytes() == b''
print('omitted SSA mutant: exit 0, empty stderr, production bytes catch it', flush=True)
# Retained serialized data must survive release, and fresh questions must refuse.
probe = scratch / 'released.a'
probe.write_text("import {programArguments,tsgoProgram,tsgoRelease} from 'adamic';import {ReactHIR} from " + repr(str(OWN / 'react_hir.a')) + ";const args=programArguments();const file=args[1]??'';const p=tsgoProgram(args[0]??'',[file]);const view=new ReactHIR(p,file,0,Number(args[2]??'0'),'SourceFile');tsgoRelease(p);console.log(view.serialized);console.log(new ReactHIR(p,file,0,Number(args[2]??'0'),'SourceFile').serialized);")
released = scratch / 'released'
run('released-build', [stage0, 'build', probe, '-o', released, '--tsgo', archive])
output = run('released-run', [released, config, files[0], len(files[0].read_bytes())], 70)
assert output == truth.splitlines(keepends=True)[1]
assert (scratch / 'released-run.stderr').read_bytes() == b'adamic: panic: invalid or released checker handle\n'
registry = ROOT / 'bridge/tsgo/archive/main.go'
text = registry.read_text()
needle = 'delete(programs.live, uint64(handle))'
assert text.count(needle) == 1
mutant_source = scratch / 'retained.go'
mutant_source.write_text(text.replace(needle, '// Mutant retains the released handle.'))
mutant_overlay = scratch / 'retained.json'
mutant_overlay.write_text(json.dumps({'Replace': {str(registry): str(mutant_source)}}))
mutant_archive = scratch / 'retained.a'
run('registry-mutant-archive', ['go', 'build', '-overlay', mutant_overlay, '-buildmode=c-archive', '-o', mutant_archive, './bridge/tsgo/archive'])
mutant_binary = scratch / 'retained'
run('registry-mutant-build', [stage0, 'build', probe, '-o', mutant_binary, '--tsgo', mutant_archive])
assert run('registry-mutant-run', [mutant_binary, config, files[0], len(files[0].read_bytes())]) == output + output
assert (scratch / 'registry-mutant-run.stderr').read_bytes() == b''
print('released data survives; stale question panics 70; registry mutant exits 0', flush=True)
print('RAW GRAPH PASS; three lint source analyses remain unfinished.', flush=True)

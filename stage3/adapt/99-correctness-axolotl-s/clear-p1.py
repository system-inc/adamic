import sys
sys.dont_write_bytecode = True
import hashlib,json,os,subprocess,sys
from pathlib import Path
from importlib.machinery import SourceFileLoader
here=Path(__file__).resolve().parent
supervise=SourceFileLoader('supervise',str(here/'clear-supervise.py')).load_module()
out=supervise.scratch/'p1';out.mkdir(exist_ok=True)
binary=str(supervise.scratch/'adamic')
# Read the applied owner, without substituting its parameter type or body.
text=(supervise.scratch/'before/src/compiler/core.ts').read_text()
start=text.index('export function clear<T>');end=text.index('\n}',start)+2
body=text[start:end]
cases={
'class': '''class Item { value: number; constructor(value: number) { this.value = value; } }
function exercise(): void {
    const values: Item[] = [new Item(1), new Item(2), new Item(3)];
    const alias = values;
    clear(values);
    console.log(`${values.length}:${alias.length}`);
    values.push(new Item(4));
    console.log(`${values.length}:${values[0]!.value}`);
    clear(values);
    console.log(`${values.length}`);
}
exercise();
''',
'number': '''const values: number[] = [1, 2, 3];
const alias = values;
clear(values);
console.log(`${values.length}:${alias.length}`);
values.push(4);
console.log(`${values.length}:${values[0]!}`);
clear(values);
console.log(`${values.length}`);
''',
'empty': '''const values: number[] = [];
clear(values);
console.log(`${values.length}`);
clear(values);
console.log(`${values.length}`);
'''}
results={}
def run(name,cmd,env=None):
 r=supervise.run(name,cmd,env=env,limit=1800)
 if r['exit_code']!=0:sys.exit(r['exit_code'] if r['exit_code']>=0 else 128-r['exit_code'])
 return supervise.scratch/(name+'.log')
for name,tail in cases.items():
 source=out/(name+'.ts');source.write_text(body+'\n'+tail)
 js=out/(name+'.mjs');run(name+'-js',[binary,'js',str(source)]).replace(js)
 # The original TypeScript owner erased by stock tsc is Node's oracle.
 run(name+'-tsc',['tsc',str(source),'--target','es2022','--module','esnext','--outDir',str(out/'oracle')])
 oracle=out/'oracle'/(name+'.js');oracle.rename(out/'oracle'/(name+'.mjs'));oracle=out/'oracle'/(name+'.mjs')
 expected=run(name+'-node',['node',str(oracle)]).read_bytes()
 actual=run(name+'-javascript',['node',str(js)]).read_bytes();assert actual==expected,(name,'javascript')
 native=out/(name+'-native');run(name+'-build',[binary,'build',str(source),'-o',str(native)])
 assert run(name+'-native',[str(native)]).read_bytes()==expected,(name,'native')
 if name=='class':
  sanitized=out/'class-sanitized';run('class-sanitize-build',[binary,'build',str(source),'-o',str(sanitized),'--sanitize'])
  env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1:halt_on_error=1',UBSAN_OPTIONS='halt_on_error=1')
  assert run('class-sanitized',[str(sanitized)],env).read_bytes()==expected
 results[name]={'native':'Node-identical','javascript':'Node-identical','stdout':expected.decode(),'sanitizersWithLeakDetection':name=='class','sourceSha256':hashlib.sha256(source.read_bytes()).hexdigest(),'nativeSha256':hashlib.sha256(native.read_bytes()).hexdigest()}
 (here/'clear-p1-results.json').write_text(json.dumps(results,indent=2)+'\n')

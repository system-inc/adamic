from pathlib import Path
import subprocess as s,time,json,re,os
r=Path(__file__).resolve().parent;repo=Path('/workspace/scratch/watch-file-1532-merge');runner=r/'runproof.py'
while not (r/'build-results.json').exists():time.sleep(1)
assert not json.loads((r/'build-results.json').read_text())['failures']
b=json.loads((r/'builds.json').read_text());results=[]
def run(name,cmd,expected=0):
 code=s.run(['python3',str(runner),name,str(expected),*map(str,cmd)],cwd=repo).returncode
 results.append({'name':name,'matchedExpected':not code});return code
run('builtin-built-witness',['node',repo/'stage3/adapt/93-builtin-imports/proofs/built-witness.cjs',b['builtin-before'],b['builtin-after'],r/'builtin-built-fixture',r/'builtin-built-witness.json'])
run('crypto-built-witness',['node',repo/'stage3/adapt/99-wasm-crypto-import/proofs/built-witness.cjs',b['crypto-before'],b['candidate'],r/'crypto-built-fixture',r/'crypto-built-witness.json'])
run('resolution-host-runtime-fixture',['node',repo/'stage3/adapt/99-wasm-crypto-import/proofs/resolve-host-fixture.cjs',Path(b['crypto-before'])/'built/local/typescript.js',Path(b['candidate'])/'built/local/typescript.js',r/'resolution-host-runtime'])
d=r/'interval-proof';(d/'baseline-typescript.cjs').symlink_to(Path(b['baseline'])/'built/local/typescript.js')
# Original scripts are path-parameterized only. Their after-tree alias selects the freshly built candidate.
(d/'after-tree').unlink();(d/'after-tree').symlink_to(b['candidate'],target_is_directory=True)
text=(r/'after/src/compiler/sys.ts').read_bytes().decode();pattern=r'        if \(defaultPollingInterval !== PollingInterval.Low && defaultPollingInterval !== PollingInterval.Medium && defaultPollingInterval !== PollingInterval.High\) \{\r?\n            return Debug.fail\(`Invalid dynamic polling interval: \$\{defaultPollingInterval\}`\);\r?\n        \}\r?\n';mutant,n=re.subn(pattern,'',text);assert n==1;nothing=Path(d/'sys-mutant.ts').write_bytes(mutant.encode())
run('dynamic-interval-original-witness',['node',d/'witness.cjs']);run('non-dynamic-interval-10-fixture',['node',d/'interval-10.cjs'])
# Independent red control: the same expected invalid-interval error must reject a guard-restoring mutant.
script=d/'guard-check.cjs';script.write_text('''const fs=require('fs'),vm=require('vm'),assert=require('assert/strict'),ts=require('typescript');
const main=require('./baseline-typescript.cjs'),text=fs.readFileSync(process.argv[2],'utf8'),sf=ts.createSourceFile('sys.ts',text,ts.ScriptTarget.Latest,true),fn=sf.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='createDynamicPriorityPollingWatchFile');assert(fn);
const js=ts.transpileModule(fn.getText(sf),{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;
const create=vm.runInNewContext(js+'\\ncreateDynamicPriorityPollingWatchFile;',{...main});
const register=interval=>create({getModifiedTime:()=>new Date(0),setTimeout:()=>({})})('/probe.ts',()=>{},interval);
for(const interval of [250,500,2000])register(interval).close();
for(const interval of [10,undefined])assert.throws(()=>register(interval),error=>error.name==='Error'&&error.message.includes('Invalid dynamic polling interval: '+String(interval)));
console.log(JSON.stringify({members:[250,500,2000],invalid10AndUndefinedFailAtBoundary:true}));
''')
run('dynamic-interval-guard-check',['node',script,r/'after/src/compiler/sys.ts']);run('dynamic-interval-restored-guard-mutant',['node',script,d/'sys-mutant.ts'],1)
# Baseline/candidate declarations must be byte identical.
declarations=[]
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
 a=(Path(b['baseline'])/'built/local'/name).read_bytes();z=(Path(b['candidate'])/'built/local'/name).read_bytes();declarations.append({'file':name,'bytes':len(a),'identical':a==z});assert a==z
(r/'declaration-comparison.json').write_text(json.dumps(declarations,indent=2)+'\n')
(r/'runtime-results.json').write_text(json.dumps({'results':results,'allExpected':all(x['matchedExpected'] for x in results)},indent=2)+'\n');print('RUNTIME COMPLETE',flush=True)

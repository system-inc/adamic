import pathlib, subprocess, json, hashlib, shutil
out=pathlib.Path(__file__).resolve().parent
before=pathlib.Path('/workspace/tracing-before')
after=pathlib.Path('/workspace/tracing-after')
a={p.relative_to(before/'built/local'):p for p in (before/'built/local').rglob('*.js')}
b={p.relative_to(after/'built/local'):p for p in (after/'built/local').rglob('*.js')}
assert a.keys()==b.keys(),(a.keys(),b.keys())
assert len(a)==10,len(a)
records=[]
for rel in sorted(a):
 x,y=a[rel].read_bytes(),b[rel].read_bytes()
 assert x==y, f'JavaScript moved: {rel}'
 records.append(dict(file=str(rel),bytes=len(x),sha256=hashlib.sha256(x).hexdigest()))
(out/'javascript.json').write_text(json.dumps(records,indent=2)+'\n')
script="const a=require(process.argv[1]); console.log(JSON.stringify({strings:Object.getOwnPropertyNames(a).sort(),symbols:Object.getOwnPropertySymbols(a).map(String).sort()}));"
keys=[]
for tree in [before,after]:
 r=subprocess.run(['node','-e',script,str(tree/'built/local/typescript.js')],capture_output=True,text=True)
 assert r.returncode==0,r.stderr
 keys.append(json.loads(r.stdout))
assert keys[0]==keys[1]
(out/'export-keys.json').write_text(json.dumps(keys[0],indent=2)+'\n')
project=pathlib.Path('/workspace/tracing-witness');project.mkdir(exist_ok=True)
(project/'input.ts').write_text('type Pair<T> = [T, T];\ntype Maybe<T> = T extends string ? {text:T} : {value:T};\nfunction pair<T>(value:T): Pair<T> { return [value,value]; }\nconst result = pair<string | number>("trace");\n')
# Same absolute project and library paths on both runs.
(project/'tsconfig.json').write_text(json.dumps({'compilerOptions':{'strict':True,'noEmit':True,'lib':[]},'files':['input.ts',str(before/'src/lib/es5.d.ts')]}))
trace=project/'trace'
for label,tree in [('before',before),('after',after)]:
 if trace.exists():shutil.rmtree(trace)
 with (out/('witness-'+label+'.log')).open('w') as log:
  r=subprocess.run(['node','-e',"const ts=require(process.argv[1]); ts.sys.getExecutingFilePath=()=>'/workspace/tracing-before/built/local/tsc.js'; ts.sys.exit=code=>{ts.dumpTracingLegend();process.exitCode=code;}; ts.executeCommandLine(ts.sys,ts.noop,process.argv.slice(2));",str(tree/'built/local/typescript.js'),'--project',str(project/'tsconfig.json'),'--generateTrace',str(trace)],stdout=log,stderr=subprocess.STDOUT,cwd=project)
 (out/('witness-'+label+'.exit.json')).write_text(json.dumps({'exit':r.returncode})+'\n')
 assert r.returncode==0,r.returncode
 for file in ['types.json','legend.json']:shutil.copyfile(trace/file,out/(label+'-'+file))
for file in ['types.json','legend.json']:
 x=(out/('before-'+file)).read_bytes();y=(out/('after-'+file)).read_bytes();assert x==y,file
 print(file,len(x),hashlib.sha256(x).hexdigest())
print('10 JavaScript files identical; export keys identical; witness identical')

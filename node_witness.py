import hashlib,json,pathlib,shutil,subprocess
out=pathlib.Path(__file__).parent
project=pathlib.Path('/workspace/scratch/tracing-witness-project');project.mkdir(exist_ok=True)
(project/'main.ts').write_text('export function greet(name: string): string { return "Hello, " + name; }\nexport const result = greet("trace");\n')
(project/'tsconfig.json').write_text(json.dumps(dict(compilerOptions=dict(strict=True,declaration=True,outDir='out',target='es2020',module='commonjs'),files=['main.ts']),indent=2)+'\n')
records=[];observations={}
def snapshot(root):
    return {str(p.relative_to(root)):p.read_bytes() for p in sorted(root.rglob('*')) if p.is_file()}
def normalize(value):
    if isinstance(value,list): return [normalize(v) for v in value]
    if isinstance(value,dict): return {k:('<timestamp>' if k in ['ts','dur'] else '<pid>' if k=='pid' else normalize(v)) for k,v in value.items()}
    return value
for mode in ['plain','trace']:
    observations[mode]={}
    for label in ['control','after']:
        for name in ['out','trace']:
            p=project/name
            if p.exists():shutil.rmtree(p)
        command=['node','/workspace/scratch/tracing-'+label+'/built/local/tsc.js','--project',str(project/'tsconfig.json'),'--pretty','false']
        if mode=='trace':command+=['--generateTrace',str(project/'trace')]
        child=subprocess.run(command,cwd=project,capture_output=True)
        prefix=label+'-'+mode
        (out/(prefix+'.stdout')).write_bytes(child.stdout);(out/(prefix+'.stderr')).write_bytes(child.stderr)
        row=dict(mode=mode,tree=label,argv=command,exit=child.returncode);records.append(row)
        (out/'node-witness-commands.json').write_text(json.dumps(records,indent=2)+'\n')
        print(prefix,'child exit='+str(child.returncode),flush=True)
        if child.returncode:raise SystemExit(child.returncode)
        emitted=snapshot(project/'out'); traces=snapshot(project/'trace') if mode=='trace' else {}
        target=out/(prefix+'-files');target.mkdir()
        for name,data in {**{'out/'+k:v for k,v in emitted.items()},**{'trace/'+k:v for k,v in traces.items()}}.items():
            p=target/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(data)
        observations[mode][label]=dict(stdout=child.stdout,stderr=child.stderr,emitted=emitted,traces=traces)
    a,b=observations[mode]['control'],observations[mode]['after']
    for key in ['stdout','stderr','emitted']:assert a[key]==b[key],(mode,key)
    assert a['traces'].keys()==b['traces'].keys()
    for name,data in a['traces'].items():
        assert normalize(json.loads(data))==normalize(json.loads(b['traces'][name])),(mode,name,'trace changes outside timestamps and pids')
exports=[]
for label in ['control','after']:
    command=['node','-e','process.stdout.write(JSON.stringify(Object.keys(require(process.argv[1])).sort()))','/workspace/scratch/tracing-'+label+'/built/local/typescript.js']
    child=subprocess.run(command,capture_output=True);print(label,'export child exit='+str(child.returncode),flush=True)
    if child.returncode:raise SystemExit(child.returncode)
    (out/(label+'-export-keys.json')).write_bytes(child.stdout);exports.append(child.stdout)
assert exports[0]==exports[1]
(out/'node-witness.json').write_text(json.dumps(dict(plainIdentical=True,traceIdenticalExcept=['ts','dur','pid'],emittedIdentical=True,runtimeExportKeysIdentical=True),indent=2)+'\n')

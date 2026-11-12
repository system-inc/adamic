import pathlib,subprocess,json,hashlib,shutil
out=pathlib.Path(__file__).parent;roots=[pathlib.Path('/workspace/scratch/tracing-control2'),pathlib.Path('/workspace/scratch/tracing-after3')];keys=[];commands=[];project=pathlib.Path('/workspace/scratch/tracing-witness-project2');plain=[]
for label,root in zip(['control2','after3'],roots):
 args=['node','-e','process.stdout.write(JSON.stringify(Object.keys(require(process.argv[1])).sort()))',str(root/'built/local/typescript.js')];child=subprocess.run(args,capture_output=True);print(label,'export child exit='+str(child.returncode),flush=True);commands.append(dict(phase='exports',label=label,argv=args,exit=child.returncode));(out/(label+'-runtime-keys.json')).write_bytes(child.stdout)
 if child.returncode:raise SystemExit(child.returncode)
 keys.append(json.loads(child.stdout));target=project/'out'
 if target.exists():shutil.rmtree(target)
 args=['node',str(root/'built/local/tsc.js'),'--project',str(project/'tsconfig.json'),'--pretty','false'];child=subprocess.run(args,capture_output=True);print(label,'plain CLI child exit='+str(child.returncode),flush=True);commands.append(dict(phase='plain',label=label,argv=args,exit=child.returncode));(out/(label+'-plain.stdout')).write_bytes(child.stdout);(out/(label+'-plain.stderr')).write_bytes(child.stderr)
 if child.returncode:raise SystemExit(child.returncode)
 plain.append((child.stdout,child.stderr,{str(p.relative_to(target)):p.read_bytes() for p in target.rglob('*') if p.is_file()}))
assert keys[0]==keys[1] and len(keys[0])==2256 and 'errorMessage' in keys[0];assert plain[0]==plain[1]
rows=[]
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
 a=(roots[0]/'built/local'/name).read_bytes();b=(roots[1]/'built/local'/name).read_bytes();assert a==b;rows.append(dict(file=name,bytes=len(a),sha256=hashlib.sha256(a).hexdigest()))
(out/'public2-commands.json').write_text(json.dumps(commands,indent=2)+'\n');(out/'public2-contracts.json').write_text(json.dumps(dict(runtimeExportKeysIdentical=True,runtimeExportCount=2256,errorMessageIsRuntimeExport=True,publicDeclarations=rows,plainStdoutStderrAndEmittedFilesIdentical=True),indent=2)+'\n')

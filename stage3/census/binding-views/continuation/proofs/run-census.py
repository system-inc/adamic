import pathlib,subprocess,os,time,json,hashlib,sys
root=pathlib.Path('/workspace/scratch/binding-views')
name=sys.argv[1];tree=root/name
out=root/'evidence'/name;out.mkdir(parents=True,exist_ok=True)
for file,mode,binary in [('resolutionCache.ts','full','latent'),('program.ts','scan only','scan-only')]:
 env=os.environ|dict(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(tree/'src/compiler'/file),CENSUS_MODULE_FORMAT='CommonJSBundle')
 command=[str(root/binary),str(tree/'src/compiler'),str(out/(file+'.jsonl'))]
 start=time.monotonic()
 with (out/(file+'.log')).open('w') as log: p=subprocess.run(command,env=env,stdout=log,stderr=subprocess.STDOUT)
 result=dict(command=command,exit=p.returncode,mode=mode,seconds=time.monotonic()-start,binary_sha256=hashlib.sha256((root/binary).read_bytes()).hexdigest(),input_sha256=hashlib.sha256((tree/'src/compiler'/file).read_bytes()).hexdigest())
 (out/(file+'.run.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
 if p.returncode:sys.exit(p.returncode)
 data=[json.loads(line) for line in (out/(file+'.jsonl')).read_text().splitlines()]
 assert data[-1].get('status')=='relation_report'
 assert data[0]['module_format']=='CommonJSBundle' and data[0]['latent_mode']==mode

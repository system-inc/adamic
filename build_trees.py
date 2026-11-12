import json,pathlib,subprocess,time
out=pathlib.Path(__file__).parent; rows=[]
for label in ['control','after']:
    tree=pathlib.Path('/workspace/scratch/tracing-'+label)
    for phase,args in [('install',['npm','ci','--no-audit','--no-fund']),('build',['npm','run','build']),('harness',['npm','exec','hereby','--','tests'])]:
        begin=time.monotonic()
        with (out/(label+'-'+phase+'.log')).open('w') as log: child=subprocess.run(args,cwd=tree,stdout=log,stderr=subprocess.STDOUT)
        rows.append(dict(tree=label,phase=phase,argv=args,exit=child.returncode,seconds=time.monotonic()-begin));(out/'build-commands.json').write_text(json.dumps(rows,indent=2)+'\n')
        print(label,phase,'child exit='+str(child.returncode),flush=True)
        if child.returncode:raise SystemExit(child.returncode)

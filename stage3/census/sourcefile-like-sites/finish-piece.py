import pathlib,subprocess,sys
out=pathlib.Path(__file__).resolve().parent;repo=out.parents[2]
commands=[['python3',str(out/'seal-inputs.py')],['python3',str(out/'collect.py')],['node',str(out/'classify.cjs')],['python3',str(out/'report.py')],['python3',str(out/'checkpoint.py'),sys.argv[1]]]
for i,command in enumerate(commands):
 log=out/'preparation'/('piece-'+sys.argv[1]+'-step-'+str(i)+'.log')
 with log.open('wb') as dest:child=subprocess.run(command,cwd=repo,stdout=dest,stderr=subprocess.STDOUT)
 print('child exit='+str(child.returncode)+' '+repr(command),flush=True)
 if child.returncode:sys.exit(child.returncode)

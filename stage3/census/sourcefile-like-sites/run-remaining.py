import pathlib,subprocess,sys,json
out=pathlib.Path(__file__).resolve().parent;repo=out.parents[2]
for piece in range(int(sys.argv[1]),9):
 for name,args in [('run',['python3',str(out/'run-piece.py'),str(piece)]),('finish',['python3',str(out/'finish-piece.py'),str(piece)])]:
  log=out/'preparation'/('piece-'+str(piece)+'-'+name+'.log')
  with log.open('wb') as dest: child=subprocess.run(args,cwd=repo,stdout=dest,stderr=subprocess.STDOUT)
  result=dict(piece=piece,step=name,child_exit=child.returncode)
  (out/'preparation'/('piece-'+str(piece)+'-'+name+'-status.json')).write_text(json.dumps(result,indent=2)+'\n')
  print(json.dumps(result),flush=True)
  if child.returncode:sys.exit(child.returncode)

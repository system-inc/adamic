"""Run unfinished top-level tests without repeating completed tests."""
import json, pathlib, subprocess, sys
out=pathlib.Path(__file__).resolve().parent
package=sys.argv[1]; names=json.loads(pathlib.Path(sys.argv[2]).read_text()); results=[]
for index in range(0,len(names),24):
 selected=names[index:index+24]; cmd=['go','test','./'+package+'/...','-run','^('+'|'.join(selected)+')$','-count=1','-parallel=4','-timeout','90s','-json'];log=out/'logs'/('batch-'+package.replace('/','_')+'-'+str(index//24)+'.jsonl')
 with log.open('w') as f:
  try:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=120);code=r.returncode
  except subprocess.TimeoutExpired:code=124
 results.append(dict(command=cmd,exit=code,log=str(log.relative_to(out)),tests=selected));(out/('batches-'+package.replace('/','_')+'.json')).write_text(json.dumps(results,indent=2)+'\n')
 print(package,index//24,code,flush=True)
sys.exit(int(any(r['exit'] for r in results)))

import json,subprocess,os,pathlib,time,concurrent.futures
rows=json.load(open('/tmp/u053/rows.json'))
def run(id):
 results=[]
 for row in rows:
  env=os.environ.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u053/cache/'+id
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^'+row+'$'];start=time.monotonic()
  with open('/tmp/u053/'+id+'-alone-'+row+'.log','w') as out:p=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
  results.append(dict(id=id,test=row,returncode=p.returncode,wall=time.monotonic()-start,command='ADAMIC_MUTANT='+id+' '+' '.join(cmd)));print(id,row,p.returncode,flush=True)
 pathlib.Path('/tmp/u053/'+id+'-alone-times.json').write_text(json.dumps(results,indent=2))
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:list(pool.map(run,['P15','P22']))

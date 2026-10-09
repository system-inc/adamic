import pathlib,json,subprocess,os,time
root=pathlib.Path('review/test-audit/internal-native-split_units'); menu=json.loads((root/'menu.json').read_text());probes=json.loads((root/'probes.json').read_text()); rows=menu['rows']; regex='^('+'|'.join(rows)+')$'
ids=['clean']+[m['id'] for m in menu['mutations'] if m['id']!='S4']+[p['id'] for p in probes]
results=[]
for id in ids:
 env=os.environ.copy();env['ADAMIC_MUTANT']='' if id=='clean' else id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u053/cache/'+id
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',regex]
 start=time.monotonic()
 with open('/tmp/u053/'+id+'.log','w') as out:p=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
 elapsed=time.monotonic()-start;results.append(dict(id=id,returncode=p.returncode,wall=elapsed,command='ADAMIC_MUTANT='+env['ADAMIC_MUTANT']+' ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd)))
 pathlib.Path('/tmp/u053/run-times.json').write_text(json.dumps(results,indent=2))
 print(id,p.returncode,round(elapsed,3),flush=True)
 if id=='clean' and p.returncode:break

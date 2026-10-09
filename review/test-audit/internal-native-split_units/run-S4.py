import json,pathlib,subprocess,time,os
root=pathlib.Path('review/test-audit/internal-native-split_units');menu=json.loads((root/'menu.json').read_text());m=next(x for x in menu['mutations'] if x['id']=='S4');p=pathlib.Path(m['file']);s=p.read_text();start=time.monotonic()
try:
 p.write_text(s.replace(m['before'],m['after']));env=os.environ.copy();env['ADAMIC_MUTANT']='';env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u053/cache/S4';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^('+'|'.join(menu['rows'])+')$']
 with open('/tmp/u053/S4.log','w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
 pathlib.Path('/tmp/u053/S4-time.json').write_text(json.dumps(dict(id='S4',returncode=r.returncode,wall=time.monotonic()-start,command=' '.join(cmd)),indent=2))
finally:p.write_text(s)

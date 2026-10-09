import pathlib,json,subprocess,os,time
root=pathlib.Path('review/test-audit/internal-native-split_units');rows=json.load(open('/tmp/u053/rows.json'));probes=json.loads((root/'probes.json').read_text());orig=json.load(open('/tmp/u053/originals.json'));f='internal/native/split_units_test.go';results=[]
for m in probes[-3:]:
 id=m['id'];sig=m['signature'];s=orig[f];guard='total == total' if id=='P23' else 'value == value' if id=='P24' else 'piece == piece';s=s.replace(sig,sig+'\n\tif '+guard+' { '+m['statement']+' }',1)
 temp=pathlib.Path('/tmp/u053/'+id+'.go');temp.write_text(s);ov=pathlib.Path('/tmp/u053/'+id+'-overlay.json');ov.write_text(json.dumps({'Replace':{str(pathlib.Path(f).resolve()):str(temp)}}))
 env=os.environ.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u053/cache/'+id
 cmd=['timeout','120','go','test','-overlay',str(ov),'-json','-count=1','-timeout','90s','./internal/native/','-run','^('+'|'.join(rows)+')$'];start=time.monotonic()
 with open('/tmp/u053/'+id+'.log','w') as out:p=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
 with open('/tmp/u053/validate-'+id+'.log','w') as out:v=subprocess.run(['go','vet','-overlay',str(ov),'./internal/native/'],stdout=out,stderr=subprocess.STDOUT)
 results.append(dict(id=id,returncode=p.returncode,wall=time.monotonic()-start,command='ADAMIC_MUTANT='+id+' '+' '.join(cmd),validation_returncode=v.returncode));print(id,p.returncode,v.returncode,flush=True)
pathlib.Path('/tmp/u053/setup-probe-times.json').write_text(json.dumps(results,indent=2))

import hashlib,json,os,statistics,subprocess,time
scratch='/workspace/wave-25-landing-second/fifth'
rows=[]
for population,config in [('compiler','/workspace/wave-25-corpus/src/compiler/tsconfig.json'),('repository','/workspace/adamic/tsconfig.json')]:
 manifest=scratch+'/'+population+'.manifest'
 samples=[]
 for index in range(3):
  outputs={};times={}
  for label,binary in [('go','wave25-fifth-oracle'),('native','wave25-fifth')]:
   path=scratch+'/bench-'+population+'-'+str(index)+'-'+label
   env=os.environ.copy()
   if label=='native':env['ADAMIC_TSGO_TIMING']='1'
   with open(path+'.stdout','wb') as out,open(path+'.stderr','wb') as err:
    started=time.perf_counter_ns();result=subprocess.run([scratch+'/'+binary,config,manifest],stdout=out,stderr=err,env=env);elapsed=time.perf_counter_ns()-started
   assert result.returncode==0,(population,label,result.returncode)
   outputs[label]=open(path+'.stdout','rb').read();times[label]=elapsed
  assert outputs['go']==outputs['native'],population
  samples.append(times)
 row={'population':population,'samples_ns':samples,'go_median_ns':statistics.median(s['go'] for s in samples),'native_median_ns':statistics.median(s['native'] for s in samples),'stdout_bytes':len(outputs['go']),'stdout_sha256':hashlib.sha256(outputs['go']).hexdigest()}
 row['native_over_go']=row['native_median_ns']/row['go_median_ns'];rows.append(row)
json.dump(rows,open('/workspace/wave-25-validation/landing-second-bench.json','w'),indent=2)
print(json.dumps(rows,indent=2))

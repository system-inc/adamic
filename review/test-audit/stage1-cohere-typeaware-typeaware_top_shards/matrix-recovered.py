import os,pathlib,json,subprocess,time
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/stage1-cohere-typeaware-typeaware_top_shards';plan=json.loads((p/'plan.json').read_text());scope=json.loads((p/'scope.json').read_text())['names'];records=json.loads((p/'runs.json').read_text()) if (p/'runs.json').exists() else []
typeNames=[n for n in scope if n.startswith('TestTypeAware')];volumeNames=[n for n in scope if n.startswith('TestVolume') or n.startswith('TestProduct')]
# Reach sets come from main.ts and volume_suite.ts imports and the shared builders.
sets={m['id']:[('typeaware',typeNames),('volume',volumeNames)] for m in plan if m['kind']=='production'}
sets['M3']=[('volume-agreement',[n for n in volumeNames if n.startswith('TestVolumeConfigGuardAndMutant_') and n[-3:].isdigit() and n[-3:]!='000']+['TestVolumeConfigGuardAndMutantUnion']),('volume-guard',['TestVolumeConfigGuardAndMutant_000'])]
sets['M4']=sets['M3']
sets.update(W1=[('witness',['TestTypeAwareTopPlanted_000','TestTypeAwareTopPlanted_001','TestTypeAwareTopShardPlantedFailure'])],W2=[('witness',['TestVolumeConfigGuardAndMutantUnion','TestVolumeConfigGuardAndMutantPlantedFailure'])],S1=[('setup',['TestTypeAwareTopPlanted_000','TestTypeAwareTopPlanted_001'])],S2=[('setup',['TestTypeAwareAgreementAndMutants_Setup','TestTypeAwareAgreementAndMutantsUnion'])],S3=[('setup',['TestVolumeConfigGuardAndMutantUnion'])],S4=[('setup',['TestVolumeConfigGuardAndMutant_Setup'])],S5=[('setup',[n for n in scope if n.startswith('TestProduct_volume_guard_')])],P1=[('probe',typeNames)],P2=[('probe',volumeNames)])
def run(id,label,names,env):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run','^('+'|'.join(names)+')$'];log=p/(id+'-'+label+'.log');start=time.monotonic()
 with log.open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
 records.append(dict(id=id,label=label,rows=names,wall=time.monotonic()-start,exit=r.returncode,command=cmd,log=log.name,corpus=env.get('ADAMIC_TYPESCRIPT_SOURCE')));(p/'runs.json').write_text(json.dumps(records,indent=2)+'\n');print(id,label,r.returncode,round(records[-1]['wall'],3),flush=True)
 return log.read_text()
for m in plan:
 if m['id'] in ['M1','M2']: continue
 if m['id']=='M2': sets[m['id']]=[('volume-generated-resume',volumeNames)]
 subprocess.run(['git','apply',str(p/'diffs'/(m['id']+'.diff'))],cwd=root,check=True)
 try:
  env=os.environ.copy();env.update(ADAMIC_TYPESCRIPT_SOURCE='/tmp/u146/pinned-typescript',ADAMIC_TYPEAWARE_BENCH='1',TMPDIR='/workspace/scratch/u146/tmp',ADAMIC_BUILD_CACHE_DIR='/workspace/scratch/u146/cache/'+m['id'],ADAMIC_BUILD_LOG=str(p/(m['id']+'-recovered-builds.log')))
  if m['file'].endswith('.go'):
   with (p/(m['id']+'-vet.log')).open('w') as f:v=subprocess.run(['timeout','90','go','vet','./stage1/cohere/typeaware/'],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
   assert v.returncode==0,(m['id'],'vet')
  if m['id']=='M2': env.pop('ADAMIC_TYPESCRIPT_SOURCE',None);env.pop('ADAMIC_TYPEAWARE_BENCH',None)
  for label,names in sets[m['id']]:
   label+='-recovered'
   output=run(m['id'],label,names,env)
   if 'test timed out' in output:
    # Retain reached generated control rows; optional corpus work now unknown.
    bounded=env.copy();bounded.pop('ADAMIC_TYPESCRIPT_SOURCE',None);bounded.pop('ADAMIC_TYPEAWARE_BENCH',None)
    output=run(m['id'],label+'-generated-only',names,bounded)
   if any(line.startswith('panic: ') for line in [e.get('Output','') for e in [json.loads(l) for l in output.splitlines() if l.startswith('{')] ]) and 'test timed out' not in output:
    # Each grouped row must be independently observed after a binary abort.
    groups=json.loads((p/'groups.json').read_text())
    for group,members in groups.items():
     selected=[n for n in members if n in names]
     if selected:run(m['id'],label+'-alone-'+group.replace(' ','-'),selected,env)
 finally:subprocess.run(['git','restore','--source=HEAD','--',m['file']],cwd=root,check=True)

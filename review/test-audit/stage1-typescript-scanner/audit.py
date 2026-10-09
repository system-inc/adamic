import os,pathlib,subprocess,json,time,difflib
R=pathlib.Path('/workspace/adamic'); E=pathlib.Path('/tmp/u159/evidence'); P='stage1/typescript/scanner'
env=os.environ.copy();env.update(ADAMIC_TYPESCRIPT_SOURCE='/tmp/u159/typescript',ADAMIC_SCANNER_BENCH='1',ADAMIC_SCANNER_PROFILE_DIR='/tmp/u159/artifacts/clean',ADAMIC_SCANNER_PROFILE_SNAPSHOTS='/tmp/u159/artifacts/clean')
rows={'gappush':'^TestGapStandsWhereGapsMdSays$','gapbigint':'^TestBigintGapStandsWhereGapsMdSays$','products':'^TestProduct_Scanner','agreement':'^(TestScannerAgreesWithTypescriptGo_|TestScannerShardCoverage$)','performance':'^TestPerformance$','artifacts':'^TestProfileArtifacts$','snapshots':'^TestProfileSnapshotsAgree$'}
meta=[]
def run(id,pattern='.',v=None):
 en=env.copy()
 if v:
  en.update(ADAMIC_BUILD_CACHE_DIR='/tmp/u159/cache/'+v,ADAMIC_SCANNER_PROFILE_DIR='/tmp/u159/artifacts/'+v,ADAMIC_SCANNER_PROFILE_SNAPSHOTS='/tmp/u159/artifacts/'+v)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./'+P+'/', '-run',pattern]; t=time.monotonic()
 with open(E/(id+'.log'),'w') as f: res=subprocess.run(cmd,cwd=R,env=en,stdout=f,stderr=subprocess.STDOUT)
 item=dict(id=id,command=' '.join(cmd),env={k:en[k] for k in ['ADAMIC_BUILD_CACHE_DIR','ADAMIC_TYPESCRIPT_SOURCE','ADAMIC_SCANNER_BENCH','ADAMIC_SCANNER_PROFILE_DIR','ADAMIC_SCANNER_PROFILE_SNAPSHOTS'] if k in en},wall=time.monotonic()-t,exit=res.returncode)
 meta.append(item);(E/'runs.json').write_text(json.dumps(meta,indent=2));print(id,res.returncode,round(item['wall'],2),flush=True)
for name,pattern in rows.items():
 for i in range(3):run('timing-'+name+'-'+str(i+1),pattern)

import os,pathlib,subprocess,json,time,signal
out=pathlib.Path('review/test-audit/internal-native-math');results=[];pattern='^TestNormalizeMatchesNode(Points[0-9][0-9]|Contexts)$'
for mid in ['family-control-0','family-control-1','family-control-2','M11','M12','M13','P09']:
 env=os.environ.copy();env['ADAMIC_MUTANT']='' if mid.startswith('family') else mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u049/cache/family-'+mid
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',pattern];start=time.monotonic()
 with open(out/('family-'+mid+'.log'),'w') as log:
  p=subprocess.Popen(cmd,stdout=log,stderr=subprocess.STDOUT,env=env,start_new_session=True)
  try:code=p.wait(timeout=120)
  except subprocess.TimeoutExpired:code=124
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
 results.append(dict(id=mid,command='ADAMIC_MUTANT='+env['ADAMIC_MUTANT']+' ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),exit=code,wall=time.monotonic()-start));(out/'family.json').write_text(json.dumps(results,indent=2));print(mid,code,round(results[-1]['wall'],3),flush=True)
 if mid.startswith('family') and code:break

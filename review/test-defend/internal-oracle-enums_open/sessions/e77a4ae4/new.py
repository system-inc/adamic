import os,json,time,subprocess,pathlib
root=pathlib.Path('/workspace/adamic');out=pathlib.Path(__file__).resolve().parent;scratch=pathlib.Path('/tmp/defend-enums');plan=json.loads((out/'plan.json').read_text());variants=plan['variants'];files={f for _,c in variants for f,_,_ in c};base={f:(root/f).read_text() for f in files}
rows=['TestFractionalPowersReachRuntime','TestReviewProgramsAgreeWithNode','TestReviewProgramsRefuse','TestReviewProgramsNoLooseFiles','TestReviewProgramsSelfTest'];pat='^('+'|'.join(rows)+')$';results=[]
def run(id,pat,label):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pat];env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';env['ADAMIC_BUILD_CACHE_DIR']=str(scratch/'cache'/id);t=time.monotonic()
 with (out/(id+'-'+label+'.log')).open('w') as log:r=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 d={'id':id,'label':label,'command':cmd,'wall_seconds':time.monotonic()-t,'exit':r.returncode};results.append(d);(out/'new-commands.json').write_text(json.dumps(results,indent=2));print(d,flush=True);return r.returncode
try:
 for id,changes in [('clean',[])]+variants:
  for f,s in base.items():(root/f).write_text(s)
  for f,a,b in changes:
   s=(root/f).read_text();assert s.count(a)==1;(root/f).write_text(s.replace(a,b))
  run(id,pat,'new-tests')
finally:
 for f,s in base.items():(root/f).write_text(s)

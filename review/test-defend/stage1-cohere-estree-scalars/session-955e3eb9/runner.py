from pathlib import Path
import json,subprocess,difflib,os,time
repo=Path('/workspace/adamic'); p=repo/'review/test-defend/stage1-cohere-estree-scalars/session-955e3eb9'; plan=json.loads((p/'plan.json').read_text()); statuses=[]
pattern='^(TestBoundedPortParser(_[0-9]{3}|Union|PlantedDisagreement)|TestPortStallControlUnion|TestPortStallControlPlantedSurvivor|TestSyntaxMutants_00[012]|TestSyntaxMutantsUnion)$'
for m in plan[1:]:
 for f in {x['file'] for x in plan}:
  (repo/f).write_bytes(subprocess.check_output(['git','show','origin/main:'+f],cwd=repo))
 f=repo/m['file']; s=f.read_text(); assert s.count(m['from'])==1; n=s.replace(m['from'],m['to'],1); f.write_text(n); line=s[:s.index(m['from'])].count('\n')+1; m['line']=line
 (p/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),n.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-estree-scalars/cache/'+m['id'];env['TMPDIR']='/tmp/defend-estree-scalars-tmp'
 for name,pat in [('prepare','^TestProduct_SyntaxMutantsSetup_00[012]$'),('matrix',pattern),('refusals','^TestSyntaxRefusals$')]:
  args=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',pat]; start=time.monotonic()
  with (p/(m['id']+'-'+name+'.log')).open('w') as log:r=subprocess.run(args,cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
  row={'id':m['id'],'phase':name,'exit':r.returncode,'wall':time.monotonic()-start,'command':'ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(args)}; statuses.append(row);(p/'next-status.json').write_text(json.dumps(statuses,indent=2)+'\n');print(row,flush=True)
for f in {x['file'] for x in plan}:(repo/f).write_bytes(subprocess.check_output(['git','show','origin/main:'+f],cwd=repo))
(p/'plan.json').write_text(json.dumps(plan,indent=2)+'\n')

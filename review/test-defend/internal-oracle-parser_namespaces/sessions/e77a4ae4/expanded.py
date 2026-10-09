from pathlib import Path
import subprocess,os,json,time
p=Path('/tmp/defend-parser-namespaces');root=Path('/workspace/adamic')
for n in ['expanded-baseline','review-baseline']:
 es=[json.loads(s) for s in (p/(n+'.log')).read_text().splitlines() if s.startswith('{')]
 assert any(e['Action']=='pass' and not e.get('Test') for e in es),n
commands=json.loads((p/'commands.json').read_text())
for mid in ['D1','D2','D3']:
 subprocess.run(['git','apply',str(p/(mid+'.diff'))],cwd=root,check=True)
 try:
  env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/mid)
  for label,regex in [('expanded','^(TestFractionalPowersReachRuntime|TestReviewProgramsNoLooseFiles|TestReviewProgramsRefuse|TestReviewProgramsSelfTest|TestModuleNamespaceReadsMatchNode|TestNamespaceLiveExportBoundary)$'),('review','^TestReviewProgramsAgreeWithNode$/^smoke\\.a$')]:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',regex];start=time.monotonic()
   with open(p/(mid+'.'+label+'.log'),'w') as log:r=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
   item={'mutant':mid,'label':label,'command':cmd,'exit':r.returncode,'wall':time.monotonic()-start,'cache':env['ADAMIC_BUILD_CACHE_DIR']};commands.append(item);(p/'commands.json').write_text(json.dumps(commands,indent=2)+'\n');print(item,flush=True)
 finally:subprocess.run(['git','apply','-R',str(p/(mid+'.diff'))],cwd=root,check=True)

import json,pathlib,subprocess,os,time
out=pathlib.Path('/tmp/def-view');root=pathlib.Path('/workspace/adamic');cmd=json.loads((out/'expanded-command.json').read_text())+['-skip','^TestNativeAgreesWithNode$']
env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';results=[]
for mid in ['clean','D1','D2','D3']:
 if mid!='clean':subprocess.run(['git','apply',str(out/(mid+'.diff'))],cwd=root,check=True)
 try:
  env['ADAMIC_BUILD_CACHE_DIR']=str(out/'cache'/('expanded-'+mid));t=time.monotonic()
  with open(out/('expanded-'+mid+'.log'),'w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  ev=[json.loads(s) for s in (out/('expanded-'+mid+'.log')).read_text().splitlines() if s.startswith('{')]
  results.append(dict(mutant=mid,seconds=time.monotonic()-t,command=' '.join(cmd),returncode=r.returncode,rows_failed=sorted(set(x['Test'].split('/')[0] for x in ev if x.get('Action')=='fail' and x.get('Test'))),rows_passed=sorted(set(x['Test'] for x in ev if x.get('Action')=='pass' and x.get('Test') and '/' not in x['Test']))))
  (out/'expanded-matrix.json').write_text(json.dumps(results,indent=2))
  if mid=='clean' and r.returncode:break
 finally:
  if mid!='clean':subprocess.run(['git','apply','-R',str(out/(mid+'.diff'))],cwd=root,check=True)

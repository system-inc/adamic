import pathlib,subprocess,json,time,os,shlex
p=pathlib.Path('review/test-audit/stage1-cohere-formatfiles-formatfiles_products');plan=json.loads((p/'plan.json').read_text());names=json.loads((p/'scope.json').read_text())['names'];rs=[]
def run(mid,label,pat):
 env=os.environ.copy();env['ADAMIC_MUTANT']=mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u091/cache/'+(mid if mid in ['S03','S04'] else 'port');pathlib.Path('/tmp/u091-selector').write_text(mid);cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/formatfiles/','-run',pat];t=time.monotonic()
 with open(p/(label+'.log'),'w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
 ev=[]
 for l in (p/(label+'.log')).read_text().splitlines():
  try:ev.append(json.loads(l))
  except:pass
 fails=sorted(set(e['Test'].split('/')[0] for e in ev if e.get('Action')=='fail' and e.get('Test')));panic=any(e.get('Output','').startswith('panic:') for e in ev);rec=dict(id=mid,label=label,command='ADAMIC_MUTANT='+mid+' ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+shlex.join(cmd)+' > '+label+'.log 2>&1',selector_file='/tmp/u091-selector',exit=r.returncode,wall_seconds=time.monotonic()-t,failed_tests=fails,panic=panic,completed_tests=sorted(set(e['Test'].split('/')[0] for e in ev if e.get('Action') in ['fail','pass','skip'] and e.get('Test'))));rs.append(rec);(p/'matrix.json').write_text(json.dumps(rs,indent=2));print(mid,label,r.returncode,len(fails),round(rec['wall_seconds'],2),flush=True);return rec
for m in plan:
 mid=m['id'];rec=run(mid,mid,'.')
 if rec['panic'] or rec['exit']==124:
  for n in names:
   if n not in rec['completed_tests']:run(mid,mid+'-'+n,'^'+n+'$')

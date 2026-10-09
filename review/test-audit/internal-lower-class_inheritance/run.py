import pathlib,subprocess,json,time,os,re
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-audit/internal-lower-class_inheritance'; plans=json.loads((out/'plan.json').read_text()); names=json.loads(pathlib.Path('/tmp/u029/names.json').read_text()); originals={m['file']:(root/m['file']).read_text() for m in plans}; results=[]
def call(cmd,log,env=None):
 start=time.monotonic()
 with (out/log).open('w') as f:r=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT,env=env)
 return {'code':r.returncode,'wall':time.monotonic()-start}
# Verify each replay diff separately on starting sources.
for m in []:
 (root/m['file']).write_text(originals[m['file']].replace(m['old'],m['new']))
 v=call(['timeout','90','go','vet','./'+str(pathlib.Path(m['file']).parent)+'/'],m['id']+'-vet.log')
 (root/m['file']).write_text(originals[m['file']]); m['vet']=v
 if v['code']: print('VET_FAILED',m['id'],flush=True)
(out/'plan.json').write_text(json.dumps(plans,indent=2))
# Switch infrastructure does not enter standalone diffs.
for file,s in originals.items():
 for m in [m for m in plans if m['file']==file]:
  old,new=m['old'],m['new']; selector='os.Getenv("ADAMIC_MUTANT") == "'+m['id']+'"'
  if m['id'].startswith('P_'):
   replacement=old+'\n if '+selector+' { return nil'+(', nil' if m['id']!='P_FIELDS' else '')+' }'
  elif old.startswith('if '):
   # Conditions differ; retain complete condition tail including multiline return.
   a=old.index('{'); b=new.index('{'); replacement='if (('+selector+') && ('+new[3:b].strip()+')) || (('+ 'os.Getenv("ADAMIC_MUTANT") != "'+m['id']+'") && ('+old[3:a].strip()+')) '+old[a:]
  elif old.startswith('property.Flags'):
   replacement='(('+selector+') && ('+new+') || (os.Getenv("ADAMIC_MUTANT") != "'+m['id']+'") && ('+old+'))'
  elif old.startswith('l.classAssignable') or old.startswith('!l.checker'):
   replacement='(('+selector+') && '+new+' || (os.Getenv("ADAMIC_MUTANT") != "'+m['id']+'") && '+old+')'
  elif old.startswith('&CheckError'):
   replacement='func() *CheckError { if '+selector+' { return '+new+' }; return '+old+' }()'
  elif old.startswith('ast.ModifierFlags'):
   replacement='func() ast.ModifierFlags { if '+selector+' { return '+new+' }; return '+old+' }()'
  else: replacement='if '+selector+' { '+new+' } else { '+old+' }'
  s=s.replace(old,replacement)
 if '"os"' not in s:s=s.replace('import (','import (\n "os"',1)
 (root/file).write_text(s)
call(['gofmt','-w',*originals.keys()],'gofmt.log')
(out/'switch.diff').write_text(subprocess.check_output(['git','diff','--',*originals.keys()],cwd=root,text=True))
b=call(['timeout','90','go','test','-c','-o','/tmp/u029/lower.test','./internal/lower/'],'build.log'); (out/'build.json').write_text(json.dumps(b))
if b['code']:print('BUILD_FAILED',flush=True);raise SystemExit(1)
for m in plans:
 env=os.environ.copy();env['ADAMIC_MUTANT']=m['id'];env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u029/cache/'+m['id']
 r=call(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],m['id']+'.log',env); r['id']=m['id']
 text=(out/(m['id']+'.log')).read_text();r['panic']='panic:' in text;r['bounded']=r['panic'] or r['code']==124 or 'test timed out' in text
 if r['bounded']:
  for n in names:call(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+n+'$'],m['id']+'-'+n+'.log',env)
 results.append(r);(out/'runs.json').write_text(json.dumps(results,indent=2));print(m['id'],r,flush=True)
for file,s in originals.items():(root/file).write_text(s)
print('RESTORED',flush=True)

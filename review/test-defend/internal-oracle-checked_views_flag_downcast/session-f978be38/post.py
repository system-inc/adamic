import pathlib,subprocess,os,json,time,difflib
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/internal-oracle-checked_views_flag_downcast/session-f978be38';env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1'
# Wait for the mutation driver to restore source before starting this script.
runs=json.loads((p/'matrix.json').read_text())
def run(label,selector,cover=False,mut='clean'):
 e=env.copy();e['ADAMIC_BUILD_CACHE_DIR']='/tmp/checked-views-defense/cache/'+mut;cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s']
 if cover:cmd+=['-coverpkg=./internal/lower,./internal/native,./internal/javascript','-coverprofile='+str(p/(label+'.cover'))]
 cmd+=['./internal/oracle/','-run',selector];now=time.monotonic()
 with (p/(label+'.log')).open('w') as o:r=subprocess.run(cmd,cwd=root,env=e,stdout=o,stderr=subprocess.STDOUT)
 es=[]
 for s in (p/(label+'.log')).read_text().splitlines():
  try:es.append(json.loads(s))
  except:pass
 rec=dict(label=label,mutant=mut,command='ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+e['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),exit=r.returncode,wall_seconds=time.monotonic()-now,failures=[x.get('Test') for x in es if x.get('Action')=='fail' and x.get('Test')],passes=[x.get('Test') for x in es if x.get('Action')=='pass' and x.get('Test')],cooked=any('test timed out' in x.get('Output','') for x in es))
 runs.append(rec);(p/'matrix.json').write_text(json.dumps(runs,indent=2));print(label,rec['exit'],round(rec['wall_seconds'],2),rec['failures'],flush=True);return rec
names=[]
for label in ['clean-0','clean-1']:
 for s in (p/(label+'.log')).read_text().splitlines():
  try:x=json.loads(s)
  except:continue
  if x['Action']=='run' and '/' not in x.get('Test','') and x.get('Test'):names.append(x['Test'])
rest='^('+'|'.join(n for n in names if n not in ['TestCheckedViewV2TupleIdentityMutant','TestCheckedViewV2CallableProducerMutant'])+')$'
assert not subprocess.check_output(['git','diff','--name-only'],cwd=root)
run('rest-coverage',rest,cover=True)
s='^TestNativeAgreesWithNode$/internal/oracle/testdata/.*(cast|view|predicate|bitwise|truth|tuple|logical|while).*'
if run('native-clean-corrected',s)['exit']:raise RuntimeError('red native baseline')
menu=json.loads((p/'menu.json').read_text())
for m in menu:
 if m['id']=='D3':
  m['id']='D3b';m['old']='value.Type() != ir.Object || source.Flags()';m['new']='value.Type() == ir.Object || source.Flags()';m['change']='flip structural cast receiver admission condition'
 f=root/m['file'];orig=f.read_text();assert orig.count(m['old'])==1
 m['line']=orig[:orig.index(m['old'])].count('\n')+1;new=orig.replace(m['old'],m['new']);(p/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(orig.splitlines(True),new.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
 f.write_text(new)
 try:
  if m['id']=='D3b':
   with (p/'D3b-vet.log').open('w') as o:v=subprocess.run(['go','vet','./internal/lower/'],cwd=root,env=env,stdout=o,stderr=subprocess.STDOUT)
   assert v.returncode==0
   for i in [0,1]:run('D3b-'+str(i),runs[i]['command'].split(' -run ',1)[1],mut=m['id'])
  run(m['id']+'-native-corrected',s,mut=m['id'])
  if m['id'] in ['D1','D2']:
   fixture='stage3/interface-downcasts/v2/'+('tuple-identity-wrong.a' if m['id']=='D1' else 'callable-producer-wrong.a')
   js=p/(m['id']+'-actual.js.txt')
   with js.open('w') as o, (p/(m['id']+'-emit.log')).open('w') as err:r=subprocess.run(['go','run','./cmd/adamic','js',fixture],cwd=root,env=env,stdout=o,stderr=err)
   assert r.returncode==0
   with (p/(m['id']+'-actual.stdout')).open('w') as o,(p/(m['id']+'-actual.stderr')).open('w') as err:r=subprocess.run(['node','--input-type=commonjs'],input=js.read_text(),text=True,cwd=root,env=env,stdout=o,stderr=err)
   (p/(m['id']+'-observation.json')).write_text(json.dumps(dict(fixture=fixture,command='go run ./cmd/adamic js '+fixture+' > '+str(js)+'; node --input-type=commonjs < '+str(js),exit=r.returncode,stdout=(p/(m['id']+'-actual.stdout')).read_text(),stderr=(p/(m['id']+'-actual.stderr')).read_text()),indent=2))
 finally:f.write_text(orig)
(p/'valid-menu.json').write_text(json.dumps(menu,indent=2))
print('done',flush=True)

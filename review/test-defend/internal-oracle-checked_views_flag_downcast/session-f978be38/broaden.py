import pathlib,json,subprocess,os,time
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/internal-oracle-checked_views_flag_downcast/session-f978be38';env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1'
runs=json.loads((p/'matrix.json').read_text());menu=json.loads((p/'valid-menu.json').read_text())
def run(label,s,mut):
 e=env.copy();e['ADAMIC_BUILD_CACHE_DIR']='/tmp/checked-views-defense/cache/'+mut;cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',s];now=time.monotonic()
 with (p/(label+'.log')).open('w') as o:r=subprocess.run(cmd,cwd=root,env=e,stdout=o,stderr=subprocess.STDOUT)
 es=[]
 for x in (p/(label+'.log')).read_text().splitlines():
  try:es.append(json.loads(x))
  except:pass
 rec=dict(label=label,mutant=mut,command='ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+e['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),exit=r.returncode,wall_seconds=time.monotonic()-now,failures=[x.get('Test') for x in es if x.get('Action')=='fail' and x.get('Test')],passes=[x.get('Test') for x in es if x.get('Action')=='pass' and x.get('Test')],cooked=any('test timed out' in x.get('Output','') for x in es));runs.append(rec);(p/'matrix.json').write_text(json.dumps(runs,indent=2));print(label,rec['exit'],round(rec['wall_seconds'],2),rec['failures'],flush=True);return rec
s='^(TestPredicate.*|TestUnknownNarrowingMutants|TestLiteralUndefinedOracleCatchesMutant|TestLiteralOptionalOracleCatchesMutant|TestInUnionNarrowingIsRefused|TestTypeOfStringLiteralMutant|TestNonNullLiteralUnionInitializers)$'
if run('extra-clean',s,'clean')['exit']:raise RuntimeError('red extra baseline')
for m in menu:
 f=root/m['file'];old=f.read_text();assert old.count(m['old'])==1;f.write_text(old.replace(m['old'],m['new']))
 try:run(m['id']+'-extra',s,m['id'])
 finally:f.write_text(old)
r=run('full-native-clean','^TestNativeAgreesWithNode$','clean')
if not r['exit']:
 for m in menu[:2]:
  f=root/m['file'];old=f.read_text();f.write_text(old.replace(m['old'],m['new']))
  try:run(m['id']+'-full-native','^TestNativeAgreesWithNode$',m['id'])
  finally:f.write_text(old)
print('done',flush=True)

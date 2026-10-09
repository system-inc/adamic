import pathlib,subprocess,os,json,time
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/stage1-cohere-graphql-printer-grain_mutant/session'
es=[json.loads(x) for x in (p/'baseline-warm.log').read_text().splitlines() if x.startswith('{')];assert es[-1]['Action']=='pass'
env=os.environ.copy();env.update(ADAMIC_GRAPHQL_PRETTIER='/tmp/printer-defense-prettier',ADAMIC_GRAPHQL_PRINTER_BENCH='1')
selector='^(TestPrinterThroughput|TestPrinterAsGoCohere_[0-9]+|TestPrinterFileDriver|TestPrinterShardPlantedDisagreement|TestPrinterWhitespaceGap_[0-9]+|TestPrinterMutants_[0-9]+)$'
results=[]
for m in json.loads((p/'plan.json').read_text()):
 mid=m['mutant'];f=root/m['file'];original=f.read_bytes();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/printer-defense/cache/'+mid
 try:
  subprocess.run(['git','apply','--check',str(p/(mid+'.diff'))],cwd=root,check=True)
  subprocess.run(['git','apply',str(p/(mid+'.diff'))],cwd=root,check=True)
  runs=[]
  for suffix,sel in [('', '.'),('-bounded',selector)]:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/printer/','-run',sel]
   start=time.monotonic()
   with (p/(mid+suffix+'.log')).open('w') as log:r=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
   ev=[]
   for line in (p/(mid+suffix+'.log')).read_text().splitlines():
    try:ev.append(json.loads(line))
    except:pass
   cooked=r.returncode==124 or any('test timed out' in e.get('Output','') for e in ev)
   result=dict(command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],wall=time.monotonic()-start,exit=r.returncode,cooked=cooked,log=mid+suffix+'.log',rows_failed=[e['Test'] for e in ev if e['Action']=='fail' and e.get('Test')],rows_passed=[e['Test'] for e in ev if e['Action']=='pass' and e.get('Test')],errors=[e['Output'].strip() for e in ev if e.get('OutputType')=='error'])
   runs.append(result);(p/(mid+suffix+'.meta.json')).write_text(json.dumps(result,indent=2))
   if not cooked:break
  results.append(dict(**m,runs=runs));(p/'matrix.json').write_text(json.dumps(results,indent=2))
 finally:f.write_bytes(original)
print('completed')

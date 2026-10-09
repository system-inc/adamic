from pathlib import Path
import subprocess,os,json,time,difflib
p=Path('review/test-defend/internal-lower-library_object')
plans=[
 {'id':'D1','target':'TestLibraryStringRefusals','file':'internal/lower/library_string.go','old':'if method == "toString" || method == "valueOf" {','new':'if method == "toString" || method == "valueOff" {','kind':'change constant','reason':'Stop recognizing valueOf in the String internal-slot guard, admitting valueOf.call(42) through numeric string conversion.'},
 {'id':'D2','target':'TestATupleSeenAsAnArrayIsNotYet','file':'internal/lower/object.go','old':'\tif receiverType == ir.Object && checker.IsTupleType(l.checker.GetTypeAtLocation(receiver)) {\n\t\t// A tuple is held as an object: called as an object, an array method would be read as a field.\n\t\treturn nil, true, l.notYet(node, name+" on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it\'s made)")\n\t}\n','new':'','kind':'drop statement','reason':'Drop the whole tuple-method dispatch guard, leaving tuple.join to the ordinary fallback instead of its representation-specific refusal.'},
 {'id':'D3','target':'TestAMethodReadAsAValueIsRefused','file':'internal/lower/refusals.go','old':'\t\t\t\t\tfor _, parameter := range signatures[0].Parameters() {\n\t\t\t\t\t\tnames = append(names, parameter.Name)\n\t\t\t\t\t}\n','new':'','kind':'drop statement','reason':'Drop the whole parameter-name collection loop; suggested arrows for detached methods no longer forward required parameters.'}
]
for m in plans:
 original=Path(m['file']).read_text();assert original.count(m['old'])==1,(m['id'],original.count(m['old']))
 m['line']=original[:original.index(m['old'])].count('\n')+1
 (p/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),original.replace(m['old'],m['new'],1).splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
(p/'plan.json').write_text(json.dumps(plans,indent=2))
results=[]
def run(argv,name,env):
 start=time.monotonic()
 with (p/(name+'.log')).open('w') as f:r=subprocess.run(argv,stdout=f,stderr=subprocess.STDOUT,env=env)
 events=[]
 for line in (p/(name+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 item={'id':name,'command':argv,'exit':r.returncode,'wall':time.monotonic()-start,'failed':sorted(set(x['Test'].split('/')[0] for x in events if x.get('Test') and x.get('Action')=='fail')),'passed':sorted(set(x['Test'] for x in events if x.get('Test') and '/' not in x['Test'] and x.get('Action')=='pass')),'skipped':sorted(set(x['Test'] for x in events if x.get('Test') and x.get('Action')=='skip')),'binary_seconds':next((x.get('Elapsed') for x in events[::-1] if not x.get('Test') and x.get('Action') in ['pass','fail']),None),'fail_lines':[x.get('Output','').strip() for x in events if x.get('Test') and any(s in x.get('Output','') for s in ['want refusal','want a refusal','want a not-yet','panic:'])],'fail_events':[x for x in events if x.get('Action')=='fail']}
 results.append(item);(p/'results.json').write_text(json.dumps(results,indent=2));print(name,r.returncode,item['wall'],item['failed'],flush=True)
 return item
for m in plans:
 path=Path(m['file']);original=path.read_text();env=os.environ|{'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/defend-library-object/cache/'+m['id']}
 try:
  path.write_text(original.replace(m['old'],m['new'],1))
  vetted=run(['go','vet','./internal/lower/'],m['id']+'-vet',env)
  if vetted['exit']:break
  matrix=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],m['id'],env)
 finally:path.write_text(original)

import json,os,re,subprocess,time
from pathlib import Path
out=Path(__file__).resolve().parent
results=json.loads((out/'checks.json').read_text())
def run(label,command,expected=0,minimum=1):
 start=time.monotonic();log=out/(label+'.log')
 with log.open('w') as stream:
  completed=subprocess.run(command,stdout=stream,stderr=subprocess.STDOUT,timeout=240)
 result=dict(label=label,command=command,exit=completed.returncode,seconds=round(time.monotonic()-start,3))
 results.append(result);(out/'checks.json').write_text(json.dumps(results,indent=2))
 print(label,result['exit'],result['seconds'],flush=True)
 text=log.read_text();events=[]
 for line in text.splitlines():
  if line.startswith('{'):
   try:events.append(json.loads(line))
   except json.JSONDecodeError:pass
 tests=[r for r in events if r.get('Action') in ['pass','fail'] and 'Test' in r]
 assert completed.returncode==expected,(label,completed.returncode)
 assert len(tests)>=minimum,(label,'matched no tests')
 assert not any(r.get('Elapsed',0)>=60 for r in tests),(label,'test crossed 60 seconds')
 if expected==1:
  assert '[build failed]' not in text and any(r['Action']=='fail' for r in tests),(label,'not a test catcher')
 return tests

standard=['-count=1','-timeout','90s','-parallel','4','-json']
audit=json.loads((out/'stage3-audit.json').read_text())
targets=[audit['diagnostic_updates'][0],audit['pending_native_proof'][0]]
records=[]
for item in targets:
 p=Path(item['manifest']);original=p.read_text();decoder=json.JSONDecoder();cursor=0;mutated=None
 while cursor<len(original):
  if original[cursor]!='{':cursor+=1;continue
  entry,end=decoder.raw_decode(original,cursor)
  if entry.get('file')==item['file']:
   raw=original[cursor:end];match=re.search(r'"stage0"\s*:\s*',raw);_,value_end=decoder.raw_decode(raw,match.end());mutated=original[:cursor+match.end()]+json.dumps(item['before'],separators=(',',':'))+original[cursor+value_end:];break
  cursor=end
 assert mutated is not None and mutated!=original
 label='mutant-record-'+Path(item['manifest']).parent.name
 try:
  p.write_text(mutated)
  bucket=p.parent.name;test='TestFixtures'+bucket.title()
  run(label,['go','test','./stage3/fixtures','-run','^'+test+'$/'+bucket+'$/'+re.escape(item['file'])+'$/stage0$',*standard],1)
  assert 'gap changed' in (out/(label+'.log')).read_text()
  records.append({'manifest':str(p),'file':item['file'],'mutation':'restore the stale main stage0 record','caught_by':'exact stage0 record comparison','exit':1})
 finally:p.write_text(original)
run('stage3-records-restored',['go','test','./stage3/fixtures','-run','^TestFixtures(Assertions|Namespaces)$',*standard])
(out/'record-mutants.json').write_text(json.dumps(records,indent=2))

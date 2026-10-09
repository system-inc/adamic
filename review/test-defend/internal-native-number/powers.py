import pathlib,subprocess,os,json,time
r=pathlib.Path('/workspace/adamic');o=r/'review/test-defend/internal-native-number';p=r/'internal/native/runtime/dtoa.c';s=p.read_text();results=json.loads((o/'results.json').read_text());plan=json.loads((o/'plan.json').read_text())
try:
 for result in results:
  mid=result['id']; edit=next((x for x in plan if x['id']==mid),None);p.write_text(s.replace(edit['old'],edit['new']) if edit else s)
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-number/cache/'+mid
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^TestPowersOfTwoFormatAsJavaScriptDoes$'];start=time.monotonic()
  with (o/(mid+'-powers.log')).open('w') as f:rc=subprocess.run(cmd,cwd=r,env=env,stdout=f,stderr=subprocess.STDOUT).returncode
  result['powers_seconds']=time.monotonic()-start;result['powers_command']=' '.join(cmd)
  events=[json.loads(x) for x in (o/(mid+'-powers.log')).read_text().splitlines() if x.startswith('{')]
  for e in events:
   if e.get('Test')=='TestPowersOfTwoFormatAsJavaScriptDoes' and e.get('Action') in ('pass','fail'):result['rows_'+('passed' if e['Action']=='pass' else 'failed')].append(e['Test'])
  (o/'results.json').write_text(json.dumps(results,indent=2))
finally:p.write_text(s)
rows=json.loads((o/'matrix-rows.json').read_text());rows.append('TestPowersOfTwoFormatAsJavaScriptDoes');(o/'matrix-rows.json').write_text(json.dumps(rows,indent=2))

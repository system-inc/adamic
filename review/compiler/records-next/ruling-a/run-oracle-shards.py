import subprocess,pathlib,json,re,os
root=pathlib.Path('/workspace/adamic');out=root/'review/compiler/records-next/ruling-a';selection=json.loads((out/'oracle-selection.json').read_text());results=[]
groups=[('fixtures-'+str(i+1), '^TestNativeAgreesWithNode$/^('+ '|'.join(re.escape(x) for x in selection['fixtures'][i::8])+')$') for i in range(8)]
groups += [('mutants-'+str(i+1),'^('+'|'.join(selection['tests'][i::4])+')$') for i in range(4)]
groups += [('entries','^TestEntries(Acceptance|Provenance|RuntimeReadiness)$'),('buckets','^TestRecordCensusComparisonBuckets$')]
for name,pattern in groups:
 command=['/workspace/records-next-scratch/records-ruling-oracle.test','-test.run',pattern,'-test.count=1','-test.timeout=85s','-test.v']
 with (out/(name+'.log')).open('w') as log:
  try: code=subprocess.run(command,cwd=root/'internal/oracle',env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=log,stderr=subprocess.STDOUT,timeout=89).returncode
  except subprocess.TimeoutExpired: code=124
 results.append({'name':name,'pattern':pattern,'exit':code});(out/'oracle-shards.json').write_text(json.dumps(results,indent=2)+'\n');print(name,'exit',code,flush=True)
assert all(x['exit']==0 for x in results)

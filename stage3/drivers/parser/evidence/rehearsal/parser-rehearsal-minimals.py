from pathlib import Path
import subprocess,json,os
out=Path('/workspace/scratch/parser-rehearsal-minimals');out.mkdir();root=Path('/workspace/adamic/stage3/drivers/parser');scratch=Path('/workspace/scratch/parser-rehearsal-main');rows=[]
def run(command,stem):
 with (out/(stem+'.stdout')).open('wb') as stdout,(out/(stem+'.stderr')).open('wb') as stderr:

  try: code=subprocess.run(command,cwd=scratch,stdout=stdout,stderr=stderr,timeout=90).returncode
  except subprocess.TimeoutExpired: code=124
 return {'command':command,'exit':code,'stdout':(out/(stem+'.stdout')).read_text(),'stderr':(out/(stem+'.stderr')).read_text()}
for name in ['native-error-capture-stack.a','native-namespace-object-receiver.a','native-namespace-class.a','native-callable-namespace.a','native-enum-map.a','native-call-before-enum.a','native-maplike-index.a','native-nonnull.a','native-predicate-overload.a','native-predicate-callback.a','native-same-map-return-cast.a','native-key-array-cast.a','native-arguments-length.a','native-arguments-length-value.a','native-generic-array-cast.a','native-generic-empty-array.a','native-sorted-array-brand.a','native-predicate-callback-parameter.a','native-some-predicate-overload.a','native-debug-namespace.a','native-map-before-namespace.a']:
 row={'file':name};binary=out/(name+'.bin')
 row['node']=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(root/name)],name+'.node')
 row['build']=run(['/workspace/scratch/parser-rehearsal-adamic','build',str(root/name),'-o',str(binary)],name+'.build')
 row['pass']=False
 if row['build']['exit']==0:
  row['native']=run([str(binary)],name+'.native')
  row['comparison']={stream:run(['cmp',str(out/(name+'.node.'+stream)),str(out/(name+'.native.'+stream))],name+'.'+stream+'-cmp')['exit'] for stream in ['stdout','stderr']}
  row['pass']=row['node']['exit']==row['native']['exit']==0 and all(x==0 for x in row['comparison'].values())
  row['binary_bytes']=binary.stat().st_size
  if row['pass']:
   data=(out/(name+'.native.stdout')).read_bytes();assert data
   mutant=out/(name+'.mutant.stdout');mutant.write_bytes(bytes([data[0]^1])+data[1:])
   row['native_byte_mutant_cmp']=run(['cmp',str(out/(name+'.node.stdout')),str(mutant)],name+'.mutant-cmp')['exit'];assert row['native_byte_mutant_cmp']==1
 rows.append(row)
report={'probes':rows,'passing':sum(r['pass'] for r in rows),'total':len(rows),'full_slice':'Source-stop minimal witnesses; native comparison and byte mutants only if any builds'}
(out/'report.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report,indent=2))

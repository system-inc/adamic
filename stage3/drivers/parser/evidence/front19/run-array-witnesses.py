from pathlib import Path
import subprocess,json,os
out=Path('/workspace/scratch/parser-front19-array-witnesses-probes');out.mkdir();root=Path('/workspace/adamic/stage3/drivers/parser');scratch=Path('/tmp/parser-front7-scratch');rows=[]
def run(command,stem):
 with (out/(stem+'.stdout')).open('wb') as stdout,(out/(stem+'.stderr')).open('wb') as stderr:
  code=subprocess.run(command,cwd=scratch,stdout=stdout,stderr=stderr).returncode
 return {'command':command,'exit':code,'stdout':(out/(stem+'.stdout')).read_text(),'stderr':(out/(stem+'.stderr')).read_text()}
for name in ['native-same-map-return-cast.a','native-sorted-empty-conditional.a','native-map-before-namespace.a']:
 row={'file':name};binary=out/(name+'.bin')
 row['node']=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(root/name)],name+'.node')
 row['build']=run(['/workspace/scratch/parser-front19-adamic','build',str(root/name),'-o',str(binary)],name+'.build')
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
report={'probes':rows,'passing':sum(r['pass'] for r in rows),'total':len(rows),'full_slice':'Validated slice first stop core.ts:11:52; standalone witnesses only'}
(out/'report.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report,indent=2))

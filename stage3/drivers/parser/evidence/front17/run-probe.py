from pathlib import Path
import subprocess,json
out=Path('/workspace/scratch/parser-front17-probe');out.mkdir();scratch=Path('/tmp/parser-front7-scratch');source='/workspace/adamic/stage3/drivers/parser/native-array-is-array-predicate.a';binary=out/'is-array';report={'source':source,'signature':'isArray(value: unknown): value is readonly unknown[]','full_slice':'held for lane 2; no full rerun'}
def run(command,stem):
 with (out/(stem+'.stdout')).open('wb') as stdout,(out/(stem+'.stderr')).open('wb') as stderr:code=subprocess.run(command,cwd=scratch,stdout=stdout,stderr=stderr).returncode
 return {'command':command,'exit':code,'stdout':(out/(stem+'.stdout')).read_text(),'stderr':(out/(stem+'.stderr')).read_text()}
report['node_original']=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs','/workspace/scratch/parser-front17-original-array-predicate.a'],'node-original')
report['node']=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',source],'node')
report['signature_node_cmp']={s:run(['cmp',str(out/('node-original.'+s)),str(out/('node.'+s))],'signature-'+s+'-cmp')['exit'] for s in ['stdout','stderr']}
report['build']=run(['/workspace/scratch/parser-front17-adamic','build',source,'-o',str(binary)],'build');report['pass']=False
if report['build']['exit']==0:
 report['native']=run([str(binary)],'native');report['binary_bytes']=binary.stat().st_size
 report['comparison']={s:run(['cmp',str(out/('node.'+s)),str(out/('native.'+s))],s+'-cmp')['exit'] for s in ['stdout','stderr']}
 report['pass']=report['node']['exit']==report['native']['exit']==0 and all(x==0 for x in report['comparison'].values())
 if report['pass']:
  data=(out/'native.stdout').read_bytes();assert data;(out/'mutant.stdout').write_bytes(bytes([data[0]^1])+data[1:])
  report['native_byte_mutant_cmp']=run(['cmp',str(out/'node.stdout'),str(out/'mutant.stdout')],'mutant-cmp')['exit'];assert report['native_byte_mutant_cmp']==1
(out/'report.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report,indent=2))

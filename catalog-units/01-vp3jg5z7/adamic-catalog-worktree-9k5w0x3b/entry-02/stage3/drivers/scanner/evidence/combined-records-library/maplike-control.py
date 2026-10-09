from pathlib import Path
import subprocess,json
r=Path('/workspace/scratch/native3-combined-probes')
rows=json.loads((r/'results.json').read_text());row=rows[0];assert row['name']=='01-index' and row['node_exit']==row['build_exit']==0
with (r/'01-index.native.stdout').open('wb') as out,(r/'01-index.native.stderr').open('wb') as err:code=subprocess.run([str(r/'01-index.native')],stdout=out,stderr=err).returncode
assert code==0 and (r/'01-index.native.stdout').read_bytes()==(r/'01-index.node.stdout').read_bytes() and not (r/'01-index.native.stderr').read_bytes()
original=(r/'01-index.native.stdout').read_bytes();assert original==b'ok\n';(r/'01-index.mutant.stdout').write_bytes(b'n'+original[1:])
results={}
for name,left,right in [('control',r/'01-index.node.stdout',r/'01-index.native.stdout'),('mutant',r/'01-index.node.stdout',r/'01-index.mutant.stdout')]:
 with (r/('01-index.'+name+'.diff')).open('wb') as out:results[name]=subprocess.run(['diff','-u',str(left),str(right)],stdout=out).returncode
assert results=={'control':0,'mutant':1}
(r/'01-index.control.json').write_text(json.dumps({'node_exit':0,'native_build_exit':0,'native_exit':0,'native_stdout':'ok\n','diff_exit':0,'one_byte_mutant_diff_exit':1,'scope':'type-only MapLike admission control, not native scanner'},indent=2)+'\n')
print('PASS: unchanged type-only MapLike probe prints ok on Node and native; diff 0, one-byte native output mutant diff 1.')

"""Compare the old complete manifest with the union of new file manifests."""
import hashlib, json, os, subprocess
from pathlib import Path
root=Path(__file__).resolve().parents[4]
out=root/'review/compiler/test-split-bridge-main'
cache=Path('/workspace/bridge-main-products')
def product(name):
 matches=[p.with_suffix('')/name for p in cache.glob('*.inputs') if p.read_text().splitlines()[0]=='name bridge-test-'+name]
 assert matches,name
 return str(max(matches,key=lambda p:p.stat().st_mtime))
scratch=Path('/workspace/bridge-main-union'); scratch.mkdir(exist_ok=True)
results=[]
for mode,roots in [('sample',[root/'bridge/tsgo/testdata/sample.ts']),('corpus',[Path('/workspace/test-split-typescript/src/compiler')/name for name in ['checker.ts','parser.ts','types.ts','utilities.ts']])]:
 parts=[]
 for path in roots:
  data=path.read_bytes(); count=min(400,len(data))
  parts.append(''.join(str(path)+'\t'+str(i*len(data)//count)+'\n' for i in range(count)))
 def answers(name,manifest,label):
  path=scratch/(label+'.tsv'); path.write_text(manifest)
  command=[product(name),str(root/'bridge/tsgo/testdata/tsconfig.json'),str(path),*[str(p) for p in roots]]
  with (out/('union-'+label+'-'+name+'.stderr')).open('w') as log:
   result=subprocess.run(command,cwd=root,stdout=subprocess.PIPE,stderr=log)
  assert result.returncode==0
  return result.stdout
 whole=answers('oracle',''.join(parts),mode+'-whole')
 split=b''.join(answers('oracle',part,mode+'-'+str(i)) for i,part in enumerate(parts))
 native=answers('native-asan',''.join(parts),mode+'-whole')
 assert whole==split==native
 results.append(dict(mode=mode,positions=sum(p.count('\n') for p in parts),roots=[str(p) for p in roots],source_sha256={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in roots},answer_bytes=len(whole),answer_sha256=hashlib.sha256(whole).hexdigest(),whole_equals_split=True,whole_equals_native=True))
(out/'union.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results,indent=2))

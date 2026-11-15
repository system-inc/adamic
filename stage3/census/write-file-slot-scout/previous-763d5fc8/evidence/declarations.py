from pathlib import Path
import json,subprocess,hashlib
r=Path('/workspace/scratch/write-file-slot-run')
results=[]
for label in ['main','a','b']:
 config=json.loads((r/(label+'-tsconfig.json')).read_text())
 config['compilerOptions'].update(noEmit=False,emitDeclarationOnly=True,declaration=True,stripInternal=True,rootDir=str(r/label/'src/compiler'),outDir=str(r/(label+'-public-dts')))
 p=r/(label+'-dtsconfig.json');p.write_text(json.dumps(config))
 with (r/(label+'-dts.log')).open('w') as log:
  result=subprocess.run(['node','/workspace/scratch/write-file-slot-scout/stage3/api/node_modules/typescript/lib/tsc.js','-p',str(p)],stdout=log,stderr=subprocess.STDOUT)
 results.append({'label':label,'exit':result.returncode})
main={str(p.relative_to(r/'main-public-dts')):hashlib.sha256(p.read_bytes()).hexdigest() for p in (r/'main-public-dts').rglob('*.d.ts')}
for result in results:
 label=result['label']; hashes={str(p.relative_to(r/(label+'-public-dts'))):hashlib.sha256(p.read_bytes()).hexdigest() for p in (r/(label+'-public-dts')).rglob('*.d.ts')}
 result.update(files=len(hashes),changed=[n for n in sorted(set(main)|set(hashes)) if main.get(n)!=hashes.get(n)])
(r/'declaration-emits.json').write_text(json.dumps(results,indent=2))

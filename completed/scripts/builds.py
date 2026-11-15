from pathlib import Path
import subprocess as s,time,shutil,tarfile,io,os,json
r=Path(__file__).resolve().parent;src=Path('/workspace/scratch/watch-file-typescript');repo=Path('/workspace/scratch/watch-file-1532-merge')
for side in ['before','after']:
 while not (r/(side+'-apply.exit')).exists():time.sleep(1)
 assert (r/(side+'-apply.exit')).read_text().strip()=='0'
builds={};root=r/'builds';root.mkdir()
variants=[('builtin-before','after','93-builtin-imports','before'),('builtin-after','after','93-builtin-imports','after'),('crypto-before','after','99-wasm-crypto-import','before'),('baseline','before','99-wasm-crypto-import','after'),('candidate','after','99-wasm-crypto-import','after')]
for label,side,name,state in variants:
 out=root/label;out.mkdir()
 for item in ['package.json','package-lock.json','Herebyfile.mjs','ThirdPartyNoticeText.txt','LICENSE.txt']:shutil.copy2(src/item,out/item)
 shutil.copytree(src/'scripts',out/'scripts')
 data=s.check_output(['git','-C',str(r/side),'archive','refs/stage3/'+name+'-'+state,'src'])
 with tarfile.open(fileobj=io.BytesIO(data)) as t:t.extractall(out,filter='data')
 (out/'node_modules').symlink_to(r/'dependencies/node_modules',target_is_directory=True)
 builds[label]=str(out)
(r/'builds.json').write_text(json.dumps(builds,indent=2)+'\n')
fail=[]
for label,tree in builds.items():
 code=s.run(['python3',str(r/'runproof.py'),'build-'+label,'0','npm','run','build:compiler'],cwd=tree).returncode
 if code:fail.append(label)
(r/'build-results.json').write_text(json.dumps({'failures':fail},indent=2)+'\n');print('BUILD FAILURES',fail,flush=True)

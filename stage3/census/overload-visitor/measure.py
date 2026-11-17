import os, pathlib, subprocess, json, time, sys
scratch=pathlib.Path('/workspace/overload-visitor-scratch')
probe=scratch/'buildcache/local/f41b91c97d420fc9eb22166e8bd071817a2199431b4480a919fd6657089e0f5e/probe'
repo=scratch/'adamic'; name=sys.argv[1];tree=scratch/name
out=repo/'stage3/census/overload-visitor'/name;out.mkdir(parents=True,exist_ok=True)
for file in ['visitorPublic.ts','utilitiesPublic.ts'] if name=='baseline' else ['visitorPublic.ts']:
 start=time.monotonic()
 with open(out/(file+'.scan.json'),'w') as f,open(out/(file+'.scan.stderr'),'w') as e:
  p=subprocess.run([str(probe),str(tree/'src/compiler'/file)],stdout=f,stderr=e)
 meta={'exit':p.returncode,'seconds':time.monotonic()-start,'root':str(tree/'src/compiler'/file)}
 (out/(file+'.scan.meta.json')).write_text(json.dumps(meta,indent=2)+'\n')
 print(name,file,meta,flush=True)
 if p.returncode:sys.exit(p.returncode)

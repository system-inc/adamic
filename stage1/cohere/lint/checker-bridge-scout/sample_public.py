import json,pathlib,subprocess,re
import argparse
parser=argparse.ArgumentParser(description='Fetch pinned public sources only; never install dependencies or run repository scripts.')
parser.add_argument('directory',type=pathlib.Path)
base=parser.parse_args().directory.resolve();samples=base/'samples';samples.mkdir(exist_ok=True)
records=[]
for repo in json.loads((base/'pins.json').read_text()):
 if 'error' in repo: raise RuntimeError('required public pin failed: '+repo['repo'])
 root=pathlib.Path(repo['root']); sources=json.loads((root/'sources.json').read_text())
 candidates=[p for p in sources if not any(x in p for x in ['/__tests__/','/tests/','/test/','.test.','.spec.','fixtures/'])]
 if repo['repo']=='microsoft/TypeScript':candidates=['src/compiler/core.ts' if 'src/compiler/core.ts' in sources else 'tsc/testdata/fixtures/compiler/core.ts']
 fallback=None;selected=None
 for path in candidates[:40]:
  data=subprocess.check_output(['git','-C',str(root),'-c','remote.origin.url=https://github.com/'+repo['repo']+'.git','show','FETCH_HEAD:'+path])
  if len(data)>40000:continue
  if fallback is None:fallback=(path,data)
  if re.search(rb'(===|!==|==|!=)\s*(true|false)\b',data): selected=(path,data);break
 if selected is None:selected=fallback
 if selected is None:
  path=candidates[0];selected=(path,subprocess.check_output(['git','-C',str(root),'-c','remote.origin.url=https://github.com/'+repo['repo']+'.git','show','FETCH_HEAD:'+path]))
 path,data=selected;destination=samples/root.name/path;destination.parent.mkdir(parents=True,exist_ok=True);destination.write_bytes(data)
 records.append(dict(repo=repo['repo'],sha=repo['sha'],path=path,local=str(destination),bytes=len(data)))
 print(repo['repo'],path,len(data),flush=True)
(samples/'fixtures.json').write_text(json.dumps(records,indent=2)+'\n')
(samples/'manifest').write_text(''.join(r['local']+'\n' for r in records))
(samples/'tsconfig.json').write_text(json.dumps(dict(compilerOptions=dict(strict=True,target='ESNext',noResolve=True),files=[r['local'] for r in records]),indent=2)+'\n')

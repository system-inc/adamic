import concurrent.futures,json,subprocess
from pathlib import Path
pins='''actualbudget/actual 9732a4463aac2909ac2aad1627085e0eb7a207b5
angular/angular c0dc8c4bbeea70879aef54e9fcc7888359dfd1a5
babel/babel f67453d563918a1ea4a1dcb1d7af01d4a528d99b
backstage/backstage 532931240eec03b318efaadb39e24d9ee6d7076c
calcom/cal.diy 54343aa685ae8f33159d2f485ec4a57bad5c574a
date-fns/date-fns 717ce0a807ea4c6b540d015b5c408723175b2838
excalidraw/excalidraw ed10ac7dca7e40f3f4a31269b4bfba980d0db41e
grafana/grafana c7a7b797c1980a886efdd01b433a8d92b7ae7166
microsoft/playwright 2a8ba77a33a5a39a52372c42f12d254506296c76
n8n-io/n8n e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073
nestjs/nest 35142c3eca8edaaf6abc5984d915da2fbd458aa2
outline/outline 478e8121cbe517b9d7773e9d68d20bfe26056c59
prisma/prisma c882b03377e70c090d7bbe05cf85bf04fe9e33b2
shadcn-ui/ui a2e305c2e15b6affdb760e65058be69e8212713f
supabase/supabase 87681812a0b4aef538d18ea78a78e5ab6257952e
TanStack/query eaa75f4f8f819237febca9f7e887455367b7ab97
tldraw/tldraw db1c86ea7857483c47aa333cf4e92abf455a67cf
trpc/trpc d756e591a5e37ef20b8d75ecd4d736c195497289
twentyhq/twenty ddd166250b03e4cb262766034e8ef3e444b516c0
typeorm/typeorm c64a1f052fc39f6688b6b73b83d065d7147ba8bb
microsoft/TypeScript 50d70a3f5f453a79a4323b263165da51f656a4e3
microsoft/TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8
vuejs/core 4ab865a848a1da3d10fb674f857e5fff13094644'''
root=Path('/tmp/formatter-public');root.mkdir(exist_ok=True)
def fetch(line):
 repo,sha=line.split();folder=root/(repo.replace('/','__')+'-'+sha[:8]);folder.mkdir(exist_ok=True)
 def run(args): return subprocess.run(args,cwd=folder,text=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE,timeout=600)
 run(['git','init','-q']);run(['git','remote','add','origin','https://github.com/'+repo+'.git'])
 result=run(['git','fetch','--depth','1','origin',sha])
 if result.returncode: return dict(repo=repo,pin=sha,error=result.stderr)
 result=run(['git','checkout','--detach',sha])
 if result.returncode: return dict(repo=repo,pin=sha,error=result.stderr)
 paths=run(['git','ls-files']).stdout.splitlines();files=[p for p in paths if p.endswith(('.ts','.tsx'))]
 sparse=run(['git','sparse-checkout','set','--no-cone','*.ts','*.tsx'])
 if sparse.returncode:return dict(repo=repo,pin=sha,error=sparse.stderr)
 print(repo,sha,len(files),flush=True)
 return dict(repo=repo,pin=sha,checkout=str(folder),files=files)
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
 results=list(pool.map(fetch,pins.splitlines()))
(root/'manifest.json').write_text(json.dumps(results,indent=2)+'\n')
print('manifest ready',flush=True)

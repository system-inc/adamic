from pathlib import Path
import json
p=Path(__file__).resolve().parent/'raw';pairs=0;changed=[]
for side in ['base','head']:
 rows=[json.loads(f.read_text()) for f in p.glob(side+'-root-*-status.json')]
 print(side,len(rows),'complete',sum(r['exit_code']!=0 for r in rows),'failed')
for h in p.glob('head-root-*.jsonl'):
 b=p/h.name.replace('head-','base-');hl=h.read_text().splitlines();bl=b.read_text().splitlines() if b.exists() else []
 if len(hl)!=2 or len(bl)!=2:continue
 x,y=json.loads(bl[0]),json.loads(hl[0]);pairs+=1
 if tuple(x.get(k) for k in ['where','kind','reason'])!=tuple(y.get(k) for k in ['where','kind','reason']):changed.append(h.name)
print('paired',pairs,'changed',changed)
print('checkpoints',*[f.name.removesuffix('-status.json') for f in sorted(p.glob('*checkpoint-group*-status.json')) if json.loads(f.read_text())['exit_code']==0])

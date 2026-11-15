from pathlib import Path
import json,re
out=Path(__file__).resolve().parent;data={};pairs=[]
for which in ['main','head']:
 data[which]={}
 for p in (out/'evidence'/which).glob('root-*.time.json'):
  if json.loads(p.read_text())['exit_code']!=0:continue
  obs=json.loads(p.with_name(p.name.replace('.time.json','.json')).read_text());obs.pop('seconds',None);obs.pop('roots',None)
  for k in ['where','reason','fix']:
   if k in obs:obs[k]=obs[k].replace('/workspace/loose-slot-scratch/adapted-'+which+'/src/compiler/','')
  data[which][p.name.split('.')[0]]=obs
for key in sorted(data['main'].keys()&data['head'].keys()):pairs.append((key,data['main'][key]==data['head'][key]))
r={'completed_by_sha':{k:len(v)for k,v in data.items()},'compared_pairs':len(pairs),'unchanged_pairs':sum(same for _,same in pairs),'changed_pairs':[key for key,same in pairs if not same]}
(out/'evidence/progress.json').write_text(json.dumps(r,indent=2)+'\n')
p=out/'REPORT.md';t=p.read_text();block=f'<!--ROOT-PROGRESS-->\nCompleted pairs so far: **{len(pairs)}/79**, **{r["unchanged_pairs"]} unchanged**, **{len(r["changed_pairs"])} changed**. Remaining roots are unmeasured, not assumed unchanged.\n<!--END-ROOT-PROGRESS-->'
if '<!--ROOT-PROGRESS-->' in t:t=re.sub(r'<!--ROOT-PROGRESS-->.*?<!--END-ROOT-PROGRESS-->',block,t,flags=re.S)
else:t=t.replace('## Production next stops','## Production next stops\n\n'+block)
p.write_text(t);print(r)

from pathlib import Path
import hashlib,json
s=Path('/workspace/scratch/getfunctionname');fresh=s/'tree';product=Path('/workspace/scratch/debug-messages/after');names={'typescript.d.ts','tsserverlibrary.d.ts','typescript.internal.d.ts'};files=[]
for root,prefix in [(fresh,'checkout'),(product/'built/local','built/local')]:
 for p in sorted(root.rglob('*')):
  if p.name not in names or not p.is_file():continue
  rel=p.relative_to(root)
  if rel.parts[0] in ['node_modules','.git','built']:continue
  files.append((prefix+'/'+str(rel),p))
reference=fresh/'tests/baselines/reference/api/typescript.d.ts';assert reference.exists();assert any(p==reference for _,p in files)
rows=[]
for label,p in files:
 lines=p.read_text().splitlines();matches=[{'line':i+1,'text':text} for i,text in enumerate(lines) if 'getFunctionName' in text]
 rows.append({'file':label,'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'matches':matches})
(s/'raw/declarations.json').write_text(json.dumps({'search':'Exact spelling getFunctionName, in every checkout file with one of the three declaration basenames, plus all three actual emitted declarations. node_modules, .git and measurement output excluded.','files':rows,'publicApiBaseline':next(r for r in rows if r['file']=='checkout/tests/baselines/reference/api/typescript.d.ts')},indent=2)+'\n')
print(json.dumps(rows,indent=2))

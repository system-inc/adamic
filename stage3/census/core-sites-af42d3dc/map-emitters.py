"""Relocate the previous report's exact emitter lines in the pinned compiler."""
import json,re
from pathlib import Path
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent
old=json.loads((out/'previous-emitters.json').read_text());rows=json.loads((out/'rows.json').read_text());emitters={};unmapped=[]
def resolve(entry):
 p=repo/entry['file'];lines=p.read_text().splitlines();needle=entry['source_line']
 hits=[i for i,line in enumerate(lines) if line==needle]
 hits=[i for i in hits if next((line for line in reversed(lines[:i+1]) if line.startswith('func ')),None)==entry['function_declaration']]
 if len(hits)!=1:raise ValueError((entry,hits))
 n=hits[0];decl=max(i for i,line in enumerate(lines[:n+1]) if line.startswith('func '))
 return dict(entry,line=n+1,source_line=lines[n],function_declaration_line=decl+1)
for f in rows['sites']:
 key=f['phase']+'|'+f['kind']+'|'+f['reason'];entries=old.get(key)
 if f['reason'].startswith('the host binding process, pending '):
  oldsites=json.loads((out/'previous-rows.json').read_text())['sites']
  match=next((s for s in oldsites if s['phase']==f['phase'] and s['site_unit_name']==f['site_unit_name'] and s['reason']==f['reason']),None)
  if match:entries=old[match['id']]
 if entries:
  try:emitters[f['id']]=[resolve(e) for e in entries]
  except ValueError as e:unmapped.append({'site':f,'error':str(e)})
 else:unmapped.append({'site':f})
(out/'emitters.json').write_text(json.dumps(emitters,indent=2)+'\n');(out/'unmapped-emitters.json').write_text(json.dumps(unmapped,indent=2)+'\n')
print('mapped',len(emitters),'unmapped',len(unmapped))
for u in unmapped:print(u['site']['id'],u['site']['reason'],u.get('error',''))

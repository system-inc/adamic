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
def pinned(file,line,function,note=''):
 lines=(repo/file).read_text().splitlines();n=line-1;decl=max(i for i,t in enumerate(lines[:n+1]) if t.startswith('func '))
 return dict(file=file,line=line,function=function,note=note,source_line=lines[n],function_declaration_line=decl+1,function_declaration=lines[decl])
for f in rows['sites']:
 key=f['phase']+'|'+f['kind']+'|'+f['reason'];entries=old.get(key)
 if f['reason'].startswith('the host binding process, pending '):
  oldsites=json.loads((out/'previous-rows.json').read_text())['sites']
  match=next((s for s in oldsites if s['phase']==f['phase'] and s['site_unit_name']==f['site_unit_name'] and s['reason']==f['reason']),None)
  if match:entries=old[match['id']]
 if entries is None:
  reason=f['reason']
  if reason.startswith('reading '):entries=old['lowering|NotYet|reading comparer']
  elif reason.startswith('a value of type '):entries=old['lowering|NotYet|a value of type any']
  elif reason.startswith('an array of unknown with erased element storage'):
   entries=[pinned('internal/lower/object.go',396,'(*lowering).elementType','Unknown element storage has no declared runtime representation.')]
  elif reason.startswith('an array of '):entries=[pinned('internal/lower/object.go',401,'(*lowering).elementType','The element representation is unknown or is slotless without being the boxed union representation.')]
  elif reason.startswith('a function returning '):entries=[pinned('internal/lower/functions.go',129,'(*lowering).signature','Signature return representation is unknown after clockGenericReturnsT01 fallback.')]
  elif reason=='Object.getPrototypeOf':
   entries=[pinned('internal/lower/library_object.go',20,'(*lowering).objectCallArguments','The local refused closure emits Refused; getPrototypeOf is selected by the switch at lines 25 to 26.'),pinned('internal/lower/library_object.go',26,'(*lowering).objectCallArguments','Deciding switch arm for getPrototypeOf, setPrototypeOf and create.')]
 if entries and f['reason']=='an ambient value declaration':entries=entries+[pinned('internal/lower/ambient.go',91,'(*lowering).ambientRefusal','Constructs the Refused error after the reason is selected at line 79.')]
 if entries:
  try:emitters[f['id']]=[resolve(e) for e in entries]
  except ValueError as e:unmapped.append({'site':f,'error':str(e)})
 else:unmapped.append({'site':f})
(out/'emitters.json').write_text(json.dumps(emitters,indent=2)+'\n');(out/'unmapped-emitters.json').write_text(json.dumps(unmapped,indent=2)+'\n')
print('mapped',len(emitters),'unmapped',len(unmapped))
for u in unmapped:print(u['site']['id'],u['site']['reason'],u.get('error',''))

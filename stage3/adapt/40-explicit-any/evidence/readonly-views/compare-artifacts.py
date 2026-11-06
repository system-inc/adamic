"""Verify the exact rule-owned source delta and complete normal compiler artifacts."""
import hashlib,json,subprocess,sys
from pathlib import Path
before,after,out=map(Path,sys.argv[1:]);repo=Path(__file__).resolve().parents[5]
old=json.loads(subprocess.check_output(['git','show','63e14a0b2b61cf251ad74292e42d5fd09af804b3:stage3/adapt/40-explicit-any/class-rules.json'],cwd=repo))
new=json.loads((repo/'stage3/adapt/40-explicit-any/class-rules.json').read_text())
changes=[(a,b) for key in old for a,b in zip(old[key],new[key]) if a!=b]
assert len(changes)==22
sha=lambda b:hashlib.sha256(b).hexdigest()
files=lambda root:sorted(str(p.relative_to(root)) for p in root.rglob('*') if p.is_file())
names=files(before/'src');assert names==files(after/'src');changed=[]
for name in names:
 a=(before/'src'/name).read_bytes();b=(after/'src'/name).read_bytes()
 expected=a
 for left,right in changes:
  if left['file']!='src/'+name:continue
  prefix='enumObject: ' if left['parentKind']=='Parameter' else 'new Map<' if left['id']==70 else 'ts as '
  token=(prefix+left['replacement']).encode();replacement=(prefix+right['replacement']).encode()
  assert token in expected,(name,left['id'])
  expected=expected.replace(token,replacement,1)
 assert b==expected,('unexpected source delta',name)
 if a!=b:changed.append({'file':'src/'+name,'before':sha(a),'after':sha(b)})
assert [r['file'] for r in changed]==['src/compiler/debug.ts','src/compiler/emitter.ts']
js_names=[name for name in files(before/'built/local') if name.endswith(('.js','.mjs','.cjs'))]
assert js_names==[name for name in files(after/'built/local') if name.endswith(('.js','.mjs','.cjs'))]
assert len(js_names)==9,len(js_names)
emitted=[]
for name in js_names:
 a=(before/'built/local'/name).read_bytes();b=(after/'built/local'/name).read_bytes()
 assert a==b,('emitted JavaScript differs',name)
 emitted.append({'file':name,'bytes':len(a),'sha256':sha(a)})
public=[]
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
 a=(before/'built/local'/name).read_bytes();b=(after/'built/local'/name).read_bytes()
 assert a==b,('public declaration differs: stop',name)
 public.append({'file':name,'bytes':len(a),'sha256':sha(a),'byte_identical':True,'public_lines_moved':False})
declarations=[]
for name in files(before/'built/local'):
 if not name.endswith('.d.ts'):continue
 a=(before/'built/local'/name).read_bytes();b=(after/'built/local'/name).read_bytes()
 if a==b:continue
 assert a.replace(b'enumObject: Record<string, string | number>',b'enumObject: Readonly<Record<string, string | number>>')==b,('unexplained declaration',name)
 assert a.count(b'\n')==b.count(b'\n'),('declaration lines moved',name)
 declarations.append({'file':name,'before':sha(a),'after':sha(b),'only_readonly_enum_parameter':True})
result={'status':'pass','source_files':len(names),'changed_sources':changed,'changed_rules':len(changes),'javascript_count':len(emitted),'javascript_byte_identical':True,'emitted':emitted,'public_declarations':public,'other_changed_declarations':declarations}
out.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))

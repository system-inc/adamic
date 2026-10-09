from pathlib import Path
import json,difflib
root=Path('/workspace/adamic'); ev=root/'review/test-audit/internal-unicodeproperties-alias'; src='internal/unicodeproperties/unicodeproperties.go'; tab='internal/unicodeproperties/tables.go'
original={f:(root/f).read_text() for f in (src,tab)}
for f,s in original.items():Path('/tmp/u075_'+Path(f).name).write_text(s)
menu=[
(src,'ranges[mid].End < cp','ranges[mid].End <= cp'),
(src,'ranges[lo].Start <= cp','ranges[lo].Start < cp'),
(src,'codePoint > 0x10FFFF','codePoint > 0xFFFF'),
(src,'r.Start > next','r.Start >= next'),
(src,'End: r.Start - 1','End: r.Start'),
(src,'next = r.End + 1','next = r.End'),
(src,'End: 0x10FFFF','End: 0x10FFFE'),
(src,'p.Kind != KindCodePoints','p.Kind == KindCodePoints'),
(src,'p.Kind != KindStrings','p.Kind == KindStrings'),
(src,'p.Sequences[i] >= text','p.Sequences[i] > text'),
(src,'p.Sequences[i] == text','p.Sequences[i] != text'),
(src,'if unicodeSets {','if !unicodeSets {'),
(src,'table = scriptExtensionNames','table = scriptNames'),
(src,'Value: e.value','Value: e.name'),
(src,'if expression == "" || expression == "=" {\n\t\treturn false','if expression == "" || expression == "=" {\n\t\treturn true'),
(src,'equals > 1','equals > 2'),
(src,'expression[i+1:]','expression[i:]'),
(tab,'"WSpace":                       {name: "White_Space", value: "True", set: 50}','"WSpace":                       {name: "White_Space", value: "False", set: 50}'),
(tab,'"Lu":                    {name: "General_Category", value: "Uppercase_Letter", set: 65}','"Lu":                    {name: "General_Category", value: "Lowercase_Letter", set: 65}'),
(tab,'{0x0000, 0x007F}','{0x0000, 0x007E}')]
(ev/'diffs').mkdir(exist_ok=True)
records=[]
instrument=original.copy()
for i,(f,a,b) in enumerate(menu,1):
 mid=f'M{i:02}'; s=original[f]; assert a in s,(mid,a)
 changed=s.replace(a,b,1); (ev/'diffs'/f'{mid}.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 line=s[:s.index(a)].count('\n')+1; records.append(dict(id=mid,file=f,line=line,before=a,after=b))
 # derive actual expression replacing condition/constant only, retaining syntactic surrounding statement
 old,new=a,b
 if i==15:old='false';new='true'
 elif a.startswith('if '):old=a[3:-2];new=b[3:-2]
 elif a.startswith('table = '):old=a[8:];new=b[8:]
 elif a.startswith('next = '):old=a[7:];new=b[7:]
 elif a.startswith('End: '):old=a[5:];new=b[5:]
 elif a.startswith('Value: '):old=a[7:];new=b[7:]
 elif f==tab:
  if mid=='M18':old='"True"';new='"False"'
  elif mid=='M19':old='"Uppercase_Letter"';new='"Lowercase_Letter"'
  else:old='0x007F';new='0x007E'
 target=a.replace(old,f'auditChoose("{mid}", {old}, {new})',1)
 if mid in ('M07','M20'):target=a.replace(old,f'auditChoose[uint32]("{mid}", {old}, {new})',1)
 assert a in instrument[f]; instrument[f]=instrument[f].replace(a,target,1)
# probes replace body with early return instrumentation; standalone probe drops body
probes=[('P1','func Lookup(expression string, unicodeSets bool) (Property, bool) {','return Property{}, false'),('P2','func (s *Set) Contains(codePoint rune) bool {','return false'),('P3','func (s *Set) Complement() *Set {','return &Set{}'),('P4','func (p Property) Contains(codePoint rune) bool {','return false'),('P5','func (p Property) ContainsSequence(text string) bool {','return false')]
for pid,header,ret in probes:
 instrument[src]=instrument[src].replace(header,header+'\n if auditSelected("'+pid+'") { '+ret+' }',1)
 s=original[src]; start=s.index(header); body=start+len(header); depth=1; end=body
 while depth:
  if s[end]=='{':depth+=1
  elif s[end]=='}':depth-=1
  end+=1
 changed=s[:body]+'\n\t'+ret+'\n}'+s[end:]
 if pid=='P5':changed=changed.replace('import "sort"','')
 (ev/'diffs'/f'{pid}.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+src,tofile='b/'+src)))
 records.append(dict(id=pid,file=src,line=s[:start].count('\n')+1,before=header,after=ret,probe=True))
(ev/'menu.json').write_text(json.dumps(records,indent=2))
for f,s in instrument.items():(root/f).write_text(s)
(root/'internal/unicodeproperties/audit_switch.go').write_text('package unicodeproperties\nimport "os"\nfunc auditSelected(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\nfunc auditChoose[T any](id string, normal, mutant T) T { if auditSelected(id) { return mutant }; return normal }\n')

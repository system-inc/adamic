import pathlib,json,difflib,subprocess,time,os
root=pathlib.Path('/workspace/adamic'); ev=root/'review/test-audit/stage1-cohere-lint-registry'; p=root/'stage1/cohere/lint/registry/registry.go'; original=p.read_text(); (ev/'original.go.txt').write_text(original)
menu=[
('M1','if module != "" {','if false {','flip ambiguity condition'),
('M2','decoder.DisallowUnknownFields()','// decoder.DisallowUnknownFields()','drop strict JSON option'),
('M3','names[d.Name] = true','names[d.Name] = false','change recorded name constant'),
('M4','if adapters[d.Oracle] {','if false {','flip duplicate adapter condition'),
('M5','if d.Visit == "" || len(d.Kinds) == 0 {','if d.Visit == "" && len(d.Kinds) == 0 {','change required-listener condition'),
('M6','if !identifier.MatchString(kind) || seen[kind] || !kinds[kind] {','if !identifier.MatchString(kind) || seen[kind] {','drop known-kind condition'),
('M7','if adapter.Name.Name != "main" || !functions[d.Oracle] || !functions[d.Oracle+"Options"] {','if adapter.Name.Name != "main" {','drop adapter export conditions'),
('M8','if a.Order != b.Order {','if a.Order == b.Order {','flip ordering condition'),
('M9','arguments += ", parent"','arguments += ", index"','change rendered argument constant'),
('M10','if d.Module == "" {','if d.Module != "" {','flip module fallback condition'),
('M11','err == nil && bytes.Equal(existing, file.data)','err == nil && !bytes.Equal(existing, file.data)','flip unchanged byte check'),
('M12','sort.Strings(paths)','// sort.Strings(paths)','drop witness sorting'),
('P1','func Generate(root string) ([]Descriptor, error) {','func Generate(root string) ([]Descriptor, error) {\n return nil, nil','empty Generate'),
('P2','func Render(descriptors []Descriptor) (typescript, golang []byte) {','func Render(descriptors []Descriptor) (typescript, golang []byte) {\n return nil, nil','empty Render')]
manifest=[]; switched=original
for mid,old,new,kind in menu:
 assert original.count(old)==1,(mid,original.count(old))
 line=original[:original.index(old)].count('\n')+1
 modified=original.replace(old,new)
 diff=''.join(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile='a/'+str(p.relative_to(root)),tofile='b/'+str(p.relative_to(root))))
 (ev/(mid+'.diff')).write_text(diff)
 if mid.startswith('P'):
  replacement=old+'\n if os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { return nil, nil }'
 else:
  replacement='if os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { '+new+' } else { '+old+' }' if not old.startswith('if ') else old.replace(old,'')
  if old.startswith('if '):
   condition=old[3:-2]; altered=new[3:-2]
   replacement='if (os.Getenv("ADAMIC_MUTANT") == "'+mid+'" && ('+altered+')) || (os.Getenv("ADAMIC_MUTANT") != "'+mid+'" && ('+condition+')) {'
  elif new.startswith('//'):
   replacement='if os.Getenv("ADAMIC_MUTANT") != "'+mid+'" { '+old+' }'
  elif old=='err == nil && bytes.Equal(existing, file.data)':
   replacement='err == nil && (bytes.Equal(existing, file.data) != (os.Getenv("ADAMIC_MUTANT") == "'+mid+'"))'
 switched=switched.replace(old,replacement)
 manifest.append(dict(id=mid,line=line,old=old,new=new,menu=kind))
(ev/'menu.json').write_text(json.dumps(manifest,indent=2)); (ev/'functions.txt').write_text('Named production functions reached: Generate, Discover, RuleModule, Witnesses, Render. Anonymous callbacks: Witnesses WalkDir; Discover sort.Slice comparator. All four rows call Generate. DeterministicRegeneration and AdamicRuleModule also call Render directly. No native port entry is called.\nMenu fixed before mutant results; validation tests check malformed descriptors directly, not an agreement harness.\n')
p.write_text(switched)

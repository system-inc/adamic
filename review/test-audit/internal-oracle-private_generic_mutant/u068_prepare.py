import pathlib,json,re,subprocess,difflib
p=pathlib.Path('review/test-audit/internal-oracle-private_generic_mutant');root=pathlib.Path.cwd()
menu=[('M01','internal/native/runtime/string_search_impl.h','return (double)adamic_string_units(string);','return 0;','change constant'),('M02','internal/native/runtime/string_search_impl.h','string->length - search->length + 1','string->length - search->length','off-by-one bound'),('M03','internal/native/emit_expressions.go','if identity > 0 {','if identity < 0 {','flip condition'),('M04','internal/javascript/javascript.go','if (!values.has(code))','if (values.has(code))','flip condition')]
orig={f:(root/f).read_text() for _,f,*_ in menu};metadata=[]
for id,f,a,b,kind in menu:
 assert orig[f].count(a)==1,(id,orig[f].count(a));changed=orig[f].replace(a,b)
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(orig[f].splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 metadata.append({'id':id,'file':f,'line':orig[f][:orig[f].index(a)].count('\n')+1,'before':a,'after':b,'menu':kind})
(p/'mutant-plan.json').write_text(json.dumps(metadata,indent=2))
# Conservative runtime definition list supplements measured Go reachability, including pointer returns.
inv=[]
for f in sorted(pathlib.Path('internal/native/runtime').glob('*')):
 if f.suffix not in ['.c','.h']:continue
 txt=f.read_text()
 for m in re.finditer(r'(?m)^[A-Za-z_ \t*]+?\b([A-Za-z_]\w*)\s*\([^;{}]*\)\s*\{',txt):
  if m.group(1) not in ['if','while','for','switch']:inv.append({'file':str(f),'line':txt[:m.start()].count('\n')+1,'function':m.group(1)})
(p/'runtime-functions-static-superset.json').write_text(json.dumps(inv,indent=2))
# Runtime selector cached once per executable, allowing one switched source for all selectors.
helper='''\nstatic bool audit_selected(const char *id) { static bool loaded; static const char *selected; if (!loaded) {selected=getenv("ADAMIC_MUTANT");loaded=true;} return selected!=NULL && strcmp(selected,id)==0; }\n'''
rt=root/'internal/native/runtime/string.c';rt.write_text(rt.read_text().replace('#include "string_search_impl.h"',helper+'\n#include "string_search_impl.h"'))
texts=orig.copy()
texts[menu[0][1]]=texts[menu[0][1]].replace(menu[0][2],'return audit_selected("M01") ? 0 : (double)adamic_string_units(string);').replace(menu[1][2],'(audit_selected("M02") ? (string->length - search->length) : (string->length - search->length + 1))')
texts[menu[2][1]]=texts[menu[2][1]].replace(menu[2][2],'if (os.Getenv("ADAMIC_MUTANT") == "M03" && identity < 0) || (os.Getenv("ADAMIC_MUTANT") != "M03" && identity > 0) {').replace('import (','import (\n "os"',1)
# Emitted JS helper is one fixed string, with the selected condition changed before emission.
f=menu[3][1];texts[f]=texts[f].replace('import (','import (\n "os"',1)
a='builder.WriteString("const adamicCanonical';start=texts[f].index(a);end=texts[f].index('\n',start);line=texts[f][start:end]
texts[f]=texts[f][:start]+'if os.Getenv("ADAMIC_MUTANT") == "M04" { '+line.replace(menu[3][2],menu[3][3])+' } else { '+line+' }'+texts[f][end:]
for f,text in texts.items():(root/f).write_text(text)
probes=[('P01','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','return nil, nil'),('P02','internal/native/emit.go','func C(program *ir.Program) string {','return ""'),('P03','internal/javascript/javascript.go','func JavaScript(program *ir.Program) string {','return ""')]
for id,f,entry,ret in probes:
 original=subprocess.check_output(['git','show','HEAD:'+f],text=True);assert entry in original
 # Standalone early return leaves original code unreachable. Use an always-true runtime condition for vet.
 # A drop of the complete original body yields the exact empty answer without unreachable statements.
 start=original.index(entry)+len(entry);depth=1;end=start
 while depth:
  if original[end]=='{':depth+=1
  elif original[end]=='}':depth-=1
  end+=1
 changed=original[:start]+'\n '+ret+'\n}'+original[end:]
 if f=='internal/lower/lower.go':
  # Removed body referenced imports; standalone probe retains the function but needs all surrounding package code.
  pass
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 current=(root/f).read_text();
 if '"os"' not in current:current=current.replace('import (','import (\n "os"',1)
 current=current.replace(entry,entry+'\n if os.Getenv("ADAMIC_MUTANT") == "'+id+'" { '+ret+' }')
 (root/f).write_text(current)
(p/'probe-plan.json').write_text(json.dumps(probes,indent=2))
(p/'switch.diff').write_text(subprocess.check_output(['git','diff','--','internal/native','internal/lower','internal/javascript'],text=True))

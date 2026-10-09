from pathlib import Path
import re,json,difflib
root=Path('/workspace/adamic'); out=root/'review/test-audit/stage1-cohere-yaml-scalars_match_go_shards';out.mkdir(parents=True,exist_ok=True)
seeds=['scalar_main.ts','schema_main.ts','unist_main.ts','width_main.ts','gaps/stringUnitScan.ts','gaps/arenaNodeRead.ts','gaps/numericMapLookup.ts'];seen=set();pending=[root/'stage1/cohere/yaml'/x for x in seeds]; inventory=[]
while pending:
 p=pending.pop()
 if p in seen or not p.exists():continue
 seen.add(p);s=p.read_text()
 for n,l in enumerate(s.splitlines(),1):
  if re.match(r'\s*(?:export )?function \w+|\s+(?:constructor|[A-Za-z]\w*)\([^;]*\)\s*(?::[^=]+)?\s*\{',l):inventory.append(f'{p.relative_to(root)}:{n}: {l.strip()}')
 for imp in re.findall(r"from ['\"](\.[^'\"]+)['\"]",s):pending.append((p.parent/imp).resolve())
inventory+=['internal/native/runtime/adamic.h:741: adamic_string_char_code_at (native charCodeAt path; other runtime paths not dynamically enumerated)']
(out/'functions.txt').write_text('Conservative static module closure and declared functions. Reachability is not dynamically proven for every declaration.\n'+'\n'.join(sorted(inventory))+'\n')
menu=[('M1','stage1/cohere/yaml/scalar.ts',"result.type = mode === '>' ? 'BLOCK_FOLDED' : 'BLOCK_LITERAL';","result.type = mode === '|' ? 'BLOCK_FOLDED' : 'BLOCK_LITERAL';",'change constant'),('M2','stage1/cohere/yaml/schemaPattern.ts',"minimum = quantifier === '+' ? 1 : 0;","minimum = quantifier === '+' ? 0 : 0;",'change constant'),('M3','stage1/cohere/yaml/unistContext.ts','new UnistPoint(line, column, offset)','new UnistPoint(line, column + 1, offset)','off-by-one'),('M4','internal/native/runtime/adamic.h','return (double)(unsigned char)string->bytes[(size_t)position];','return (double)(unsigned char)string->bytes[(size_t)position] + 1;','change constant')]
records=[]
for ident,file,old,new,kind in menu:
 s=(root/file).read_text();assert s.count(old)==1
 line=s[:s.index(old)].count('\n')+1
 (out/(ident+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(old,new,1).splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 records.append(dict(id=ident,file=file,line=line,old=old,new=new,kind=kind))
(out/'menu.json').write_text(json.dumps(records,indent=2));Path('/tmp/u153/originals.json').write_text(json.dumps({r['file']:(root/r['file']).read_text() for r in records}))
print(len(seen),'source files',len(inventory),'function declarations; fixed menu saved without planting')

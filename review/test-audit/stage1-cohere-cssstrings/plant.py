from pathlib import Path
import subprocess,difflib,json,sys
R=Path('/workspace/adamic');E=R/'review/test-audit/stage1-cohere-cssstrings'; port='stage1/cohere/cssstrings/'
def base(path):return subprocess.check_output(['git','show','553ad06abb069736161b2d46532d9036e92f30b3:'+path],cwd=R,text=True)
plan=[
 ('M1',port+'strings.ts','    const content = raw.slice(1, -1);','    const content = raw.slice(1);','off-by-one slice bound'),
 ('M2',port+'strings.ts','    parts.push(quote);','', 'drop closing-quote statement'),
 ('M3',port+'strings.ts','    if(quote !== 34 && quote !== 39) return -1;','    if(!(quote !== 34 && quote !== 39)) return -1;','flip quote-recognition condition'),
 ('M4',port+'strings.ts','        index = end - 1;','        index = end;','off-by-one scanner cursor'),
 ('M5',port+'main.ts','        index++;','', 'drop decode escape advance'),
 ('M6',port+'main.ts',"code === 10 ? '\\\\n'","code === 10 ? '\\\\r'",'change encoded newline constant'),
 ('M7','internal/lower/object.go','if len(arguments) != 1 {\n\t\t\treturn nil, true, l.notYet(node, "push with other than one value")','if len(arguments) > 2 {\n\t\t\treturn nil, true, l.notYet(node, "push with other than one value")','loosen push arity bound'),
 ('M8','internal/lower/object.go','if len(arguments) != 1 {\n\t\t\treturn nil, true, l.notYet(node, "push with other than one value")','if len(arguments) == 1 {\n\t\t\treturn nil, true, l.notYet(node, "push with other than one value")','flip push arity condition'),
 ('M9','internal/lower/diagnostics.go','return &NotYet{Where: l.program.Where(node), What: what}','return &NotYet{Where: l.program.Where(node), What: ""}','change returned diagnostic option'),
 ('E1',port+'strings.ts','export function adjustStrings(value: string, singleQuote: boolean): string {',"export function adjustStrings(value: string, singleQuote: boolean): string { return '';",'empty-answer probe for port'),
 ('E2','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) { return nil, nil','empty-answer probe for gap checker'),
 ('W1',port+'shards_test.go','func stringsOutputError(unit stringsUnit, side string, got, want []byte) error {','func stringsOutputError(unit stringsUnit, side string, got, want []byte) error { return nil','allowed witness comparison weakening'),
 ('W2',port+'port_test.go','if bytes.Equal(r.stdout, want.stdout) {','if true {','allowed built-in witness comparison always declares agreement'),
]
records=[]
for mid,path,old,new,menu in plan:
 s=base(path);assert s.count(old)==1,(mid,s.count(old),old)
 diff=''.join(difflib.unified_diff(s.splitlines(True),s.replace(old,new,1).splitlines(True),fromfile='a/'+path,tofile='b/'+path))
 (E/(mid+'.diff')).write_text(diff)
 records.append(dict(id=mid,file=path,line=s[:s.index(old)].count('\n')+1,before=old,after=new,menu=menu))
(E/'mutants.json').write_text(json.dumps(records,indent=2)+'\n')
if sys.argv[-1]=='plant':
 sources={path:base(path) for _,path,_,_,_ in plan}
 def replace(path,old,new):
  assert sources[path].count(old)==1,(path,old)
  sources[path]=sources[path].replace(old,new,1)
 for mid,path,old,new,menu in plan[:6]:
  if mid=='M1':replacement="    const content = auditMutant === 'M1' ? raw.slice(1) : raw.slice(1, -1);"
  elif mid=='M2':replacement="    if(auditMutant !== 'M2') parts.push(quote);"
  elif mid=='M3':replacement="    if(auditMutant === 'M3' ? !(quote !== 34 && quote !== 39) : quote !== 34 && quote !== 39) return -1;"
  elif mid=='M4':replacement="        index = auditMutant === 'M4' ? end : end - 1;"
  elif mid=='M5':replacement="        if(auditMutant !== 'M5') index++;"
  elif mid=='M6':replacement="code === 10 ? (auditMutant === 'M6' ? '\\\\r' : '\\\\n')"
  replace(path,old,replacement)
 replace(port+'strings.ts',plan[9][2],plan[9][2]+"\n    if(auditMutant === 'E1') return '';")
 # Keep the selector in strings.ts because built-in mutant copies include only these two port files.
 sources[port+'strings.ts']="import { readTextFile } from 'adamic';\nconst auditSelection = readTextFile('/tmp/adamic-u083-mutant');\nexport const auditMutant = auditSelection.kind === 'Ok' ? auditSelection.text : '';\n"+sources[port+'strings.ts']
 sources[port+'main.ts']=sources[port+'main.ts'].replace("import { adjustStrings }", "import { adjustStrings, auditMutant }")
 old=plan[6][2]
 replacement='if (len(arguments) != 1 && os.Getenv("ADAMIC_U083_MUTANT") != "M7" && os.Getenv("ADAMIC_U083_MUTANT") != "M8") || (os.Getenv("ADAMIC_U083_MUTANT") == "M7" && len(arguments) > 2) || (os.Getenv("ADAMIC_U083_MUTANT") == "M8" && len(arguments) == 1) {\n\t\t\treturn nil, true, l.notYet(node, "push with other than one value")'
 replace('internal/lower/object.go',old,replacement)
 replace('internal/lower/diagnostics.go',plan[8][2],'if os.Getenv("ADAMIC_U083_MUTANT") == "M9" { what = "" }; '+plan[8][2])
 replace('internal/lower/lower.go',plan[10][2],plan[10][2]+'\n\tif os.Getenv("ADAMIC_U083_MUTANT") == "E2" { return nil, nil }')
 for path in ['internal/lower/object.go','internal/lower/diagnostics.go','internal/lower/lower.go']:sources[path]=sources[path].replace('import (','import (\n\t"os"',1)
 replace(port+'shards_test.go',plan[11][2],plan[11][2]+'\n\tif os.Getenv("ADAMIC_U083_WITNESS") == "W1" { return nil }')
 replace(port+'port_test.go',plan[12][2],'if os.Getenv("ADAMIC_U083_WITNESS") == "W2" || bytes.Equal(r.stdout, want.stdout) {')
 for path,s in sources.items():(R/path).write_text(s)
 subprocess.run(['gofmt','-w','internal/lower/object.go','internal/lower/diagnostics.go','internal/lower/lower.go',port+'shards_test.go',port+'port_test.go'],cwd=R,check=True)
 Path('/tmp/adamic-u083-mutant').write_text('')
print(json.dumps(records,indent=2))

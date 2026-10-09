import pathlib,json,difflib,re
out=pathlib.Path('review/test-audit/stage1-cohere-css-print'); diffdir=out/'diffs'; diffdir.mkdir(exist_ok=True)
changes={
'M1':('stage1/cohere/css/print.ts','tabWidth: 2','tabWidth: 3',"tabWidth: auditMutation === 'M1' ? 3 : 2"),
'M2':('stage1/cohere/css/print_doc.ts','while(remaining >= 0)','while(remaining > 0)',"while(auditMutation === 'M2' ? remaining > 0 : remaining >= 0)"),
'M3':('stage1/cohere/mediaquery/index.ts','\tconst nodes = parseMediaList(params);',"\tif (params.length >= 0) return {kind: 'Error', message: 'audit early refusal'};\n\tconst nodes = parseMediaList(params);", "\tif (auditMutation === 'M3') return {kind: 'Error', message: 'audit early refusal'};\n\tconst nodes = parseMediaList(params);"),
'M4':('internal/native/runtime/string_append.c','\tresult->length = written;','', '\tif (getenv("ADAMIC_MUTANT") == NULL || strcmp(getenv("ADAMIC_MUTANT"), "M4") != 0) result->length = written;'),
'P1':('stage1/cohere/css/print.ts','    const parsed = compose(text, scss);',"    if(text.length >= 0) return {kind:'Formatted',text:''};\n    const parsed = compose(text, scss);", "    if(auditMutation === 'P1') return {kind:'Formatted',text:''};\n    const parsed = compose(text, scss);"),
'P2':('stage1/cohere/mediaquery/index.ts','\tconst nodes = parseMediaList(params);',"\tif(params.length >= 0) return {kind:'Ok',value:newContainer(undefined,undefined,'','',undefined,[])};\n\tconst nodes = parseMediaList(params);", "\tif(auditMutation === 'P2') return {kind:'Ok',value:newContainer(undefined,undefined,'','',undefined,[])};\n\tconst nodes = parseMediaList(params);"),
'P3':('internal/native/runtime/string_append.c','\tsize_t length = string->length, added = 0;', '\tadamic_release(string); return adamic_string_concat(0, NULL);\n\tsize_t length = string->length, added = 0;', '\tif (getenv("ADAMIC_MUTANT") != NULL && strcmp(getenv("ADAMIC_MUTANT"), "P3") == 0) { adamic_release(string); return adamic_string_concat(0, NULL); }\n\tsize_t length = string->length, added = 0;'),
'P4':('internal/native/runtime/regexp.c','bool adamic_regex_test(adamic_object *regex, adamic_string *input) {','bool adamic_regex_test(adamic_object *regex, adamic_string *input) {\n\treturn false;','bool adamic_regex_test(adamic_object *regex, adamic_string *input) {\n\tif (getenv("ADAMIC_MUTANT") != NULL && strcmp(getenv("ADAMIC_MUTANT"), "P4") == 0) return false;'),
'W1':('stage1/cohere/css/printer_shards_test.go','difference != "" {','difference != "" && false {','difference != "" && os.Getenv("ADAMIC_MUTANT") != "W1" {'),
'S1':('stage1/cohere/css/printer_shards_test.go','\t\t\treturn native.Build(code, filepath.Join(dir, "native"), native.Options{Sanitize: sanitize})','\t\t\treturn nil','\t\t\tif os.Getenv("ADAMIC_MUTANT") == "S1" { return nil }\n\t\t\treturn native.Build(code, filepath.Join(dir, "native"), native.Options{Sanitize: sanitize})'),
'S4':('stage1/cohere/css/printer_shards_test.go','func cssModeUnion(lines []string, plan []cssModeShard, count int) error {','func cssModeUnion(lines []string, plan []cssModeShard, count int) error {\n\tif count >= 0 { return nil }','func cssModeUnion(lines []string, plan []cssModeShard, count int) error {\n\tif os.Getenv("ADAMIC_MUTANT") == "S4" { return nil }'),
}
p='stage1/cohere/css/printer_shards_test.go'; s=pathlib.Path(p).read_text(); a=s.index('\t\t\toutput, err := cmd.CombinedOutput()'); b=s.index('\t\t\treturn nil',a); before=s[a:b]
changes['S2']=(p,before,'','\t\t\tif os.Getenv("ADAMIC_MUTANT") != "S2" {\n'+before+'\t\t\t}\n')
p='stage1/cohere/css/profile_test.go'
before='"expected.txt": []byte(expected)'
changes['S3']=(p,before,'"missing-expected.txt": []byte(expected)','map[bool]string{true:"missing-expected.txt",false:"expected.txt"}[os.Getenv("ADAMIC_MUTANT") == "S3"]: []byte(expected)')
original={p:pathlib.Path(p).read_text() for p,_,_,_ in changes.values()}; (out/'originals.json').write_text(json.dumps(original))
meta=[]
for ident,(p,before,after,switch) in changes.items():
 s=original[p]; assert s.count(before)==1,(ident,s.count(before))
 changed=s.replace(before,after,1)
 (diffdir/f'{ident}.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+p,tofile='b/'+p)))
 meta.append(dict(id=ident,file=p,line=s[:s.index(before)].count('\n')+1,before=before,after=after,category='production' if ident.startswith('M') else 'probe' if ident.startswith('P') else 'construction'))
(out/'mutants.json').write_text(json.dumps(meta,indent=2))
# Installation is deliberately gated until all clean individual baselines complete.
if '--apply' not in __import__('sys').argv: raise SystemExit(0)
assert not (out/'STOP-RED.txt').exists()
t=json.loads((out/'timings.json').read_text()); assert len(t)==33 and all(x['code']==0 for x in t), 'baseline incomplete or over budget'
scratch=dict(original)
for ident,(p,before,after,switch) in changes.items():
 assert scratch[p].count(before)==1, ident
 scratch[p]=scratch[p].replace(before,switch,1)
for p in ['stage1/cohere/css/print.ts','stage1/cohere/css/print_doc.ts']:
 scratch[p]="import { auditMutation } from './audit_selector.ts';\n"+scratch[p]
p='stage1/cohere/mediaquery/index.ts'; scratch[p]="import { auditMutation } from '../css/audit_selector.ts';\n"+scratch[p]
p='internal/native/runtime/string_append.c'; scratch[p]=scratch[p].replace('#include <stdint.h>','#include <stdint.h>\n#include <stdlib.h>',1)
for p,s in scratch.items(): pathlib.Path(p).write_text(s)
pathlib.Path('stage1/cohere/css/audit_selector.ts').write_text("import { readTextFile } from 'adamic';\nconst selected = readTextFile('/tmp/u079/selector');\nexport const auditMutation = selected.kind === 'Error' ? '' : selected.text.trim();\n")
pathlib.Path('/tmp/u079/selector').write_text('')

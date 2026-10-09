from pathlib import Path
import json,difflib
root=Path('/workspace/adamic');E=root/'review/test-audit/stage1-cohere-estree-syntax';menu=json.loads((E/'menu.json').read_text());original={}
for m in menu:
 f=m['file'];s=(root/f).read_text();original[f]=s;a=m['before'];b=m['after'];mid=m['id']
 if mid=='M1':replacement="if(auditMutant === 'M1' ? false : text.includes('\\ufffd'))"
 elif mid=='M4':replacement=".slice(0, auditMutant === 'M4' ? -2 : -1)"
 else:replacement="(auditMutant === '"+mid+"' ? "+b+" : "+a+")"
 assert s.count(a)==1;s=s.replace(a,replacement,1);s="import { auditMutant } from './auditSelector.ts';\n"+s
 if mid=='M1':header='export function answer(path: string, text: string): string {';s=s.replace(header,header+"\n    if(auditMutant === 'P1') { return ''; }",1)
 (root/f).write_text(s)
(root/'stage1/cohere/estree/auditSelector.ts').write_text("import { readTextFile } from 'adamic';\nconst selection = readTextFile('/tmp/u089/mutant');\nexport const auditMutant = selection.kind === 'Error' ? '' : selection.text.trim();\n")
Path('/tmp/u089/originals.json').write_text(json.dumps(original))
f='stage1/cohere/estree/pipeline.ts';s=original[f];header='export function answer(path: string, text: string): string {';start=s.index(header);changed=s[:start]+header+"\n    return '';\n}\n";(E/'diffs/P1.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))

import pathlib,json,subprocess,difflib
p=pathlib.Path('/tmp/u099/evidence');plan=json.loads((p/'plan.json').read_text())['production_mutants'];files=sorted({m['file'] for m in plan}|{'stage1/cohere/graphql/printer/main.ts'});base={f:pathlib.Path(f).read_text() for f in files}
def diff(id,f,text):
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(base[f].splitlines(True),text.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
for m in plan:diff(m['id'],m['file'],base[m['file']].replace(m['before'],m['after']))
main='stage1/cohere/graphql/printer/main.ts';pos=base[main].index('const args = programArguments();');diff('P01',main,base[main][:pos]+'function auditMain(): void { return; }\nauditMain();\n')
replacements={'M01':"    tabWidth: auditMutant === 'M01' ? 3 : 4,",'M02':"doc: node.parts[command.flat ? (auditMutant === 'M02' ? 0 : 1) : 0] ?? panic('branch'),",'M03':"this.docs.text(`${auditMutant === 'M03' ? '##' : '#'}${decoded(comment.value).trimEnd()}`)",'M04':"this.expectToken(auditMutant === 'M04' ? '<EOF>' : '<SOF>');"}
for f,text in base.items():
 if f!=main:text=text.replace("import { panic } from 'adamic';","import { readTextFile, panic } from 'adamic';",1)
 pos=text.index('\n',text.index("from 'adamic';"))+1;text=text[:pos]+"\nconst auditRead = readTextFile('/tmp/u099-mutant');\nconst auditMutant = auditRead.kind === 'Error' ? '' : auditRead.text;\n"+text[pos:]
 for m in plan:
  if m['file']==f:assert text.count(m['before'])==1;text=text.replace(m['before'],replacements[m['id']])
 if f==main:
  pos=text.index('const args = programArguments();');text=text[:pos]+"function auditMain(): void {\n    if(auditMutant === 'P01') { return; }\n"+text[pos:]+"\n}\nauditMain();\n"
 pathlib.Path(f).write_text(text)
(p/'original-source.json').write_text(json.dumps(base));(p/'switch.diff').write_text(subprocess.check_output(['git','diff','--']+files,text=True));pathlib.Path('/tmp/u099-mutant').write_text('')

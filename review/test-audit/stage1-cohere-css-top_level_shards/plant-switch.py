import pathlib,json,difflib,subprocess,os,time
p=pathlib.Path('review/test-audit/stage1-cohere-css-top_level_shards'); plan=json.loads((p/'mutant-plan.json').read_text())['mutants']; base={m['file']:pathlib.Path(m['file']).read_text() for m in plan};main='stage1/cohere/css/main.ts';base[main]=pathlib.Path(main).read_text()
def diff(id,f,new):
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(base[f].splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
for m in plan:diff(m['id'],m['file'],base[m['file']].replace(m['before'],m['after']))
body=base[main].index('const args = programArguments();');diff('P01',main,base[main][:body]+'function auditMain(): void {\n    return;\n}\nauditMain();\n')
for f,s in base.items():
 if f!=main:s=s.replace("import { panic", "import { readTextFile, panic",1)
 pos=s.index('\n',s.index("from 'adamic';"))+1
 s=s[:pos]+"\nconst auditRead = readTextFile('/tmp/u080-mutant');\nconst auditMutant = auditRead.kind === 'Error' ? '' : auditRead.text;\n"+s[pos:]
 m=next((m for m in plan if m['file']==f),None)
 if m:
  replacements={'M01':'increment: number = (auditMutant === \'M01\' ? 0 : 1)','M02':"code === 13 || code === (auditMutant === 'M02' ? 11 : 12);",'M03':"new Position(offset, low + (auditMutant === 'M03' ? 2 : 1),",'M04':"fields.set(auditMutant === 'M04' ? 'kind' : 'type', quote(node.type));"}
  assert s.count(m['before'])==1;s=s.replace(m['before'],replacements[m['id']])
 if f==main:
  pos=s.index('const args = programArguments();');s=s[:pos]+"function auditMain(): void {\n    if(auditMutant === 'P01') { return; }\n"+s[pos:]+"\n}\nauditMain();\n"
 pathlib.Path(f).write_text(s)
(p/'switch.diff').write_text(subprocess.check_output(['git','diff','--']+list(base),text=True));pathlib.Path('/tmp/u080-mutant').write_text('')
(p/'original-source.json').write_text(json.dumps(base))

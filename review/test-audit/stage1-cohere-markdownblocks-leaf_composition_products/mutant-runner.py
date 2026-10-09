from pathlib import Path
import subprocess,json,time,difflib,os
out=Path('review/test-audit/stage1-cohere-markdownblocks-leaf_composition_products')
os.environ['ADAMIC_MARKDOWNLISTS_KEEP']='/tmp/u127-witness'
root=Path('stage1/cohere/markdownblocks'); plan=[('M1','codeblocks.ts','Math.max(3, longest + 1)','Math.max(4, longest + 1)'),('M2','htmlblocks.ts','code === 32','code === 33'),('M3','root.ts','    parts.push(arena.hardline());\n    return arena.concat(parts);','    return arena.concat(parts);'),('M4','mdastCompile.ts','previous.type !== token.type','previous.type === token.type')]
original={str(root/f): (root/f).read_text() for _,f,_,_ in plan};results=[]
original[str(root/'document.ts')] = (root/'document.ts').read_text()
groups=json.loads((out/'clean-runs.json').read_text());patterns={x['group']:x['command'][-1] for x in groups}
for name in list(patterns):
 if name not in ['list-family','malformed-family']: patterns.pop(name,None)
patterns['block-layouts']='^(TestMarkdownCodeBlockLayout|TestMarkdownHTMLBlockLayout|TestMarkdownRootLayout)$'
def run(mid,name,pattern):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run',pattern]
 started=time.monotonic()
 with (out/f'{mid}-{name}.log').open('w') as log:p=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
 data=[]
 for line in (out/f'{mid}-{name}.log').read_text().splitlines():
  try:data.append(json.loads(line))
  except:pass
 results.append({'id':mid,'group':name,'command':cmd,'wall':time.monotonic()-started,'exit':p.returncode,'cooked':any('test timed out' in x.get('Output','') for x in data),'failed':[x['Test'] for x in data if x.get('Action')=='fail' and 'Test' in x], 'errors':[x for x in data if x.get('OutputType')=='error']})
 (out/'matrix-runs.json').write_text(json.dumps(results,indent=2))
 return p.returncode
try:
 for mid,f,a,b in plan:
  path=str(root/f);old=original[path];assert old.count(a)==1,(mid,old.count(a))
  new=old.replace(a,b)
  (out/f'{mid}.diff').write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+path,tofile='b/'+path)))
  helper="\nfunction auditChoice(): string { const value = readTextFile('/tmp/u127-mutant'); return value.kind === 'Ok' ? value.text.trim() : ''; }\nconst auditMutant = auditChoice();\n"
  switched=old.replace("import { panic } from 'adamic';","import { panic, readTextFile } from 'adamic';"+helper)
  if mid=='M1': switched=switched.replace("return '`'.repeat(Math.max(3, longest + 1));", "return auditMutant === 'M1' && longest < 3 ? '`'.repeat(4) : '`'.repeat(Math.max(3, longest + 1));")
  elif mid=='M2':switched=switched.replace(a,"(auditMutant === 'M2' ? code === 33 : code === 32)")
  elif mid=='M3':switched=switched.replace(a,"    if(auditMutant !== 'M3') parts.push(arena.hardline());\n    return arena.concat(parts);")
  else:switched=switched.replace(a,"(auditMutant === 'M4' ? previous.type === token.type : previous.type !== token.type)")
  if mid == 'M4': switched=switched.replace('    compile(): MdastArena {',"    compile(): MdastArena {\n        if(auditMutant === 'P2') return new MdastArena();")
  Path(path).write_text(switched)
  (out/f'{mid}-switch.txt').write_text(switched)
 old=original[str(root/'document.ts')]
 switched=old.replace("import { panic } from 'adamic';","import { panic, readTextFile } from 'adamic';"+helper).replace('    propagate(arena, root);',"    if(auditMutant === 'P1') return '';\n    propagate(arena, root);")
 (root/'document.ts').write_text(switched)
 (out/'P1-switch.txt').write_text(switched)
 Path('/tmp/u127-mutant').write_text('')
 for name,pattern in patterns.items():
  if run('CONTROL',name,pattern): raise SystemExit('neutral selector control failed: '+name)
 for mid,_,_,_ in plan:
  Path('/tmp/u127-mutant').write_text(mid)
  # Source-selected matrix, each entry runs independently so setup cannot starve following rows.
  for name,pattern in patterns.items():
   if (mid == 'M4') != (name == 'malformed-family'): continue
   run(mid,name,pattern)
 for mid in ['P1','P2']:
  Path('/tmp/u127-mutant').write_text(mid)
  for name,pattern in patterns.items():
   if (mid == 'P2') != (name == 'malformed-family'): continue
   run(mid,name,pattern)
finally:
 for p,text in original.items():Path(p).write_text(text)
 Path('/tmp/u127-mutant').unlink(missing_ok=True)

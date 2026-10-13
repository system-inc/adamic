from pathlib import Path
import json,re,shutil
root=Path.cwd(); dest=Path('/tmp/c-emission-validation')
s=(root/'internal/oracle/oracle_test.go').read_text()
s=s[s.index('var fixtures ='):s.index('\n}\n',s.index('var fixtures ='))]
rows=[]
for path,lowers in re.findall(r'\{"([^"]+)", (true|false), (?:true|false)\}',s):
 name=Path(path).name
 refuses=Path(path).suffix=='.a' and (name.startswith('non_null_refuse_') or name.startswith('non_null_possible_'))
 rows.append(dict(Name=f'oracle-{len(rows):04d}',Path=str(root/path),Lowers=lowers=='true' and not refuses))
(dest/'oracle.json').write_text(json.dumps(rows,indent=2)+'\n')
source=root/'stage1/cohere/lint'; port=dest/'yepesta-port'
shutil.copytree(source,port,dirs_exist_ok=True)
pattern=re.compile(r'''(^import\s+[^;]*?\s+from\s+)(['"])([^'"]+)(['"])''',re.M)
for p in port.rglob('*'):
 if p.suffix not in ('.ts','.a'):continue
 original=source/p.relative_to(port)
 def rewrite(m):
  target=(original.parent/m[3]).resolve()
  if m[3].startswith('.') and not target.is_relative_to(source):return m[1]+m[2]+str(target)+m[4]
  return m[0]
 p.write_text(pattern.sub(rewrite,p.read_text()))
slug='react-jsx-no-comment-textnodes'; change=json.loads((source/'rules'/slug/'mutant.json').read_text())
# Descriptors use lowercase JSON fields, just like encoding/json's case-insensitive struct mapping.
change={k.lower():v for k,v in change.items()}
p=port/'rules'/slug/change.get('file','rule.ts');text=p.read_text();assert text.count(change['from'])==1;p.write_text(text.replace(change['from'],change['to'],1))
ports=[dict(Name='port-sf-yepesta',Path=str(port/'main.ts'),Lowers=True),dict(Name='port-g-lint',Path=str(source/'main.ts'),Lowers=True),dict(Name='port-estree-scalar-edges',Path=str(root/'stage1/cohere/estree/main.ts'),Lowers=True)]
for row in ports:(dest/(row['Name']+'.json')).write_text(json.dumps([row],indent=2)+'\n')
print('oracle entries',len(rows),'emittable',sum(x['Lowers'] for x in rows),'refused',sum(not x['Lowers'] for x in rows));print('yepesta mutation',change)

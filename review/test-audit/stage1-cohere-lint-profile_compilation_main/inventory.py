from pathlib import Path
import re,json
root=Path('.').resolve();entry=root/'stage1/cohere/lint/main.ts';seen=set();todo=[entry];inventory=[];unresolved=[]
while todo:
 f=todo.pop()
 if f in seen or not f.exists():continue
 seen.add(f);s=f.read_text()
 for match in re.finditer(r'(?:from\s+|import\s+)[\'\"]([^\'\"]+)[\'\"]',s):
  name=match.group(1)
  if not name.startswith('.'):continue
  q=(f.parent/name).resolve()
  if not q.exists() and q.suffix=='.ts' and q.with_suffix('.a').exists():q=q.with_suffix('.a')
  if q.exists():todo.append(q)
  else:unresolved.append(str(q))
 for n,line in enumerate(s.splitlines(),1):
  m=re.match(r'\s*(?:(?:export|async|static|private|public|readonly)\s+)*(?:function\s+)?([A-Za-z_$][\w$]*)\s*\([^;]*',line)
  if m and (re.search(r'\bfunction\b',line) or line.startswith('    ') and not line.startswith('        ')) and m.group(1) not in ['if','for','while','switch','return','catch','super','console']:
   inventory.append({'file':str(f.relative_to(root)),'line':n,'declaration':line.strip()})
p=root/'review/test-audit/stage1-cohere-lint-profile_compilation_main';(p/'port-function-inventory.json').write_text(json.dumps(inventory,indent=2));(p/'port-import-graph.json').write_text(json.dumps({'entry':str(entry.relative_to(root)),'files':sorted(str(x.relative_to(root)) for x in seen),'unresolved':unresolved,'scope':'static transitive import graph, conservative over all registered rules; declarations listed before mutants'},indent=2))

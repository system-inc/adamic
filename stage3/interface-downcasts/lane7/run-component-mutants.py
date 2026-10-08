#!/usr/bin/env python3
"""Kill mutations at the conjunction and member-adapter seam, not source flow."""
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/views-intersection-mutants')
logs.mkdir(exist_ok=True)
c = 'internal/native/runtime/view_intersections.c'
js = 'internal/javascript/view_intersections.go'
adapter = 'internal/oracle/checked_views_intersections_test.go'
mutants = [
 ('c-skip-check', c, 'if (adamic_view_intersection_matches(value, members, count, match, context))', 'if (value != NULL || adamic_view_intersection_matches(value, members, count, match, context))'),
 ('js-skip-check', js, 'if (adamicViewIntersectionMatches(snapshot, members, match))', "if (snapshot !== undefined || adamicViewIntersectionMatches(snapshot, members, match))"),
 ('c-accept-first-shape', c, 'if (members[index] == 0 || !match(context, members[index], value)) { return false; }', 'if (members[index] != 0 && match(context, members[index], value)) { return true; }'),
 ('js-accept-first-shape', js, 'if (!member || !match(member, snapshot)) return false;', 'if (member && match(member, snapshot)) return true;'),
 ('c-drop-nested-adapter', adapter, 'return member==1 || (member==2 && p->count_kind==adamic_view_union_number);', 'return member==1 || (member==2 && true);'),
 ('js-drop-nested-adapter', adapter, "const match = (member,snapshot) => member===1 ? typeof snapshot.value.text==='string' : typeof snapshot.value.child.count === 'number';", "const match = (member,snapshot) => member===1 ? typeof snapshot.value.text==='string' : true;"),
]
for name, filename, old, new in mutants:
 path = root / filename
 original = path.read_bytes()
 assert original.decode().count(old) == 1, name
 try:
  path.write_text(original.decode().replace(old, new, 1))
  env = dict(os.environ, ADAMIC_GATE_UNCACHED='1')
  with (logs / (name+'.log')).open('w') as log:
   result = subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewIntersectionConjunction$','-count=1','-v','-timeout','10m'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  output = (logs/(name+'.log')).read_text()
  assert result.returncode != 0 and 'exit codes differ' in output and 'exitCode:0' in output, (name,output)
  assert 'error:' not in output and 'build failed' not in output, (name,output)
  print(name+': caught by exit-70 pin; valid release execution',flush=True)
 finally:
  path.write_bytes(original)
print('6 component mutants caught; source propagation remains unproven')

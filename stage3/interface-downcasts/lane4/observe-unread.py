#!/usr/bin/env python3
"""Pin the current eager-admission frontier, not a desired lazy behavior test."""
import json,pathlib,subprocess,sys
lane=pathlib.Path(__file__).resolve().parent;root=lane.parents[2];source=lane/'read-fixtures/unread-unsupported.a';rows={}
node=subprocess.run(['node','--input-type=module-typescript'],input=source.read_bytes(),capture_output=True)
assert node.returncode==0 and node.stdout==b'true\n' and not node.stderr
rows['node']={'exit':node.returncode,'stdout':node.stdout.decode(),'stderr':node.stderr.decode()}
for backend in ['c','js']:
 result=subprocess.run([sys.argv[1],backend,str(source)],cwd=root,capture_output=True)
 expected=f'adamic: {source}:6:8: stage 0 can\'t lower checked view field unread of type string | number yet\n'
 assert result.returncode==1 and result.stderr.decode()==expected,(backend,result.stderr)
 rows[backend]={'exit':result.returncode,'stdout':result.stdout.decode(),'stderr':result.stderr.decode()}
(lane/'unread-observations.json').write_text(json.dumps(rows,indent=2)+'\n')
print('OBSERVED: Node true; both backend compilers refuse an unread, valid string|number field at the cast')

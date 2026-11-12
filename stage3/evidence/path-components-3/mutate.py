import json,sys
from pathlib import Path
rule=next(r for r in json.loads(Path('/workspace/adamic/stage3/adapt/40-explicit-any/class-rules.json').read_text())['brands'] if r['id']==103)
p=Path(__file__).parent/'replay40'/rule['file'];b=p.read_bytes()
a,z=(rule['after'],rule['previous']) if sys.argv[1]=='mutate' else (rule['previous'],rule['after'])
assert b.count(a.encode())==1;p.write_bytes(b.replace(a.encode(),z.encode()))
print(sys.argv[1]+' rule 103')

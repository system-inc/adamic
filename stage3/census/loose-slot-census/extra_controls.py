import sys,json
from pathlib import Path
s=Path('/workspace/loose-slot-scratch');r=s/'head';d=r/'stage3/census/loose-slot-census';sys.path.insert(0,str(d));from measure import run
for name in ['k17_captured_unknown','k18_entry_then_declared']:
 src=r/'internal/oracle/testdata/loose_slot_cycles'/(name+'.ts');out=d/'evidence/controls'/name
 run([s/'head-census',src,str(out)+'-production.json'],str(out)+'-production',cwd=r)
 run([s/'type-count',src,str(out)+'-type.json'],str(out)+'-type',cwd=r)
 print(name,json.loads(Path(str(out)+'-production.json').read_text()),json.loads(Path(str(out)+'-type.json').read_text()),flush=True)

"""Observe the serializer through its real nongeneric initializer context."""
import json,subprocess,sys
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
label=sys.argv[1]
rows=json.loads((out/(label+'-observing-sites.json')).read_text())['checker']
row=next(r for r in rows if r['requested_line']==10911)
assert row['observing_unit_name']=='symbolTableToDeclarationStatements'
for side in ['parent','head']:
 name=f'{label}-{side}-serializer'
 cmd=['env','LATENT_FULL=1','LATENT_ASSERT_NO_OUTPUT=1',f'/tmp/branded-{side}-focus','-project',f'/tmp/branded-{label}/src/tsc/tsc.ts','-where',row['observing_unit'],'-output',str(out/(name+'.jsonl'))]
 rc=subprocess.run([sys.executable,str(out/'run.py'),name,str(repo),*cmd]).returncode
 if rc:raise SystemExit(rc)
cmd=['python3',str(repo/'stage3/census/pair/pair.py'),str(out/f'{label}-parent-serializer.jsonl'),str(out/f'{label}-head-serializer.jsonl'),f'/tmp/branded-{label}','--typescript',str(repo/'stage3/api/node_modules/typescript'),'--json',str(out/f'{label}-serializer-pair.json'),'--report',str(out/f'{label}-serializer-pair.md')]
rc=subprocess.run([sys.executable,str(out/'run.py'),label+'-serializer-pair',str(repo),*cmd]).returncode
raise SystemExit(rc)

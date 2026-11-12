import json,sys
from pathlib import Path
out=Path('/workspace/scratch/path-adapt-proof')
findings=[f for line in (out/'census-mutant.jsonl').read_text().splitlines() for f in json.loads(line).get('findings',[])]
matching=[f for f in findings if f['kind']=='Refused' and f['where'].endswith('/path.ts:513:45') and f['reason']=="a cast the runtime can't check"]
assert matching,'restoring mutant failed to restore the expected refusal'
print(json.dumps({'expectedRed':True,'restoredRefusal':matching},indent=2),flush=True)
sys.exit(1)

"""Reassemble original 20/20/20/19 analysis inputs from isolated root runs."""
import json
from pathlib import Path
out=Path(__file__).resolve().parent
for begin in range(0,79,20):
 data=b''.join((out/f'root-{i:02}.jsonl').read_bytes() for i in range(begin+1,min(begin+21,80)))
 rows=[json.loads(line) for line in data.splitlines()][::2]
 (out/f'group-{begin//20+1}.jsonl').write_bytes(data)
 (out/f'group-{begin//20+1}-rows.json').write_text(json.dumps(rows,indent=2)+'\n')

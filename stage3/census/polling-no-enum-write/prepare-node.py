#!/usr/bin/env python3
"""Prepare the stock checker host input from Adamic's pinned Node package and table."""
from pathlib import Path
import hashlib,json,shutil,sys
repo=Path(sys.argv[1]).resolve();dest=Path(sys.argv[2]).resolve()
shutil.copytree(repo/'internal/load/node_types/node_modules',dest/'node_modules')
rows=json.loads((repo/'internal/load/host_axioms.json').read_text())
for name in sorted({r['file'] for r in rows}):
 p=dest/'node_modules/@types/node'/name;lines=p.read_text().splitlines()
 for row in rows:
  if row['file']==name:
   assert lines[row['line']-1]==row['published'],(name,row['line'])
   lines[row['line']-1]=row['corrected']
 p.write_text('\n'.join(lines)+'\n')
print(json.dumps(dict(rows=len(rows),axioms_sha256=hashlib.sha256((repo/'internal/load/host_axioms.json').read_bytes()).hexdigest())))

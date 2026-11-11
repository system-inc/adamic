#!/usr/bin/env python3
import collections,gzip,json,sys
from pathlib import Path
p=Path(sys.argv[1]); op=gzip.open if p.suffix=='.gz' else open
with op(p,'rt') as f: rows=[json.loads(line) for line in f if line.strip()]
findings=[x for r in rows[1:] for x in r.get('findings',[]) or []]
units=[x for r in rows[1:] for x in r.get('units',[]) or []]
guards=[x for x in findings if 'a view of ' in x.get('reason','') or 'a view with no source type' in x.get('reason','')]
result={'header':rows[0],'files':len(rows)-1,'unit_status':dict(collections.Counter(x['status'] for x in units)),'finding_kinds':dict(collections.Counter(x['kind'] for x in findings)),'guard_findings':guards}
print(json.dumps(result,indent=2))

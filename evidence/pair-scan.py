"""Pair scan-all events, retaining the multiplicity of every stop."""
import json
import sys
from pathlib import Path

from pair import location, owner, remove_pairs, semantics

def read(path):
    report=json.loads(Path(path).read_text())
    findings=[
        dict(kind=row['Kind'],phase=row['Rule'],reason=row['Words'],text=row['Words'],
             where=row['Site'],census_file='',unit='')
        for row in (report.get('Stops') or [])
    ]
    return report,findings

(base,b),(head,h)=map(read,sys.argv[1:3])
removed,added,matched=remove_pairs(b,h,lambda row:(semantics(row),location(row)))
out={
    'base':{k:base.get(k) for k in ('Boundary','Completed','UnderlyingAbort','load_error')},
    'head':{k:head.get(k) for k in ('Boundary','Completed','UnderlyingAbort','load_error')},
    'events':{'base':len(b),'head':len(h),'unchanged':len(matched)},
    'gone':removed,
    'new_or_reached':added,
}
print(json.dumps(out,indent=2))

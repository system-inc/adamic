#!/usr/bin/env python3
"""Compare exact production census observations; reject any newly exposed diagnostic."""
import argparse
import collections
import json
from pathlib import Path
import re

p=argparse.ArgumentParser()
p.add_argument('before',type=Path)
p.add_argument('after',type=Path)
p.add_argument('--before-tree',required=True)
p.add_argument('--after-tree',required=True)
p.add_argument('--before-wrapper')
p.add_argument('--after-wrapper')
a=p.parse_args()
def observations(file,tree,wrapper):
    rows=[json.loads(line) for line in file.read_text().splitlines()]
    assert len(rows)==2 and all(row['kind']=='checker' for row in rows)
    def norm(row):
        result=[d.replace(str(Path(tree))+'/','') for d in row['diagnostics']]
        if wrapper:result=[d.replace(wrapper,'all-source-roots.a') for d in result]
        return sorted(result)
    assert norm(rows[0])==norm(rows[1]),'root and aggregate disagree'
    return norm(rows[0])
before=observations(a.before,a.before_tree,a.before_wrapper)
after=observations(a.after,a.after_tree,a.after_wrapper)
removed=sorted(set(before)-set(after))
added=sorted(set(after)-set(before))
def own(d):
    m=re.match(r'([^\n]+?):\d+:\d+: error TS',d)
    return m[1] if m else '<global>'
b,c=collections.Counter(map(own,before)),collections.Counter(map(own,after))
print(json.dumps(dict(before=len(before),after=len(after),removed=removed,added=added,
    perFile=[dict(file=f,before=b[f],after=c[f]) for f in sorted(b.keys()|c.keys())]),indent=2))
raise SystemExit(1 if added else 0)

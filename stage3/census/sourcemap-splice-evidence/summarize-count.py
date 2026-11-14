from pathlib import Path
from collections import Counter
import json,sys
root=Path(sys.argv[1]);totals=Counter();hist=Counter();post=Counter()
for p in root.glob('*.json'):
 r=json.loads(p.read_text());totals.update(r['totals']);hist.update(r['hist']);post.update(r['post'])
assert totals['flush']>0 and totals['clear']==totals['nonempty']
assert all(int(k.split(':')[1])==0 for k in post)
r=dict(totals=totals,reads={kind:sum(n for k,n in hist.items() if k.startswith(kind+':')) for kind in ['length','push','apply']},allReadLengths=dict(sorted(hist.items())),readsAfterClearBeforeNextPush=dict(post),processes=len(list(root.glob('*.json'))))
Path(sys.argv[2]).write_text(json.dumps(r,indent=2)+'\n');print(json.dumps(r,indent=2))

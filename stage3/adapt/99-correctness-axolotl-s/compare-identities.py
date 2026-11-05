#!/usr/bin/env python3
"""Compare full task/title/outcome multisets, retaining duplicate title counts."""
from collections import Counter
import gzip,hashlib,json,pathlib,sys

def read(profile):
    records=Counter();task_rows=0
    for line in pathlib.Path(profile).read_text().splitlines():
        row=json.loads(line)
        assert 'task' in row and 'passes' in row and 'errors' in row,row
        task_rows+=1
        for status,key in [('pass','passes'),('fail','errors')]:
            for record in row[key]:
                name=record['name'];assert isinstance(name,list)
                records[(row['task']['runner'],row['task']['file'],tuple(name),status)]+=1
    return records,task_rows

def canonical(counter):
    return [[runner,file,list(name),status,count] for (runner,file,name,status),count in sorted(counter.items())]

main,main_tasks=read(sys.argv[1]);accepted,accepted_tasks=read(sys.argv[2]);out=pathlib.Path(sys.argv[3])
summary={'mainTasks':main_tasks,'acceptedTasks':accepted_tasks,'identitiesEqual':main==accepted,'multiplicityRetained':True,'outcomesIncluded':True}
for label,rows in [('main',main),('accepted',accepted)]:
    data=json.dumps(canonical(rows),separators=(',',':')).encode()
    counts=Counter()
    for key,count in rows.items():counts[key[3]]+=count
    summary[label]={'counts':dict(counts),'total':sum(rows.values()),'uniqueIdentities':len(rows),'sha256':hashlib.sha256(data).hexdigest()}
    out.with_name(label+'-identities.json.gz').write_bytes(gzip.compress(data,mtime=0))
summary['missingFromAccepted']=canonical(main-accepted)
summary['extraInAccepted']=canonical(accepted-main)
out.write_text(json.dumps(summary,indent=2)+'\n')
assert main==accepted,summary
assert summary['main']['counts']=={'pass':106366,'fail':1},summary
assert summary['accepted']['counts']=={'pass':106366,'fail':1},summary
print(json.dumps({k:v for k,v in summary.items() if k not in ['missingFromAccepted','extraInAccepted']},indent=2))

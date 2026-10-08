#!/usr/bin/env python3
import collections,json,pathlib
scratch=pathlib.Path('/tmp/decode-ascii')
p=json.loads((scratch/'after.cpuprofile').read_text());frames={n['id']:n['callFrame'] for n in p['nodes']}
window=p['measurementWindow'];timestamp=p['startTime'];weights=collections.Counter();counts=collections.Counter()
for ident,delta in zip(p['samples'],p['timeDeltas']):
    timestamp+=delta
    if window['startMicros']<=timestamp<=window['endMicros']:
        f=frames[ident];key=(f['functionName'],f.get('url',''),f.get('columnNumber',-1));weights[key]+=delta;counts[key]+=1
rows=[dict(function=k[0],url=k[1],column=k[2],selfMilliseconds=v/1000,selfPercent=100*v/sum(weights.values()),samples=counts[k]) for k,v in weights.most_common()]
result=dict(sampledMilliseconds=sum(weights.values())/1000,samples=sum(counts.values()),measurementWindow=window,top10=rows[:10],allFunctions=rows)
pathlib.Path('cloud/reports/decode-ascii/profile-after.json').write_text(json.dumps(result,indent=2)+'\n')
for row in rows:
    if row['function']=='decode' and row['url'].startswith('wasm:'):print(json.dumps(row))

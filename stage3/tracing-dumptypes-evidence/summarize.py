from pathlib import Path
import json,sys,collections
S=Path(__file__).resolve().parent
for name in sys.argv[1:]:
 p=S/name;rows=[json.loads(l) for f in p.glob('*.jsonl') for l in f.read_text().splitlines()]; processes=[json.loads(f.read_text()) for f in p.glob('*.json')]
 result=dict(visits=len(rows),processes=len(processes),objectOwn=sum(r['objectOwn'] for r in rows),intrinsicOwn=sum(r['intrinsicOwn'] for r in rows),objectRawDifferences=sum(not r['objectRawEqual'] for r in rows),maskDifferences=[sum(not r['maskEqual'][i] for r in rows) for i in range(5)],nonObjectFlagsValues=sum(r['nonObjectFlagsValue'] for r in rows),intrinsicDifferences=sum(not r['intrinsicEqual'] for r in rows),inReadDifferences=sum(not r['inReadEqual'] for r in rows),nonObjectFlagCounterexamples=[r for r in rows if r['nonObjectFlagsValue']])
 (S/(name+'-summary.json')).write_text(json.dumps(result,indent=2)+'\n');print(name,json.dumps({k:v for k,v in result.items() if k!='nonObjectFlagCounterexamples'}))

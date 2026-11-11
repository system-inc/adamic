from pathlib import Path
import subprocess,json,hashlib,os
S=Path(__file__).resolve().parent;lib=S/'stock';rows=json.loads((S/'stock-projects.json').read_text()); results=[]
f=lib/'_tsc.js';text=f.read_text();a=text.index('function dumpTypes(types) {');b=text.index('function dumpLegend()',a);body=text[a:b];body=body.replace('const objectFlags = type.objectFlags;','const objectFlags = getObjectFlags(type);').replace('intrinsicName: type.intrinsicName,','intrinsicName: type.flags & 402431 ? type.intrinsicName : undefined,');f.write_text(text[:a]+body+text[b:])
counts=S/'stock-candidate-counts';counts.mkdir()
for i,row in enumerate(rows):
 types=S/('stock-trace-'+str(i))/'types.json';before=types.read_bytes();(S/('stock-before-types-'+str(i)+'.json')).write_bytes(before)
 env=dict(os.environ,NODE_OPTIONS='--require='+str(S/'preload.cjs'),DUMP_COUNTER_DIR=str(counts))
 with (S/('stock-candidate-project-'+str(i)+'.log')).open('w') as log:p=subprocess.run(row['command'],env=env,stdout=log,stderr=subprocess.STDOUT)
 after=types.read_bytes();(S/('stock-after-types-'+str(i)+'.json')).write_bytes(after)
 r=dict(project=row['project'],exit=p.returncode,beforeExit=row['exit'],equal=before==after,bytes=len(before),beforeSha256=hashlib.sha256(before).hexdigest(),afterSha256=hashlib.sha256(after).hexdigest());results.append(r);print(json.dumps(r),flush=True)
(S/'stock-byte-witness-results.json').write_text(json.dumps(results,indent=2)+'\n')
assert all(r['equal'] and r['exit']==r['beforeExit'] for r in results)

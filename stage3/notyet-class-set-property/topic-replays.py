import subprocess,json
from pathlib import Path
base='/tmp/adamic-class-set-property-census/'
sites=[('method-core','core.ts:1544:5','replacing a represented method at runtime'),('method-sys','sys.ts:1382:5','replacing a represented method at runtime'),('array-tracing','tracing.ts:164:9','assigning a field of a value'),('array-core','core.ts:307:5','assigning a field of a value'),('node-parser','parser.ts:1341:5','storing true | Node | undefined in a field'),('node-utilities','utilities.ts:8987:17','storing true | Node | undefined in a field'),('false-type','checker.ts:15245:9','storing false | Type in a field'),('path','builder.ts:668:13','storing Path | undefined in a field'),('false-symbol','checker.ts:34075:13','storing false | Symbol in a field'),('scalar','factory/emitNode.ts:242:5','storing string | number in a field'),('false-array','moduleNameResolver.ts:2277:12','storing false | string[] | undefined in a field'),('any','tsbuildPublic.ts:2080:5','storing any in a field')]
rows=[]
for name,where,reason in sites:
 log='/tmp/adamic-topic-replay-'+name+'.log'
 with open(log,'w') as out:r=subprocess.run(['/tmp/adamic-topic-replay','-project',base+'src/tsc/tsc.ts','-where',base+'src/compiler/'+where,'-kind','NotYet','-reason',reason],stdout=out,stderr=subprocess.STDOUT)
 raw=Path(log).read_text()
 try:
  x,_=json.JSONDecoder().raw_decode(raw[raw.index('{'):]);findings=[{'kind':f['kind'],'where':f['where'].replace(base,''),'reason':f['reason']} for f in x.get('findings',[]) if f['kind']!='Boundary']
 except Exception:findings=[{'error':raw[-1500:]}]
 row={'site':name,'where':'src/compiler/'+where,'reason':reason,'exit':r.returncode,'findings':findings,'log':log};rows.append(row);print(json.dumps(row),flush=True)
 Path('/tmp/adamic-topic-replays.json').write_text(json.dumps(rows,indent=2)+'\n')

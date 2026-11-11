from run import run,raw
import os,json,collections,sys
summary={}
for side,tree in [('before',raw/'before-counted'),('after',raw/'counted-product/tree')]:
 directory=raw/(side+'-reentrant-counts')
 if '--summarize-existing' not in sys.argv:directory.mkdir()
 env=dict(os.environ,NODE_OPTIONS='--require='+str(raw/'preload.cjs'),TRACING_LENGTH_COUNTER_DIR=str(directory))
 if '--summarize-existing' not in sys.argv:run(side+'-reentrant-counter-control',['node',raw/'reentrant-probe.cjs',tree],env=env)
 else:assert json.loads((raw/(side+'-reentrant-counter-control.json')).read_text())['exit']==0
 result=json.loads((raw/(side+'-reentrant-counter-control.log')).read_text());assert result['class']=='RangeError' and result['message']=='Invalid array length' and result['array']==[] and result['reentries']==1
 hist={s:collections.Counter() for s in ['popCall','popWrite','popAllCall','popAllWrite']}
 for p in directory.glob('*.json'):
  d=json.loads(p.read_text())
  for s,h in d['counts'].items():hist[s].update({int(k):v for k,v in h.items()})
 summary[side]={'result':result,'histograms':{s:dict(h) for s,h in hist.items()}}
assert summary['before']['histograms']['popWrite']=={0:1}
assert summary['after']['histograms']['popWrite']=={}
(raw/'reentrant-controls-report.json').write_text(json.dumps(summary,indent=2)+'\n');print('baseline empty decrement counter fires once; candidate guard throws before splice; outcomes match')

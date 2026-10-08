#!/usr/bin/env python3
import json,pathlib,os,gzip
here=pathlib.Path(os.environ.get('ADAMIC_CFG_FIXTURE_OUTPUT',str(pathlib.Path(__file__).resolve().parent)));cases=json.loads(gzip.decompress((here/'cases.json.gz').read_bytes()));out=[]
def bit(x):return 'true' if x else 'false'
def join(x):return ','.join(map(str,x))
def encoded(text):
 data=text.encode('utf-16-le');return ','.join(str(int.from_bytes(data[i:i+2],'little')) for i in range(0,len(data),2))
for i,c in enumerate(cases):
 out.append(f"case:{i}:{c['op']}")
 if c['op']!='paths':
  for n in c['nodes']:out.append(f"node:{n['index']}:{n['rootOwner']}:{bit(n['rootFlag'])}:{bit(n['alwaysTruthy'])}:{bit(n['destructuring'])}:{bit(n['throwable'])}:{';'.join(encoded(x) for x in (n['labels'] or []))}:{encoded(n['normalized'])}")
 if c['op']=='paths':
  for cyclic,every,exit,distance,present in c['queries']:out.append(f'path:{bit(cyclic)}:{bit(every)}:{bit(exit)}:{distance}:{bit(present)}')
  out.append(f"exit:{c['distance']}:{bit(c['present'])}:{bit(c['single'])}")
 elif c['op']=='roots':
  for node,bare,loops in c['summaries']:out.append(f'root:{node}:{bit(bare)}:{join(loops)}')
 else:
  offsets={}
  for kind,node,block,reach,events in c['trace']:
   out.append(f'trace:{kind}:{node}:{block}:{bit(reach)}:{join(events)}')
   for event in events:
    slot=offsets.get(block,0);out.append(f'emit:{block}:{slot}');offsets[block]=slot+1
  out.append(f"graph:{bit(c['end'])}:{join(c['finals'])}:{join(c['thrown'])}")
  for index,reach,incoming,final,thrown,succ,pred,barriers,count in c['blocks']:
   barr='null' if barriers is None else ','.join(bit(x) for x in barriers)
   out.append(f'block:{index}:{bit(reach)}:{bit(incoming)}:{bit(final)}:{bit(thrown)}:{join(succ)}:{join(pred)}:{barr}:{count}')
  if c['op']=='build':out.append(f"hooks:{len(c['trace'])}")
(here/'expected.txt.gz').write_bytes(gzip.compress(('\n'.join(out)+'\n').encode(),mtime=0))
print(len(cases),len(out))

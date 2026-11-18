from pathlib import Path
import json,collections,itertools
W=Path('/tmp/family5-alternating');a=Path('/workspace/family5-main-lane-alone/adapted-tree');b=Path('/workspace/family5-c02-lane/adapted-tree')
B64={c:i for i,c in enumerate('ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/')}
def decoded(mapping):
 src=line=col=name=0
 for gl,ls in enumerate(mapping.split(';')):
  gc=0
  for seg in ls.split(','):
   if not seg:continue
   vals=[];v=shift=0
   for ch in seg:
    d=B64[ch];v|=(d&31)<<shift
    if d&32:shift+=5
    else:vals.append(-(v>>1) if v&1 else v>>1);v=shift=0
   gc+=vals[0]
   if len(vals)>1:
    src+=vals[1];line+=vals[2];col+=vals[3]
    if len(vals)>4:name+=vals[4]
    yield (gl,gc,src,line,col,name if len(vals)>4 else None)
   else:yield (gl,gc,None,None,None,None)
r=[]
for f in ['built/local/_tsc.js.map','built/local/run.js.map','built/local/typescript.js.map']:
 x=json.loads((a/f).read_text());y=json.loads((b/f).read_text());keys=[k for k in sorted(x.keys()|y.keys()) if x.get(k)!=y.get(k)];counts=collections.Counter();samples=[];gen_equal=True;n=0
 for u,v in itertools.zip_longest(decoded(x['mappings']),decoded(y['mappings'])):
  n+=1
  if u is None or v is None:gen_equal=False;counts['segment count differs']+=1;continue
  gen_equal &= u[:2]==v[:2]
  if u!=v:
   source=x['sources'][u[2]] if u[2] is not None else None
   k=(source, None if u[3] is None else v[3]-u[3],None if u[4] is None else v[4]-u[4],u[2]==v[2],u[5]==v[5]);counts[str(k)]+=1
   if len(samples)<3:samples.append({'generated_line_1based':u[0]+1,'generated_column_0based':u[1],'source':source,'main_source_line_1based':u[3]+1 if u[3] is not None else None,'c02_source_line_1based':v[3]+1 if v[3] is not None else None})
 r.append({'file':f,'different_json_fields':keys,'generated_coordinates_all_identical':gen_equal,'decoded_segments':n,'changed_segment_classes':dict(counts),'examples':samples})
(W/'source-map-explanation.json').write_text(json.dumps(r,indent=2)+'\n')
# Enumerate every changed build-info field, rather than treating differing hashes as opaque.
def differences(x,y,p=''):
 if type(x)!=type(y):yield(p,x,y)
 elif isinstance(x,dict):
  for k in sorted(x.keys()|y.keys()):
   if k not in x or k not in y:yield(p+'.'+k,x.get(k),y.get(k))
   else:yield from differences(x[k],y[k],p+'.'+k)
 elif isinstance(x,list):
  if len(x)!=len(y):yield(p+'.length',len(x),len(y))
  for i,(u,v) in enumerate(zip(x,y)):yield from differences(u,v,p+'['+str(i)+']')
 elif x!=y:yield(p,x,y)
info=[]
for d in json.loads((W/'artifact-summary.json').read_text())['differing_files']:
 f=d['file']
 if f.endswith('.tsbuildinfo'):
  changes=[{'path':p,'main':x,'c02':y} for p,x,y in differences(json.loads((a/f).read_text()),json.loads((b/f).read_text()))]
  info.append({'file':f,'changes':changes,'all_changed_values_are_sha256_hashes':all(isinstance(v,str) and len(v)==64 and all(c in '0123456789abcdef' for c in v) for z in changes for v in [z['main'],z['c02']])})
(W/'build-info-explanation.json').write_text(json.dumps(info,indent=2)+'\n')
for x in r:print(x['file'],x['different_json_fields'],x['generated_coordinates_all_identical'],x['changed_segment_classes'])
print('Build-info files:',len(info),'all hash-only:',all(x['all_changed_values_are_sha256_hashes'] for x in info))

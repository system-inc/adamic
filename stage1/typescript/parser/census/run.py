#!/usr/bin/env python3
"""Compare canonical preorder node records. No parser normalization except an
explicit shape-only view that removes full node flags; diagnostics retain code,
byte position, byte length, category, and message. Each flag class counts files.
Usage: run.py MANIFEST OUTPUT_JSON [ORACLE_BINARY]."""
import sys,subprocess,tempfile,json,re,collections,pathlib,gzip
base=pathlib.Path(__file__).resolve().parent
flags={}
for line in (base.parents[3]/'cohere/TypeScript/tsc/internal/ast/nodeflags.go').read_text().splitlines():
 m=re.search(r'NodeFlags(\w+)\s+NodeFlags = 1 << (\d+)',line)
 if m:flags[1<<int(m[2])]=m[1]
paths=pathlib.Path(sys.argv[1]).read_text().splitlines(); result={'files':len(paths),'parsed':0,'identical':0,'shape_identical':0,'classes':{},'failures':[],'records':[]}
classes={}
def runs(cmd, batch):
 with tempfile.NamedTemporaryFile(mode='w',suffix='.txt') as mf:
  mf.write('\n'.join(batch)+'\n');mf.flush()
  try:
   p=subprocess.run(cmd(mf.name),capture_output=True,text=True,timeout=180)
  except subprocess.TimeoutExpired:
   if len(batch)>1:
    half=len(batch)//2;return runs(cmd,batch[:half])+runs(cmd,batch[half:])
   return [(['adapter-failure timeout'], '')]
  chunks=re.split(r'^case \d+\n',p.stdout,flags=re.M)[1:]
  if len(chunks)!=len(batch):
   if len(batch)>1:
    half=len(batch)//2;return runs(cmd,batch[:half])+runs(cmd,batch[half:])
   return [(['adapter-failure '+p.stderr[:500]],p.stderr[:500])]
  return [(c.splitlines(),p.stderr[:500]) for c in chunks]
def classify(a,b):
 ac=[x for x in a if x and x[0].isdigit()];bc=[x for x in b if x and x[0].isdigit()]
 ad=[x for x in a if x.startswith('diagnostic')];bd=[x for x in b if x.startswith('diagnostic')]
 found=set(); examples={}
 def add(c,x,y):found.add(c);examples.setdefault(c,{'port':x,'go':y})
 if ad!=bd:add('diagnostics',ad,bd)
 if len(ac)!=len(bc):add('children/count',len(ac),len(bc))
 for x,y in zip(ac,bc):
  xp=x.split('\t');yp=y.split('\t');xx=xp[0].split();yy=yp[0].split()
  if xx[:2]!=yy[:2]:add('kind/children',x,y);continue
  if xx[2:4]!=yy[2:4]:add('position',x,y)
  delta=int(xx[4])^int(yy[4])
  for bit,name in flags.items():
   if delta&bit:add('flags/'+name,x,y)
  if xx[5:]!=yy[5:] or xp[1:]!=yp[1:]:add('payload/list/token-flags',x,y)
 return found,examples
for start in range(0,len(paths),40):
 batch=paths[start:start+40]
 aa=runs(lambda m:['node','--disable-warning=ExperimentalWarning',str(base/'port.mjs'),m],batch)
 bb=runs(lambda m:[sys.argv[3] if len(sys.argv)>3 else '/tmp/parser-census-oracle','--manifest',m,'--whole','--recovery'],batch)
 for path,(a,ae),(b,be) in zip(batch,aa,bb):
  failures=[{'side':side,'error':line} for side,lines in [('port',a),('go',b)] for line in lines if line.startswith('adapter-failure')]
  if failures:result['failures'].append({'path':path,'failures':failures});continue
  result['parsed']+=1
  if a==b:result['identical']+=1
  found,examples=classify(a,b)
  if not any(not c.startswith('flags/') for c in found):result['shape_identical']+=1
  result['records'].append({'path':path,'classes':sorted(found)})
  size=pathlib.Path(path).stat().st_size
  for c in found:
   item=classes.setdefault(c,{'files':0,'shortest_bytes':None})
   item['files']+=1
   if item['shortest_bytes'] is None or size<item['shortest_bytes']:
    item.update(shortest_bytes=size,path=path,input=pathlib.Path(path).read_text(errors='replace'),example=examples[c])
    folder=pathlib.Path(sys.argv[2]).with_suffix('')/'examples';folder.mkdir(parents=True,exist_ok=True)
    label=c.replace('/','-');(folder/(label+pathlib.Path(path).suffix)).write_bytes(pathlib.Path(path).read_bytes())
    for side,tree in [('go',b),('port',a)]:
     with gzip.open(folder/(label+'.'+side+'.tree.gz'),'wt') as out:out.write('\n'.join(tree)+'\n')
 result['classes']=dict(sorted(classes.items(),key=lambda x:-x[1]['files']))
 pathlib.Path(sys.argv[2]).write_text(json.dumps(result,ensure_ascii=True,indent=2)+'\n')
 print(f'{start+len(batch)}/{len(paths)} parsed={result["parsed"]} identical={result["identical"]} shape={result["shape_identical"]} failures={len(result["failures"])}',flush=True)

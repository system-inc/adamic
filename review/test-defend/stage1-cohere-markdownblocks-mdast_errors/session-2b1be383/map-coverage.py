import pathlib,json,gzip,difflib
p=pathlib.Path('/tmp/defend-mdast');executed=json.loads(p.joinpath('port-executed-sources.json').read_text());trans=json.loads(p.joinpath('port-transforms.json').read_text());alphabet='ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/'
def numbers(s):
 out=[];v=shift=0
 for c in s:
  n=alphabet.index(c);v+=(n&31)<<shift
  if n&32:shift+=5
  else:out.append(-(v>>1) if v&1 else v>>1);v=shift=0
 return out
def points(data):
 text=data['text'];starts=[0]
 for c in text:
  starts.append(starts[-1]+(2 if ord(c)>65535 else 1))
 lineStarts=[0]+[starts[i+1] for i,c in enumerate(text) if c=='\n'];source=ol=oc=0;result=[]
 for gl,line in enumerate(data['map']['mappings'].split(';')):
  gc=0
  for seg in line.split(','):
   if not seg:continue
   vals=numbers(seg);gc+=vals[0]
   if len(vals)>=4:
    source+=vals[1];ol+=vals[2];oc+=vals[3]
    if gl<len(lineStarts):result.append((lineStarts[gl]+gc,ol+1))
 return result
profiles={}
for name in ['TestMdastIdentifierWitnesses','TestNativeMdastConstruction']:
 hit={};functions=[]
 with gzip.open(p/(name+'-port-v8.json.gz'),'rt') as f:raw=json.load(f)
 for script in raw:
  rel=script['url'][len('file:///workspace/adamic/'):]
  ranges=[r for f in script['functions'] for r in f['ranges']]
  def units(text):
   raw=text.encode('utf-16-le');return ''.join(chr(int.from_bytes(raw[i:i+2],'little')) for i in range(0,len(raw),2))
  blocks=difflib.SequenceMatcher(None,units(trans[rel]['text']),units(executed[rel]),autojunk=False).get_matching_blocks()
  functions.extend(dict(file=rel,function=f['functionName'],count=f['ranges'][0]['count']) for f in script['functions'])
  for off,line in points(trans[rel]):
   block=next((b for b in blocks if b.a<=off<b.a+b.size),None)
   if block is None:continue
   off=block.b+off-block.a
   active=[r for r in ranges if r['startOffset']<=off<r['endOffset']]
   if active:
    r=min(active,key=lambda r:r['endOffset']-r['startOffset']);key=f'{rel}:{line}';hit[key]=max(hit.get(key,0),r['count'])
 profiles[name]=dict(hit_lines=sorted(k for k,v in hit.items() if v),functions=functions)
a,b=list(profiles);profiles[a]['exclusive_lines']=sorted(set(profiles[a]['hit_lines'])-set(profiles[b]['hit_lines']));profiles[b]['exclusive_lines']=sorted(set(profiles[b]['hit_lines'])-set(profiles[a]['hit_lines']))
p.joinpath('port-line-coverage.json').write_text(json.dumps(profiles,indent=2))
aGo={s.split()[0]:int(s.split()[-1]) for s in p.joinpath(a+'.cover').read_text().splitlines()[1:]};bGo={s.split()[0]:int(s.split()[-1]) for s in p.joinpath(b+'.cover').read_text().splitlines()[1:]}
p.joinpath('go-exclusive-coverage.json').write_text(json.dumps({a:[k for k,v in aGo.items() if v and not bGo.get(k,0)],b:[k for k,v in bGo.items() if v and not aGo.get(k,0)]},indent=2))
for name in [a,b]:
 print(name,'port exclusive lines',len(profiles[name]['exclusive_lines']));print('\n'.join(k for k in profiles[name]['exclusive_lines'] if 'mdastArena' in k or 'identifier' in k))
print('Go exclusive counts',sum(bool(v) and not bGo.get(k,0) for k,v in aGo.items()),sum(bool(v) and not aGo.get(k,0) for k,v in bGo.items()))

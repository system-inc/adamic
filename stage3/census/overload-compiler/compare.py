"""Compare full observations, mapping unchanged candidate lines back through the source diff."""
import collections,difflib,json,pathlib,re,sys
mode,before,after,old_tree,new_tree,destination=sys.argv[1:]
a=pathlib.Path(old_tree);b=pathlib.Path(new_tree)
line_maps={}
for file in a.joinpath('src').rglob('*.ts'):
 rel=file.relative_to(a);other=b/rel
 if not other.exists():continue
 old=file.read_text();new=other.read_text()
 if old==new:continue
 mapping={}
 for block in difflib.SequenceMatcher(None,old.splitlines(),new.splitlines(),autojunk=False).get_matching_blocks():
  for i in range(block.size):mapping[block.b+i+1]=block.a+i+1
 line_maps[str(rel)]=mapping

def norm(s,tree,candidate=False):
 s=s.replace(str(tree)+'/','')
 def replace(m):
  file,line,col=m.groups();line=int(line)
  if candidate and file in line_maps:line=line_maps[file].get(line,line)
  return f'{file}:{line}:{col}'
 return re.sub(r'(src/[^\s:]+):(\d+):(\d+)',replace,s)
def load(file):
 if mode=='scan':return json.load(open(file))['Stops'] or []
 objects=[json.loads(s) for s in open(file) if s.strip()]
 if objects[-1].get('status')!='relation_report':raise ValueError('incomplete latent stream')
 return [f for x in objects if x.get('file') for f in x.get('findings',[])]
def key(x,tree,candidate=False):
 if mode=='scan':return tuple(norm(x[k],tree,candidate) for k in ('Kind','Rule','Site','Words'))
 return tuple(norm(x.get(k,''),tree,candidate) for k in ('root_kind','unit','phase','kind','where','reason'))
old=load(before);new=load(after)
ca=collections.Counter(key(x,a) for x in old);cb=collections.Counter(key(x,b,True) for x in new)
removed=ca-cb;added=cb-ca
changed=[]
for old_key in sorted(list(removed)):
 for new_key in sorted(list(added)):
  if old_key[:-1]==new_key[:-1] and removed[old_key] and added[new_key]:
   count=min(removed[old_key],added[new_key]);changed.append({'count':count,'before':list(old_key),'after':list(new_key)});removed[old_key]-=count;added[new_key]-=count
result={'mode':mode,'baseline_events':len(old),'candidate_events':len(new),'gone':[{'count':n,'event':list(k)} for k,n in sorted(removed.items()) if n],'new':[{'count':n,'event':list(k)} for k,n in sorted(added.items()) if n],'changed':changed,'comparison':'Unchanged source lines are mapped by SequenceMatcher. Changed words at an equal rule/site are paired and listed in full; no changed words are suppressed.'}
pathlib.Path(destination).write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v if not isinstance(v,list) else sum(x['count'] for x in v) for k,v in result.items() if k!='comparison'}))

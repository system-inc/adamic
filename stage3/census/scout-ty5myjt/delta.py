import pathlib,json,difflib,re,collections
root=pathlib.Path('/workspace/scratch/scout-ty5myjt')
def rows(p):
 data=[json.loads(line) for line in p.read_text().splitlines()]
 assert data[-1].get('status')=='relation_report','incomplete census'
 assert data[0]['module_format']=='CommonJSBundle' and data[0]['latent_mode']=='full'
 files=[r for r in data if r.get('file')];assert len(files)==1
 return files[0]['findings']
def compare(name,basepath,newpath,file):
 old=rows(basepath);new=rows(newpath)
 a=(root/'adapted/src/compiler'/file).read_text().splitlines();b=(root/name/'src/compiler'/file).read_text().splitlines();mapping={}
 for op,i,j,k,l in difflib.SequenceMatcher(None,a,b,autojunk=False).get_opcodes():
  if op=='equal' or (op=='replace' and j-i==l-k):
   mapping.update({line+1:i+(line-k)+1 for line in range(k,l)})
 def normalize(x,after):
  x=dict(x)
  def loc(m):
   line=int(m[2]);return 'src/compiler/'+file+':'+str(mapping.get(line,line) if after else line)+':'+m[3]
  for field in ['where','text','unit','reason']:
   x[field]=x.get(field,'').replace(str(root/name)+'/', '').replace(str(root/'adapted')+'/', '')
   x[field]=re.sub(r'src/compiler/('+re.escape(file)+r'):(\d+):(\d+)',loc,x.get(field,''))
  return x
 def index(xs,after):
  result={}
  for x in xs:
   n=normalize(x,after);key=tuple(n.get(k,'') for k in ['phase','kind','where','reason']);result[key]=n
  return result
 aa=index(old,False);bb=index(new,True)
 gone=[aa[k] for k in aa.keys()-bb.keys()];added=[bb[k] for k in bb.keys()-aa.keys()];changed=[{'before':aa[k],'after':bb[k]} for k in aa.keys()&bb.keys() if aa[k]['text']!=bb[k]['text']]
 return {'candidate':name,'file':file,'before_raw':len(old),'after_raw':len(new),'before_unique':len(aa),'after_unique':len(bb),'new':sorted(added,key=lambda x:x['where']),'gone':sorted(gone,key=lambda x:x['where']),'changed':changed}
if __name__=='__main__':
 import sys
 name=sys.argv[1];file='program.ts' if name.startswith('diag') else 'resolutionCache.ts';folder='results' if file=='program.ts' else 'parallel-results'
 result=compare(name,root/folder/'baseline'/(file+'.jsonl'),root/folder/name/(file+'.jsonl'),file)
 out=root/'adamic/stage3/census/scout-ty5myjt'/(name+'-delta.json');out.write_text(json.dumps(result,indent=2)+'\n');print(name,result['before_unique'],result['after_unique'],len(result['new']),len(result['gone']),len(result['changed']))

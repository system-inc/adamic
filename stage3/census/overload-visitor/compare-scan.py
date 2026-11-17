import pathlib,json,difflib,re,collections
out=pathlib.Path(__file__).resolve().parent;s=pathlib.Path('/workspace/overload-visitor-scratch')
b=json.loads((out/'baseline/visitorPublic.ts.scan.json').read_text())
result={}
for name in ['visitArrayWorker','visitParameterList','visitEachChild']:
 a=json.loads((out/name/'visitorPublic.ts.scan.json').read_text());before=(s/'baseline/src/compiler/visitorPublic.ts').read_text().splitlines();after=(s/name/'src/compiler/visitorPublic.ts').read_text().splitlines();mapping={};columns={}
 for tag,i,j,k,l in difflib.SequenceMatcher(None,before,after,autojunk=False).get_opcodes():
  if tag=='equal' or tag=='replace' and j-i==l-k:
   for offset in range(l-k):
    ln=k+offset+1;oldln=i+offset+1;mapping[ln]=oldln;cm={}
    for block in difflib.SequenceMatcher(None,before[oldln-1],after[ln-1],autojunk=False).get_matching_blocks():
     for c in range(block.size):cm[block.b+c+1]=block.a+c+1
    columns[ln]=cm
 def events(x,side):
  def norm(v):
   v=v.replace(str(s/side),'<tree>')
   if side!='baseline':v=re.sub(r'(<tree>/src/compiler/visitorPublic\.ts:)(\d+)(:\d+)',lambda m:m[1]+str(mapping.get(int(m[2]),int(m[2])))+':'+str(columns.get(int(m[2]),{}).get(int(m[3][1:]),int(m[3][1:]))),v)
   return v
  return collections.Counter(tuple(norm(r[k]) for k in ['Kind','Rule','Site','Words']) for r in x.get('Stops',[]))
 old=events(b,'baseline');new=events(a,name)
 result[name]={'baselineEvents':sum(old.values()),'candidateEvents':sum(new.values()),'new':[dict(zip(['Kind','Rule','Site','Words'],k),count=v) for k,v in sorted((new-old).items())],'gone':[dict(zip(['Kind','Rule','Site','Words'],k),count=v) for k,v in sorted((old-new).items())],'assignedOverloadObserved':False,'boundary':a['Boundary'],'underlyingAbort':a.get('UnderlyingAbort')}
 print(name,'new',sum((new-old).values()),'gone',sum((old-new).values()))
(out/'scan-diffs.json').write_text(json.dumps(result,indent=2)+'\n')

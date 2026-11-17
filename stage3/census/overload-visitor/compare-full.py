import pathlib,json,difflib,re,sys
s=pathlib.Path('/workspace/overload-visitor-scratch');out=pathlib.Path(__file__).resolve().parent
base=s/'baseline-visitor-full.jsonl'
def data(p):
 rows=[json.loads(x) for x in p.read_text().splitlines()]
 assert rows[-1].get('status')=='relation_report',f'incomplete {p}'
 return rows
b=data(base)
for name in sys.argv[1:]:
 p=s/(name+'-visitorPublic.ts-full.jsonl')
 if name=='visitParameterList':p=s/'visitParameterList-visitor-full.jsonl'
 a=data(p)
 before=(s/'baseline/src/compiler/visitorPublic.ts').read_text().splitlines()
 after=(s/name/'src/compiler/visitorPublic.ts').read_text().splitlines()
 mapping={}
 columns={}
 for tag,i,j,k,l in difflib.SequenceMatcher(None,before,after,autojunk=False).get_opcodes():
  if tag=='equal' or (tag=='replace' and j-i==l-k):
   for offset in range(l-k):
    ln=k+offset+1;oldln=i+offset+1;mapping[ln]=oldln
    cm={}
    for block in difflib.SequenceMatcher(None,before[oldln-1],after[ln-1],autojunk=False).get_matching_blocks():
     for c in range(block.size):cm[block.b+c+1]=block.a+c+1
    columns[ln]=cm
 def keys(rows,side):
  result={}
  def norm(v):
   v=str(v).replace(str(s/side),'<tree>')
   if side!='baseline':
    v=re.sub(r'(<tree>/src/compiler/visitorPublic\.ts:)(\d+)(:\d+)',lambda m:m[1]+str(mapping.get(int(m[2]),int(m[2])))+':'+str(columns.get(int(m[2]),{}).get(int(m[3][1:]),int(m[3][1:]))),v)
   return v
  for row in rows:
   for f in row.get('findings',[]):
    key=tuple(norm(f.get(k,'')) for k in ['kind','where','reason','text'])
    result[key]={k:norm(f.get(k,'')) for k in ['kind','where','reason','text']}
  return result
 old=keys(b,'baseline');new=keys(a,name)
 result={'set':name,'normalization':'tree prefix, matching source lines and matching characters on modified lines; equal-length replacement hunks map lines positionally; raw records retained','baselineUnique':len(old),'candidateUnique':len(new),'new':[new[k] for k in sorted(new.keys()-old.keys())],'gone':[old[k] for k in sorted(old.keys()-new.keys())]}
 result['newNonBoundary']=sum(x['kind']!='Boundary' for x in result['new']);result['goneNonBoundary']=sum(x['kind']!='Boundary' for x in result['gone'])
 (out/(name+'-full-diff.json')).write_text(json.dumps(result,indent=2)+'\n')
 print(name,result['newNonBoundary'],result['goneNonBoundary'])

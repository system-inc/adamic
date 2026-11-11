from summary import *
headbytes=(root/'debug-head.ts.txt').read_bytes().decode().splitlines(True)
def canonical(name, finding):
 source=(root/('debug-'+('mutant' if name=='mutant' else name)+'.ts.txt')).read_bytes().decode().splitlines(True)
 mapping={i+x+1:j+x+1 for tag,i,k,j,l in difflib.SequenceMatcher(None,source,headbytes,autojunk=False).get_opcodes() if tag=='equal' for x in range(k-i)}
 def loc(s):
  s=short(s)
  def replace(m):
   line,col=map(int,m.groups())
   return 'src/compiler/debug.ts:'+str(mapping.get(line,'EDITED-'+name+'-'+str(line)))+':'+str(col)
  return re.sub(r'src/compiler/debug\.ts:(\d+):(\d+)',replace,s)
 return (finding['kind'],loc(finding['where']),finding['reason'],loc(finding['text']))
if __name__=='__main__':
 data={n:debug(n) for n in ['head','mutant','parent']}
 canonical_sets={n:{canonical(n,f):f for f in d['findings']} for n,d in data.items()}
 delta={}
 for a,b in [('parent','head'),('mutant','head'),('parent','mutant')]:
  before,after=canonical_sets[a],canonical_sets[b]
  removed=[{'identity':list(k),'finding':before[k]} for k in before.keys()-after.keys()]
  added=[{'identity':list(k),'finding':after[k]} for k in after.keys()-before.keys()]
  delta[a+'-to-'+b]={'removed':sorted(removed,key=lambda x:x['identity']),'added':sorted(added,key=lambda x:x['identity'])}
 (root/'debug-delta.json').write_text(json.dumps(delta,indent=2)+'\n')
 print(json.dumps(delta,indent=2))

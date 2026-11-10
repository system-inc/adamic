from control import *
import difflib,re
parent=root/'adapted-parent'; head=root/'adapted-head'
a=(parent/'src/compiler/debug.ts').read_bytes().decode().splitlines(True)
b=(head/'src/compiler/debug.ts').read_bytes().decode().splitlines(True)
line_map={i+x+1:j+x+1 for tag,i,k,j,l in difflib.SequenceMatcher(None,a,b,autojunk=False).get_opcodes() if tag=='equal' for x in range(k-i)}
ledger={}
location=re.compile(re.escape(str(parent))+r'/src/compiler/debug\.ts:(\d+):(\d+)')
def convert(value):
 if isinstance(value,str):
  def remap(m):
   oldline,col=map(int,m.groups())
   if oldline not in line_map: raise ValueError('location on a changed parent source line: '+m.group())
   new= str(head)+'/src/compiler/debug.ts:'+str(line_map[oldline])+':'+str(col)
   ledger[m.group()]=new
   return new
  return location.sub(remap,value).replace(str(parent),str(head))
 if isinstance(value,list): return [convert(v) for v in value]
 if isinstance(value,dict): return {k:convert(v) for k,v in value.items()}
 return value
rows=[json.loads(s) for s in (root/'entry-parent.jsonl').read_text().splitlines() if s]
with (root/'entry-parent-mapped.jsonl').open('w') as out:
 for row in rows: out.write(json.dumps(convert(row))+'\n')
(root/'pair-location-map.json').write_text(json.dumps({'method':'Only exactly equal complete source lines are mapped; no location on an edited line is accepted. All non-debug source bytes are verified identical. Roots map to head.','parent_to_head':ledger},indent=2)+'\n')
code=run('pair',['python3','stage3/census/pair/pair.py',root/'entry-parent-mapped.jsonl',root/'entry-head.jsonl',head,'--typescript',root/'head/stage3/api/node_modules/typescript','--json',root/'pair-delta.json','--report',root/'pair-report.md'],root/'head')
sys.exit(code)

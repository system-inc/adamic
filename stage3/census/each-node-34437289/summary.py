from control import *
from collections import Counter,defaultdict
import difflib,re

def rows(name): return [json.loads(l) for l in (root/(name+'.jsonl')).read_text().splitlines() if l]
def debug(name):
 rs=rows('debug-'+name)
 fs=[f for row in rs[1:] for f in row.get('findings',[])]
 unique={ (f['kind'],f['where'],f['reason'],f['text']):f for f in fs }
 return {'header':rs[0],'units':[u for r in rs for u in r.get('units',[])],'findings':list(unique.values()),'observations':fs}

def short(s): return re.sub(re.escape(str(root))+r'/adapted-(?:head|parent|mutant)/','',s)
def md(s): return str(s).replace('|','\\|').replace('\n','<br>')
if __name__=='__main__':
 data={n:debug(n) for n in ['head','mutant','parent']}
 out={n:{'units':len(d['units']),'observations':len(d['observations']),'raw_unique_findings':len(d['findings']),'kinds':dict(Counter(f['kind'] for f in d['findings'])),'assertEachNode_findings':[f for f in d['findings'] if 256<=int(f['where'].rsplit(':',2)[1])<= (284 if n=='head' else 275) and '/compiler/debug.ts:' in f['where']]} for n,d in data.items()}
 (root/'debug-summary.json').write_text(json.dumps(out,indent=2)+'\n')
 for n,d in out.items(): print(n,json.dumps(d))

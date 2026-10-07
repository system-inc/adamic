"""Reproduce the bounded Go-oracle controls and semantic-mutant subset."""
from pathlib import Path
import gzip, itertools, json, random
here=Path(__file__).resolve().parent
rows=[]
for source,hint,arg,kind,present,entry,options,key,prefix in itertools.product(['','10px','red','var(--x)','calc(1 + 2)','💡','1/2','a b','url(x)','\n'],['','length','color'],['[*]','[length]','[color]','[]','[unknown]'],['named','arbitrary','unknown'],[False,True],[False,True],[0,1,2,3],['--x','--a b','--💡'],['','tw']):
 rows.append(dict(Source=source,Hint=hint,Argument=arg,ModifierKind=kind,ModifierPresent=present,EntryPresent=entry,Options=options,Key=key,Prefix=prefix))
data=(json.dumps(rows)+'\n').encode()
(here/'witnesses.json.gz').write_bytes(gzip.compress(data,mtime=0))
(here/'mutants.json').write_text(json.dumps(random.Random(1604).sample(rows,600))+'\n')
print(len(rows),'controls; 600 seeded mutant controls')

"""Reproduce bounded background and nested-theme controls."""
from pathlib import Path
import itertools,json
here=Path(__file__).resolve().parent
rows=[]
for source,opts,prefix,arg in itertools.product(['','cover','contain','cover,nonsense','cover,a b c','auto','10px 20%','10px 20% 30px','var(--x)','calc(1px)','cover,auto auto','cover,',' auto','cover,garbage garbage'],range(4),['','tw'],['--text-*','--text-*--height','--text-*--missing','--text-*--height-*--weight','--text','--missing-*']):
 rows.append(dict(Source=source,Candidate='sm',Argument=arg,Keys=['--text'],Nested=['--height','--weight','--missing','--empty','--height'],Prefix=prefix,Entries=[dict(Key='--text-sm',Value='12px',Options=opts),dict(Key='--text-sm--height',Value='1.5',Options=opts),dict(Key='--text-sm--weight',Value='700',Options=3-opts),dict(Key='--text-sm--empty',Value='',Options=opts)]))
rows += [dict(Source='cover',Candidate='sm',Keys=['--missing'],Nested=[],Argument='--text-*',Entries=[]),dict(Source='contain',Candidate='sm',Keys=['--text'],Nested=[],Argument='--text-*--empty',Entries=[dict(Key='--text-sm',Value='',Options=0),dict(Key='--text-sm--empty',Value='',Options=1)])]
(here/'witnesses.json').write_text(json.dumps(rows,indent=2)+'\n')
print(len(rows),'controls')

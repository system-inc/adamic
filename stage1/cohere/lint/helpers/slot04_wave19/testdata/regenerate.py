"""Deterministic controls; Go supplies parser and resolver observations at test time."""
import json
from pathlib import Path
HERE=Path(__file__).resolve().parent
values=['','1','1.25','1.3','2','-1','12.5%','50%','💡']
sources=['','plain','foo(bar','a)b','--value(integer)','--value(number)','--value(ratio)','--value(percentage)','--value(color)','--value(--text-*)','--value(--text-*--line-height)','--value(--default(4))','--modifier(integer)','--modifier(\'closest-side\')','calc(--value(integer) * -1)','calc(--value(bogus)) --modifier(integer)','--value(integer) --value(bogus) --modifier(integer)','--value(ratio) --modifier(integer)','--value(\'1\')','--value([*])','--value([length])','calc(foo(--value(number)))','--value(--literal-*)','--modifier(bogus) --value(integer)']
rows=[]
for si,source in enumerate(sources):
 for vi,val in enumerate(values):
  for mi,mod in enumerate([None,{'Kind':'named','Value':'2'},{'Kind':'named','Value':'closest-side'},{'Kind':'arbitrary','Value':'var(--x)'}]):
   value=None if vi==0 else {'Kind':'arbitrary' if vi==8 else 'named','Value':val,'Fraction':'1/2' if vi%2 else '1.5/2','DataType':'length' if vi==8 else ''}
   nodes=[]
   for k in ['rule','declaration','at-rule','declaration','context','declaration','at-root','declaration','comment','declaration','unknown','declaration']:
    nodes.append({'Kind':k,'Value':source if k=='declaration' else '', 'Present':(si+vi+mi)%3!=0,'Edges':[]})
   nodes[0]['Edges']=[1,2,4,6,8,10]
   for i in [2,4,6,8,10]:nodes[i]['Edges']=[i+1]
   rows.append({'Name':f'control-{si}-{vi}-{mi}','Source':source,'Value':value,'Modifier':mod,'Nodes':nodes,'Roots':[0], 'Flags':(si+vi+mi)%32,'SeedSets':(si+vi)%2==0,'Entries':[{'Key':'--text-1','Value':'12px','Options':1},{'Key':'--text-1--line-height','Value':'1.5','Options':1},{'Key':'--literal-1','Value':'--value(bogus)','Options':1}]})
# Fresh state makes each stop/flag mutation observable; every initial flag pattern is separately sampled.
for row in rows[:96]:
 copy=dict(row);copy['Name']='fresh-'+row['Name'];copy['Flags']=0;copy['SeedSets']=False;rows.append(copy)
(HERE/'witnesses.json').write_text(json.dumps(rows,ensure_ascii=True,indent=2)+'\n')
print('controls',len(rows))

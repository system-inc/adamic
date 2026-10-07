"""Generate bounded independent utility and raw-option inputs, not expected outputs."""
from pathlib import Path
import itertools
import json
HERE=Path(__file__).resolve().parent
controls=[]
values=[None,{'Kind':'named','Value':'2','Fraction':'2/3'},{'Kind':'named','Value':'bogus'},{'Kind':'named','Value':'1.3'},{'Kind':'arbitrary','Value':'var(--x)','DataType':'length'}]
modifiers=[None,{'Kind':'named','Value':'3'},{'Kind':'named','Value':'bad'}]
def decl(value):return {'Kind':'declaration','Property':'--x','Value':value,'ValuePresent':True}
expressions=['literal','--value(integer)','--value(ratio)','--value(integer) --modifier(integer)','--value(ratio) --modifier(integer)','--value(--default(4))','--value([length])','--value(number)','--modifier(integer) --value(integer)']
bodies=[[decl(value)] for value in expressions]+[[decl('--value(integer)'),decl('--value(color)')],[decl('--value(integer)'),decl('--value(ratio)'),decl('--value(color)')],[{'Kind':'rule','Selector':'&','Nodes':[decl('--value(integer)'),decl('--value(ratio)'),decl('--value(color)')]}],[],None]
for index,(body,value,modifier) in enumerate(itertools.product(bodies,values,modifiers)):
 controls.append({'Name':f'utility-{index}','OptionRaw':'{}','Definitions':{'foo':{'Name':'foo','Nodes':body}},'Candidate':{'Kind':'functional','Root':'foo','Value':value,'Modifier':modifier}})
for candidate in [None,{'Kind':'static','Root':'foo'},{'Kind':'arbitrary','Root':'foo'},{'Kind':'functional','Root':'missing'}]:controls.append({'Candidate':candidate,'Definitions':{},'OptionRaw':'{}'})
raws=['','{}',' {} \n','null','[]','[{}]','true','0','"text"','{',' ','{}{}','{"x":1}','{"z":0,"a":0}','{"x":0,"x":2}','{"environment":null}','{"\\u0061":1}','{"\\ud800":1}','{"\\ud83d\\ude00":1,"\\ue000":2}','{"é":1,"a":2}','{"line\\nkey":1}','{"":1}','{"x":[1,{},null]}']
for raw in raws:controls.append({'Name':'options-'+repr(raw),'OptionRaw':raw})
(HERE/'witnesses.json').write_text(json.dumps(controls,ensure_ascii=True,indent=2)+'\n')
print('generated',len(controls),'controls')

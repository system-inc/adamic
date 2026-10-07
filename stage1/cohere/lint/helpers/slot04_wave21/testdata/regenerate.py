"""Raw-byte and bare-value controls, independent of Adamic implementations."""
import json
from pathlib import Path
here=Path(__file__).resolve().parent
rows=[]
for b in range(256):
 for start in [-2,0,2]:
  value=[97,b,98,b,b,99];source=[120,120]+value+[121]
  rows.append({'Name':f'byte-{b}-{start}','BytesPresent':True,'SourceBytes':source,'ValueBytes':value,'Start':start,'End':start+len(value),'RangePresent':True})
for value in ['', 'a b','a\tb\nc\rd\ve\ff','💡 é 文字','a\u00a0b','a\u2003b','a\\nb']:
 for start,end in [(-1,2),(0,0),(0,len(value.encode())),(2,1),(0,len(value.encode())+1),(1,2)]:
  rows.append({'Name':f'range-{len(rows)}','Source':value,'Value':value,'ValuePresent':True,'Start':start,'End':end,'RangePresent':True})
for argument in ['number','integer','ratio','percentage','color','length','*','']:
 for value in ['', '0','1','01','1.25','1.3','-1','50%','12.5%','1e2','9007199254740993','var(--x)','💡']:
  for fraction in ['', '1/2','1.5/2','01/2','0/0','1 / 2','1/2/3']:
   rows.append({'Name':f'bare-{len(rows)}','Source':value,'Argument':argument,'Candidate':value,'Fraction':fraction})
(here/'witnesses.json').write_text(json.dumps(rows,ensure_ascii=True,indent=2)+'\n')
print('controls',len(rows))

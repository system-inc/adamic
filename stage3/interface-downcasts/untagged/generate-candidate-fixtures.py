#!/usr/bin/env python3
"""Write selector-projection fixtures from measured metadata, never source code."""
from pathlib import Path
import json,re,sys
root=Path(__file__).resolve().parent
inventory=json.loads(Path(sys.argv[1]).read_text())
selected=set(map(int,sys.argv[2:]))
folder=root/'candidates';folder.mkdir(exist_ok=True)
manifest=[]
for row in inventory['ranked_candidates']:
 if row['rank'] not in selected:continue
 selectors=row.get('member_selectors');assert selectors and all(selectors),row['rank']
 alias=row['declared_type'].replace(' | undefined','')
 if not re.fullmatch(r'[A-Za-z_][A-Za-z_0-9]*',alias):alias='CandidateTarget'
 nullable='undefined' in row['declared_type']
 prefix='// Independently written selector projection for candidate '+str(row['rank'])+'; unread compiler payloads are not copied.\n'
 prefix+='interface CandidatePayload {readonly label:string;}\n'
 for i,selector in enumerate(selectors):prefix+='interface CandidateMember%d {readonly kind:%s; readonly payload:CandidatePayload; readonly opaque%d:any;}\n'%(i,selector['type'],i)
 prefix+='type '+alias+' = '+' | '.join('CandidateMember'+str(i) for i in range(len(selectors)))+';\n'
 element=row['field']=='<element>'
 field='items' if element else row['field'];assert re.fullmatch(r'\w+',field)
 prefix+="interface ReadReceiver {readonly kind:'Receiver'; readonly "+field+('?:' if nullable and not element else ':')+('readonly '+alias+'[]' if element else alias)+';}\n'
 def source(values,nested=False,absent=False):
  text=prefix
  for i,value in enumerate(values):
   shape='{kind:'+json.dumps(value)+',payload:{label:'+('true' if nested else "'ok'")+'}}'
   target='['+shape+']' if element else shape
   raw="{kind:'Receiver' as const"+('' if absent else ','+field+':'+target)+'}'
   text+='const raw%d=%s;\nconst base%d:{readonly kind:\'Receiver\'}=raw%d;\nconst view%d=base%d as ReadReceiver;\n'%(i,raw,i,i,i,i)
   read='view%d.%s'%(i,field)+('[0]' if element else '')
   if absent:text+='console.log(`${'+read+'===undefined}`);\n'
   else:
    text+='const value%d=%s;\n'%(i,read)
    if nested:text+='if(value%d!==undefined){console.log(`${value%d.payload.label}`);}\n'%(i,i)
    else:text+='console.log(`${value%d!==undefined}`);\n'%i
  return text
 values=[selector['values'][0] for selector in selectors]
 open_kind=any(selector['type'] in ['number','string'] for selector in selectors)
 wrong=True if open_kind else 999
 variants={'good':source(values),'wrong':source([wrong]),'nested':source([values[0]],nested=True)}
 if nullable and not element:variants['absent']=source([None],absent=True)
 for variant,text in variants.items():(folder/('pair-%d-%s.a'%(row['rank'],variant))).write_text(text)
 manifest.append({'rank':row['rank'],'type_id':row['type_id'],'receiver':row['type'],'field':row['field'],'declared_type':row['declared_type'],'source_alias':alias,'candidate_reads':row['read_count'],'variants':list(variants),'scope':'measured own-kind selectors and transitive payload checks; independently written projection, not full compiler interfaces'})
(root/'candidate-fixtures.json').write_text(json.dumps(manifest,indent=2)+'\n')
print('generated',len(manifest),'candidate fixtures,',sum(r['candidate_reads'] for r in manifest),'candidate reads')

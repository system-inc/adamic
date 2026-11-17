#!/usr/bin/env python3
"""Classify observed runCode rethrows separately; retain original aggregate unchanged."""
from pathlib import Path
import hashlib,json
p=Path(__file__).resolve().parent/'evidence';d=json.loads((p/'counts.json').read_text());output={}
for label,r in d.items():
 classifications=[]
 for event in r['throws']:
  kind=event['classification'];proof=event.get('classification_evidence');matched=[]
  if kind=='unclassified' and event['frames'][0]['function']=='runCode':
   matched=[i for i,x in enumerate(r['throws']) if x['classification']=='freeze-write' and x['pid']==event['pid'] and x.get('case')==event.get('case') and x['description']==event['description']]
   if matched:
    kind='propagated-freeze-error';proof='Identical process, case and complete error stack to a provenance-backed write. src/harness/fourslashImpl.ts runCode catch accesses err.stack then executes throw err; no assignment at this observation. See early-write-observation.json source snippet.'
  classifications.append({'original_classification':event['classification'],'classification':kind,'description_sha256':hashlib.sha256(event['description'].encode()).hexdigest(),'pid':event['pid'],'case':event.get('case'),'top_frame':event['frames'][0],'matched_freeze_write_indices':matched,'evidence':proof})
 output[label]={'events':classifications,'counts':{kind:sum(x['classification']==kind for x in classifications) for kind in sorted({x['classification'] for x in classifications})},'unresolved':[i for i,x in enumerate(classifications) if x['classification']=='unclassified']}
(p/'exception-classification.json').write_text(json.dumps(output,indent=2)+'\n')
for label,r in output.items():print(label,r['counts'],'unresolved',len(r['unresolved']))

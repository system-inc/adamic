#!/usr/bin/env python3
"""Validate saved measurements independently; reject fabricated clearances."""
import copy,json,pathlib,re
HERE=pathlib.Path(__file__).resolve().parent
old=json.loads((HERE/'baseline-67.json').read_text())
def validate(data):
 assert data['combined_status'].startswith('unmeasured:')
 assert len(old)==67 and len({x['id'] for x in old})==67
 for split in (0,1):
  text=(HERE/'evidence'/f'step24-remaining-parser-{split}.log').read_text()
  ids=[f'{f}:{l}:{c}:TS{code}' for f,l,c,code in re.findall(r'/src/compiler/(.+?):(\d+):(\d+): error TS(\d+):',text)]
  record=data['main_control'][str(split)]
  assert record['identities']==ids and record['count']==len(ids)
  assert record['historical_present']==[x['id'] for x in old if x['id'] in ids]
  assert record['historical_absent']==[x['id'] for x in old if x['id'] not in ids]
  assert record['first']==ids[0]
data=json.loads((HERE/'comparison.json').read_text());validate(data)
for mutation in ('clearance','count','first','combined'):
 bad=copy.deepcopy(data)
 if mutation=='clearance':bad['main_control']['0']['historical_present'].pop()
 if mutation=='count':bad['main_control']['1']['count']+=1
 if mutation=='first':bad['main_control']['0']['first']='fake:1:1:TS0000'
 if mutation=='combined':bad['combined_status']='passed'
 try:validate(bad)
 except AssertionError:print('caught',mutation)
 else:raise AssertionError('audit mutant survived: '+mutation)
print('PASS: both raw logs validate; four evidence mutants rejected')

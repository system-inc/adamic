#!/usr/bin/env python3
"""Reconcile independent API and main proof inventories, with failing ledger mutants."""
import copy,json,sys
from collections import Counter
m=json.load(open(sys.argv[1]));p=json.load(open(sys.argv[2]));d=json.load(open(sys.argv[3]));tree=sys.argv[4].rstrip('/')+'/'
def audit(m,p,d):
 c=m['counts'];assert c['typescriptFiles']==len(m['files'])==p['roots']==79
 assert c['regularFiles']==len(m['regularFiles'])==82
 assert len(m['casts'])==c['unknownDoubleCasts']+c['anyDoubleCasts']==18
 assert c['unknownDoubleCasts']==sum(x['bridge']=='unknown' for x in m['casts'])==5
 assert c['anyDoubleCasts']==sum(x['bridge']=='any' for x in m['casts'])==13
 a={x['location']:x for x in m['predicates']};b={x['location'].removeprefix(tree):x for x in p['predicates']};assert set(a)==set(b) and len(a)==651
 assert Counter(x['status'] for x in b.values())=={'Proven':3,'Refused':648}
 assert sum(x['hasBody'] for x in a.values())==580
 assert b['src/compiler/debug.ts:365:46']['status']=='Refused'
 assert b['src/compiler/core.ts:1769:42']['status']=='Proven'
 assert b['src/compiler/core.ts:1773:39']['status']=='Proven'
 assert b['src/compiler/watchPublic.ts:739:77']['status']=='Proven'
 assert len(m['functionValues'])==1 and m['functionValues'][0]['expression']=='Function.prototype'
 assert len(d['casts'])==18 and all(x['status']=='Refused' for x in d['casts'])
 assert len(m['writes'])==c['propertyWrites']==3011
 assert len(m['descriptors'])==9 and len(m['assigns'])==1
 assert m['assigns'][0]['existing']
audit(m,p,d)
for name,change in [
 ('missing double cast',lambda a,b,c:a['casts'].pop()),
 ('empty assertion accepted',lambda a,b,c:next(x for x in b['predicates'] if x['location'].endswith('/debug.ts:365:46')).update(status='Proven')),
 ('invented checked direct cast',lambda a,b,c:c['casts'][0].update(status='Checked')),
]:
 mm,pp,dd=copy.deepcopy((m,p,d));change(mm,pp,dd)
 try:audit(mm,pp,dd)
 except AssertionError:print('CAUGHT:',name)
 else:raise AssertionError('surviving ledger mutant: '+name)
print('PASS: API/main proof join, scope, counts, positive/negative controls, 3 ledger mutants')

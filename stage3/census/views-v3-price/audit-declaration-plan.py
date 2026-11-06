#!/usr/bin/env python3
from copy import deepcopy
import json
from declaration_common import load_plan,audit_partition,split_piece
plan=load_plan();rejected=[]
for name,mutant in [('lost',plan['pieces'][:-1]),('duplicate',plan['pieces']+[plan['pieces'][0]])]:
 try:audit_partition(plan,mutant)
 except AssertionError:rejected.append(name)
 else:raise AssertionError(name+' mutant survived')
mutant=deepcopy(plan['pieces']);a,b=next((a,b) for a in mutant for b in mutant if a!=b and a['expected_units'] and b['expected_units'])
a['expected_units'][0],b['expected_units'][0]=b['expected_units'][0],a['expected_units'][0]
try:audit_partition(plan,mutant)
except AssertionError:rejected.append('moved with unchanged global inventory')
else:raise AssertionError('moved mutant survived')
parent=next(p for p in plan['pieces'] if len(p['top_atoms'])>1 and not p['seed']);children=split_piece(plan,parent)
active=[]
for p in plan['pieces']:active.extend(children if p['id']==parent['id'] else [p])
audit_partition(plan,active)
print(json.dumps(dict(status='passed',files=plan['files'],tops=plan['top_units'],full_units=plan['units'],pieces=len(plan['pieces']),mutants_rejected=rejected,contiguous_refinement_preserves_every_unit=True),indent=2))

#!/usr/bin/env python3
"""Prove exact coverage and admission audits reject independent corruptions."""
import copy, importlib.util, sys
from pathlib import Path
spec=importlib.util.spec_from_file_location('delegation_audit',Path(__file__).with_name('audit-delegation.py'))
module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
inputs=list(map(module.read,sys.argv[1:6]));tree=Path(sys.argv[6])
for name in ['drop-body','drop-call','invent-admission','source-hash']:
    baseline,before,after,optional,original=copy.deepcopy(inputs)
    if name=='drop-body':after['predicates'].pop()
    elif name=='drop-call':next(row for row in after['predicates'] if row['calls'])['calls'].pop()
    elif name=='invent-admission':next(row for row in after['predicates'] if not row['bodyProof'])['admission']='Proven'
    else:baseline['sourceHashes'][next(iter(baseline['sourceHashes']))]='0'*64
    try:module.audit(baseline,before,after,optional,original,tree)
    except AssertionError as error:print(name+': caught: '+str(error))
    else:raise AssertionError('mutant escaped: '+name)
print('all four measurement mutants caught')

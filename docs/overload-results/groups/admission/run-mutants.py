#!/usr/bin/env python3
"""Prove return covariance, visitor contravariance and relation diagnostics."""
import json, os, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[4]
out=Path(__file__).parent
call=root/'internal/lower/overload_return_call.go'
census=root/'internal/lower/census_small.go'
a=call.read_text();b=census.read_text()
parameter='if !l.censusRelated(given, takes) && !l.overloadDeferredParameter(given, takes) && !l.overloadNullableParameter(given, takes) {'
mutants=[
 ('drop-fixed-return-proof',call,a.replace('return l.censusRelated(produced, promised)','_ = produced; return false'),'TestOverloadStructuralResults','result Block'),
 ('reverse-result-covariance',call,a.replace('return l.censusRelated(produced, promised)','return l.checker.IsTypeAssignableTo(promised, produced)'),'TestOverloadStructuralRefuses/factory_literal_covariance','got <nil>'),
 ('trust-visitor-input',census,b.replace(parameter,parameter[:-1]+'&& len(l.checker.GetSignaturesOfType(l.concrete(given), checker.SignatureKindCall)) == 0 {'),'TestOverloadStructuralRefuses/visitor_domain_liar','got <nil>'),
 ('trust-writable-result',call,a.replace('return l.censusRelated(produced, promised)','return l.classAssignable(produced, promised)'),'TestOverloadStructuralRefuses/factory_writable_invariance','got <nil>'),
 ('drop-return-storage',call,a.replace(' && l.overloadReturnStorage(produced, promised)',''),'TestOverloadStructuralRefuses/factory_field_storage','got <nil>'),
 ('drop-relation-diagnostic',census,b.replace(' + "; " + l.overloadResultRelation(produced, promised)',''),'TestOverloadStructuralRefuses/factory_result_covariance','wanted refusal naming readonly covariance'),
]
rows=[]
for name,source,changed,selector,marker in mutants:
 assert changed!=source.read_text(),name
 with tempfile.TemporaryDirectory(prefix='overload-admission-mutant-') as tmp:
  tmp=Path(tmp);replacement=tmp/source.name;replacement.write_text(changed)
  overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(replacement)}}))
  selection='/'.join('^'+part+'$' for part in selector.split('/'))
  command=['go','test','-overlay',str(overlay),'./internal/lower','-run',selection,'-count=1','-v']
  logpath=out/(name+'.log.txt')
  with logpath.open('w') as log: result=subprocess.run(command,cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
  text=logpath.read_text();caught=result.returncode!=0 and '--- FAIL:' in text and '[build failed]' not in text and marker in text
  rows.append(dict(mutant=name,caught=caught,exit=result.returncode,selector=selection));print(name,'caught' if caught else 'NOT CAUGHT',flush=True)
(out/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(row['caught'] for row in rows)

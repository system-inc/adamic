#!/usr/bin/env python3
"""Attribute production-mode program refusals to all exact indexed ledger rows."""
import argparse
import csv
import json
from collections import Counter
from pathlib import Path

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--result',type=Path,required=True)
parser.add_argument('--tree',type=Path,required=True)
parser.add_argument('--output',type=Path,required=True)
args=parser.parse_args()
root=Path(__file__).resolve().parent
rows=list(csv.DictReader((root/'evidence/ledger-rows.csv').open()))
indexed={r['id']:r for r in rows if r['option']=='noUncheckedIndexedAccess'}
excluded={r['id']:r for r in rows if r['option'] in ['exactOptionalPropertyTypes','useUnknownInCatchVariables']}
if len(indexed)!=99 or len(excluded)!=72: raise SystemExit('dated ledger coverage changed')
result=json.loads(args.result.read_text())
expected_roots={str(x) for x in json.loads((root/'evidence/ledger-options.json').read_text())['project-loader']['roots'] if not x.endswith('.d.ts')}
if len(result['roots'])!=79 or set(result['roots'])!=expected_roots: raise SystemExit('original 79-root coverage changed')
if len(result['excluded'])!=72 or {r['id'] for r in result['excluded']}!=set(excluded): raise SystemExit('excluded row IDs differ from the exact 72')
for r in result['excluded']:
 original=excluded[r['id']];site=r['site']
 if site!=dict(file=str(args.tree.resolve()/original['file']),line=int(original['line']),column=int(original['column']),code=int(original['code'][2:]),message='',options=[original['option']]):
  # Go omits no fields from OptionSite; its unused message is serialized empty.
  raise SystemExit('exclusion identity/attribution changed: '+r['id'])
if result['stage']!='lower' or result['diagnostics'] or not result['requires_indexed_presence'] or result['indexed_audit_rows']!=99: raise SystemExit('production conversion did not reach lowering with all 99 indexed rows')
audit=[s for s in result['option_sites'] if 'noUncheckedIndexedAccess' in s['options']]
expected_audit={(str(args.tree.resolve()/r['file']),int(r['line']),int(r['column']),int(r['code'][2:])) for r in indexed.values()}
actual_audit={(s['file'],s['line'],s['column'],s['code']) for s in audit}
if len(audit)!=99 or actual_audit!=expected_audit:raise SystemExit('indexed audit identities differ from the 99 dated rows')
entries={a['entry']:a for a in result['attempts']}
expected_entries={'src/compiler/_namespaces/ts.ts'}|{r['file'] for r in indexed.values()}
if len(result['attempts'])!=len(expected_entries) or set(entries)!=expected_entries: raise SystemExit('missing, duplicate or extra original lowering entries')
for a in entries.values():
 if a['stage']!='lower' or not a.get('error') or a['error_kind'] not in ['Refused','NotYet'] or a['checks'] or a.get('binary'):
  raise SystemExit('new lowering/native outcome requires direct source-to-guard classification')
results=[]
for identity,row in sorted(indexed.items(),key=lambda pair:int(pair[0][1:])):
 a=entries[row['file']]
 results.append({**row,'status':'still refuses','refusal_scope':'original file dependency graph, before indexed-site lowering','reason':a['error'],'error_kind':a['error_kind'],'entry':a['entry'],'checker_error':False,'indexed_check_emitted':False,'read_local_outcome':'not reached'})
args.output.mkdir(parents=True,exist_ok=True)
(args.output/'production-indexed-status.json').write_text(json.dumps(results,indent=2)+'\n')
fields=['id','file','line','column','owner','source_expression','status','refusal_scope','reason','error_kind','checker_error','indexed_check_emitted','read_local_outcome']
with (args.output/'production-indexed-status.csv').open('w',newline='') as output:
 writer=csv.DictWriter(output,fieldnames=fields,extrasaction='ignore',lineterminator='\n');writer.writeheader();writer.writerows(results)
summary={'mode':result['mode'],'roots':79,'excluded_optional_rows':67,'excluded_catch_rows':5,'checker_diagnostics':0,'indexed_audit_rows':99,'entry_attempts':len(entries),'compiles_as_check':0,'still_refuses':99,'still_errors':0,'refusal_scope':'all entry dependency graphs refuse before indexed-site lowering','native_artifacts':0,'read_local_outcomes_unmeasured':99,'blockers':dict(Counter(a['error'] for a in entries.values()))}
(args.output/'production-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary))

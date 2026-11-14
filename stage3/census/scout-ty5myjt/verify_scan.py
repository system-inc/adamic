"""Verify the archived scan-only continuation without rerunning the compiler."""
from pathlib import Path
import json,gzip,re,subprocess
p=Path(__file__).resolve().parent;repo=p.parents[2]
runs=json.load(open(p/'scan-runs.json'));assert len(runs)==8
expected={'baseline':81,'diag-readonly':83,'diag-readonly-sorter':81,'diag-intersection':81,'diag-exact':81,'diag-overloads':79}
for run in runs:
 assert run['mode']=='scan only' and run['units']==0 and run['exit']==0 and not run['timeout']
 if run['file']=='program.ts':assert run['findings']==expected[run['candidate']]
 else:assert run['findings']==({'baseline':55,'diag-readonly-sorter':57}[run['candidate']])
for n,new,gone in [('diag-readonly',4,2),('diag-readonly-sorter',2,2),('diag-intersection',2,2),('diag-exact',2,2),('diag-overloads',0,2),('diag-readonly-sorter-utilitiesPublic',2,0)]:
 d=json.load(open(p/(n+'-scan-delta.json')));assert len(d['new'])==new and len(d['gone'])==gone and not d['changed']
for a in p.glob('scan-*.jsonl.gz'):
 data=[json.loads(x) for x in gzip.open(a,'rt')];assert len(data)==3 and data[-1]['status']=='relation_report' and data[0]['latent_mode']=='scan only' and data[0]['module_format']=='CommonJSBundle';assert data[1]['units']==[] and all(x['phase']=='refusal_scan' for x in data[1]['findings'])
 if a.name=='scan-baseline-program.ts.jsonl.gz':
  for line in [2796,2829]:assert any(x['where'].endswith(f'program.ts:{line}:16') and 'DiagnosticWithLocation seen as Diagnostic,' in x['reason'] for x in data[1]['findings'])
assert len(json.load(open(p/'diag-readonly-sorter-stock.json'))['rows'])==0
assert all(x['identical'] for x in json.load(open(p/'diag-readonly-sorter-javascript.json')))
pin=(p/'scan-wrapper-pin.txt').read_text().strip();wrapper=subprocess.check_output(['git','show',pin+':stage3/census/promise-discard-sites/scan-only.go.txt'],cwd=repo);assert wrapper==(p/'scan-only.go.txt').read_bytes()
for label,link in re.findall(r'\[([^]]+)\]\(([^)]+)\)',(p/'REPORT.md').read_text()):assert (p/link).exists(),link
assert not list(p.glob('*.go'))
for b in json.load(open(p/'sorter-call-chain.json')):assert not re.search(r'\.file\s*=(?!=)',b['text']),b['name']
result={'status':'pass','program_mode':'scan only','complete_program_runs':6,'complete_sorter_file_runs':2,'units_each':0,'baseline_targets_present':True,'overloads':{'before':81,'after':79,'new':0,'gone':2,'changed':0},'readonly_file_only':'two new numeric target refusals plus two helper handoff refusals','readonly_file_plus_sorter':'two numeric target refusals, two new comparator refusals in utilitiesPublic.ts','sorter_unique':{'before':34,'after':36},'sorter_raw':{'before':55,'after':57},'new_candidate_stock_diagnostics':0,'new_candidate_js_identical':True,'wrapper_verbatim':pin,'production_sources_changed':False}
(p/'scan-verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))

from pathlib import Path
import json,subprocess,hashlib
s=Path('/workspace/scratch/diag-read-view-1');r=s/'repo';raw=s/'raw'
p=s/'stop-type-facts.cjs';p.write_text(p.read_text().replace('relatedInformationType:', 'comparatorParameterType:'))
p=raw/'stop-type-facts.json';facts=json.loads(p.read_text())
for k in ['currentTwoComparatorParameters','virtualAllEightComparatorParameters']:
 if 'relatedInformationType' in facts[k]:facts[k]['comparatorParameterType']=facts[k].pop('relatedInformationType')
p.write_text(json.dumps(facts,indent=2)+'\n')
pair=json.loads((raw/'pair-candidate.json').read_text());sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=r,text=True).strip();parent=subprocess.check_output(['git','rev-parse','HEAD^'],cwd=r,text=True).strip()
assert parent=='8e3a784aa5a601f57342c07b68f145fe56252c21'
results={p.name:json.loads(p.read_text()) for p in raw.glob('*census.result.json')};assert len(results)==9 and all(x['exit']==0 and x['seconds']<6000 for x in results.values())
summary={'status':'STOPPED_UNACCEPTED','candidate':sha,'parent':parent,'comparatorsChanged':['compareDiagnostics','diagnosticsEqualityComparer'],'censuses':{f:{k:len(x[k]) for k in ['new','changed','removed']} for f,x in pair.items()},'fullCensusRuns':9,'allFullCensusRunsExitZeroWithin100Minutes':True,'reason':'The optional-widening findings at mapped utilities.ts:8808:39 and :8808:44 change their reason target from Diagnostic to DiagnosticReadView. All allowed comparator parameter substitutions preserve DiagnosticRelatedInformation recursive arguments without reportsUnnecessary and the DiagnosticReadView target that has it. Removing or reclassifying these changed findings is outside the specified shape and strict pairing rule.','lane':'Cancelled under the user stop condition during normal apply; oracle was not run. No final apply or 106366-lane pass is claimed.','firstApply':'Stopped at the existing whole utilities import pin; corrected by an amendment in core-multimap and verified at that owner boundary.','implementationBranch':'typescript/diag-read-view-1: unaccepted review candidate, no acceptance or Loom submission','ledger':'Unchanged; no README convention requires a new diagnostic-amendment ledger entry.','laneInputs':'All 29 sanctioned proof inputs unchanged; no lane regeneration required.','toolSha256':{name:hashlib.sha256((s/name).read_bytes()).hexdigest() for name in ['latent','adamic']}}
(raw/'STOP.json').write_text(json.dumps(summary,indent=2)+'\n')
lines=['# STOP: diagnostic read-view amendment is not accepted','',f'Candidate: {sha}; parent: {parent}. One commit on the requested base.','',
'The candidate and evidence are published under the requested branch names for review only. No Loom submission. The 0-new/0-changed gate fails; no successful final apply or whole-lane result is claimed.','',
'## Full censuses','',
'Every run used LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, exact LATENT_ONLY_FILE, CENSUS_MODULE_FORMAT=CommonJSBundle and a 6000-second bound. All eight main/candidate runs and the full restoring-mutant program run completed normally. Complete streams include the final relation_report record.','',
'| File | Main unique | Candidate unique | New | Changed | Removed | Main seconds | Candidate seconds |','| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |']
for f,x in pair.items():
 a=results['main-'+f+'-census.result.json']['seconds'];b=results['candidate-'+f+'-census.result.json']['seconds']
 lines.append(f'| {f} | {x["before"]} | {x["after"]} | {len(x["new"])} | {len(x["changed"])} | {len(x["removed"])} | {a:.3f} | {b:.3f} |')
lines += ['', 'Keys are phase, kind, mapped where and reason. Checkout roots and embedded source locations in reasons are mapped; no type names or other reason text are normalized. The outer-location-only report retains the two path/line-only program differences. Both original program refusal_scan sites, 2796:16 and 2829:16, disappear.','',
'## Why this stops inside the requested shape','',summary['reason'],'',
'The related-information array field and recursive expression compareDiagnostics(d1i, d2i) do not change under the permitted parameter substitutions. A read-only stock-checker counterfactual with all eight comparator signatures confirms identical argument and target types and missing/present reportsUnnecessary properties. This explains the full-census result; it is not another census. See stop-type-facts.json.','',
'Only compareDiagnostics and diagnosticsEqualityComparer were changed in the measured candidate. The ten new readonly-to-writable findings identify compareDiagnosticsSkipRelatedInformation, compareRelatedInformation, getDiagnosticCode, getDiagnosticMessage and getDiagnosticFilePath as further callees needing the view if work continued. No further signatures, members, call expressions, overloads or diagnostic policy were changed after the stop.','',
'## Every new finding in full','']
for f,x in pair.items():
 for n,v in enumerate(x['new'],1):
  lines += [f'### {f}, new {n}',f'Phase: {v["phase"]}. Kind: {v["kind"]}. Where: {v["where"]}.','',v['text'],'']
lines += ['## Every changed finding in full','']
for f,x in pair.items():
 for n,v in enumerate(x['changed'],1):
  lines += [f'### {f}, changed {n}',f'Mapped key: {v["mappedSite"]}.','']
  for label in ['before','after']:
   lines.append(label+':')
   for finding in v[label]:lines += [f'Phase: {finding["phase"]}. Kind: {finding["kind"]}. Where: {finding["where"]}.','',finding['text'],'']
lines += ['## Other proof results','',
'- Restoring mutant: exact original bytes restored in all four diagnostic files, and both original refusal_scan findings returned in its full program census (2710.820 seconds). Historical resolution-file restoration is retained.',
'- Stock TypeScript 6.0.3: 0 diagnostics on both main and candidate over 79 roots. Adamic types: exit 0, 0 diagnostics over 79 roots.',
'- Upstream npm ci and node_modules/.bin/hereby dts completed in each adapted checkout. typescript.d.ts and tsserverlibrary.d.ts are byte-identical; hashes are in declarations.json.',
'- All 604 JavaScript-emitting sources are byte-identical under transpilation in both comment modes. Both upstream compiler builds passed; all 9 actual JavaScript outputs are byte-identical.',
'- Drift, duplicate, partial-tree, mixed-line-ending, LF/CRLF, exact-restoration and zero-write replay controls passed with final owner routing.',
'- The resolution engine, mutant and original pin entry remain byte-identical. Program owner amendment: 95-checker-axolotl-h. Whole utilities import amendment: 99-correctness-core-multimap, with exact input/output variants.',
'- All 29 sanctioned API proof inputs are unchanged. No sanctioned-api.json, shard-plan runtime hash or generated shard test regeneration is required. No ledger line was added.',
'- First full candidate apply stopped at the later whole import owner. Its amendment was then validated at that owner boundary (17 historical runtime files, 40 historical edits). Final fresh whole lane was cancelled during apply on the census stop, before oracle tests. Final normal apply and 106366 passing are NOT proven.',
'- Additional stage-specific final-tree replay checks preserve existing refusals: 95 rejects identical normalized baseline/candidate createProgram hashes after later adapters; MultiMap rejects an unchanged later recursive-watcher sys owner. The amended utilities import owner reports adapted=true and zero edits. See the replay JSON records. No unrelated owner gate was widened.','',
'See STOP.json, pair-candidate.json, controls.json, declarations.json, source-manifests.json, every result.json and the complete jsonl.gz streams.']
(raw/'STOP.md').write_text('\n'.join(lines)+'\n')
print(json.dumps(summary,indent=2))

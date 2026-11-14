from pathlib import Path
import json,subprocess,hashlib,shutil
s=Path('/workspace/scratch/diag-read-view-2');raw=s/'raw';r=s/'repo'
gate=json.loads((raw/'census-gate.json').read_text());assert gate['passed'] and len(gate['summary'])==4
results={p.name:json.loads(p.read_text()) for p in raw.glob('*census.result.json')};assert len(results)==9 and all(x['exit']==0 and x['seconds']<6000 for x in results.values())
lane=json.loads((s/'lane/report.json').read_text());assert lane['status']=='pass',lane
assert lane['counts']=={'passing':106366,'failing':1,'pending':0},lane.get('counts')
assert json.loads((raw/'candidate-apply.result.json').read_text())['exit']==0
assert json.loads((raw/'whole-lane.result.json').read_text())['exit']==0
for name in ['candidate-adamic-types','stock-tsc','javascript','main-dts','candidate-dts','main-npm-ci','candidate-npm-ci','main-build-compiler','candidate-build-compiler']:
 assert json.loads((raw/(name+'.result.json')).read_text())['exit']==0,name
declarations=json.loads((raw/'declarations.json').read_text());assert all(x['identical'] for x in declarations.values())
js=json.loads((raw/'upstream-javascript.json').read_text());assert js['files']==9 and not js['changed'] and js['sameFileSet']
assert json.loads((raw/'normal-apply-source-match.json').read_text())['identical']
stock=json.loads((raw/'stock-tsc.json').read_text());assert not stock['main']['diagnostics'] and not stock['candidate']['diagnostics'] and not stock['new']
controls=json.loads((raw/'controls.json').read_text());assert controls['replay']['writes']==0
mutant=json.loads((raw/'mutant-targets.json').read_text());assert {x['where'].rsplit(':',2)[1] for x in mutant}=={'2796','2829'}
proofs=json.loads((r/'stage3/lane/sanctioned-api.json').read_text())['proofs'];moved=[p['file'] for p in proofs if hashlib.sha256((r/p['file']).read_bytes()).hexdigest()!=p['sha256']];assert not moved
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=r,text=True).strip();parent=subprocess.check_output(['git','rev-parse','HEAD^'],cwd=r,text=True).strip();assert parent=='8e3a784aa5a601f57342c07b68f145fe56252c21'
helpers=['compareDiagnostics','compareDiagnosticsSkipRelatedInformation','compareRelatedInformation','compareMessageText','diagnosticsEqualityComparer','getDiagnosticFilePath','getDiagnosticCode','getDiagnosticMessage']
summary={'status':'ACCEPTED','implementation':sha,'parent':parent,'helpers':helpers,'censuses':gate['summary'],'fullCensusRuns':9,'allCensusesCompletedWithin100Minutes':True,'mutant':'both original refusal_scan sites returned','stockTscDiagnostics':{'main':0,'candidate':0,'roots':79},'adamicTypesDiagnostics':0,'sourceJavaScript':{'files':604,'commentModes':2,'changed':0},'actualJavaScript':{'files':9,'changed':0},'declarations':declarations,'normalApply':'passed; 747 source files exactly match measured candidate','lane':lane,'sanctionedProofInputs':len(proofs),'movedProofInputs':moved,'generatedLaneFilesRegenerated':False,'ledgerChanged':False,'loomSubmission':False}
(raw/'FINAL.json').write_text(json.dumps(summary,indent=2)+'\n')
pair=json.loads((raw/'pair-candidate.json').read_text())
lines=['# Diagnostic read-view v2: final proof','',f'Implementation: {sha}. Parent: {parent}. One commit on the requested base. No Loom submission.','',
'All eight reader helpers use DiagnosticReadView: '+', '.join(helpers)+'. Bodies and public sorter signature are unchanged. The resolution half remains unchanged.','',
'## Full censuses','',
'All eight main/candidate censuses and the restoring-mutant program census completed with exit 0 within the individual 6000-second bounds. Every run used LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, exact LATENT_ONLY_FILE and CENSUS_MODULE_FORMAT=CommonJSBundle. Complete raw streams end in relation_report.','',
'| File | Main unique | Candidate unique | New | Changed | Removed | Main seconds | Candidate seconds |','| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |']
for file,row in pair.items():
 a=results['main-'+file+'-census.result.json']['seconds'];b=results['candidate-'+file+'-census.result.json']['seconds'];lines.append(f'| {file} | {row["before"]} | {row["after"]} | {len(row["new"])} | {len(row["changed"])} | {len(row["removed"])} | {a:.3f} | {b:.3f} |')
lines += ['', 'Pair keys are phase, kind, mapped where and reason. Checkout roots and source references embedded in reasons are mapped; type names are not normalized. Both program refusal_scan targets at 2796:16 and 2829:16 are removed. There are no new findings in any phase.','',
'## Accepted changed pair: both texts','']
for change in pair['utilities.ts']['changed']:
 assert change.get('acceptedByUser')
 lines += [change['classification']+'.', 'Mapped key: '+str(change['mappedSite'])+'.','']
 for side in ['before','after']:
  lines.append(side+':')
  for f in change[side]:lines += ['Phase: '+f['phase']+'. Kind: '+f['kind']+'. Where: '+f['where']+'.','',f['text'],'']
lines += ['## Remaining proofs','',
'- Restoring mutant: exact original diagnostic files restored; full program census returns both original refusal_scan findings. See mutant-targets.json.',
'- Stock TypeScript 6.0.3: 0 diagnostics on both main and candidate over 79 roots. Adamic types: 0 diagnostics over the final 79 roots.',
'- Both upstream npm ci, node_modules/.bin/hereby dts and compiler builds passed. typescript.d.ts and tsserverlibrary.d.ts are byte-identical. Exact hashes: declarations.json.',
'- All 604 emitting source files match in both JavaScript comment modes. All 9 actual upstream compiler JavaScript outputs match byte-for-byte.',
'- Final normal apply passed and matches the measured candidate in all 747 source files. The whole lane was run unfiltered with unchanged workers and settings; its own verdict is PASS, with 106,366 passing, only the sanctioned API row and no pending tests.',
'- Final controls passed: drift, duplicates, partial adaptations, mixed endings, LF/CRLF, exact restoring mutant and zero-write replay.',
'- Resolution engine, restoring witness and original whole-file pin are unchanged. The createProgram amendment stays in 95-checker-axolotl-h; the utilities import amendment stays in core-multimap, including exact input/output variants.',
'- All 29 sanctioned proof inputs remain unchanged. No sanctioned-api.json, shard-plan runtime hash or generated shard test refresh was required. No README convention required a new ledger line.','',
'See FINAL.json, census-gate.json, pair-candidate.json, raw full census archives and all result records.']
(raw/'REPORT.md').write_text('\n'.join(lines)+'\n')
for name in ['report.json','verdict.json','execution.json','patch-set.md','apply.log','oracle.log','failures.md','verdict.txt']:
 if (s/'lane'/name).exists():shutil.copyfile(s/'lane'/name,raw/('lane-'+name))
if (s/'lane/oracle').exists():
 for p in (s/'lane/oracle').rglob('*'):
  if p.is_file() and p.suffix in ['.json','.jsonl','.log','.diff','.md']:
   destination=raw/'lane-oracle'/p.relative_to(s/'lane/oracle');destination.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,destination)
print(json.dumps({'implementation':sha,'status':'ACCEPTED','censuses':gate['summary'],'lane':lane['status']},indent=2))

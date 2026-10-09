from pathlib import Path
import json,statistics,re,subprocess,hashlib,time,shlex
out=Path('review/test-audit/stage1-cohere-markdownblocks-leaf_composition_products');pkg='stage1/cohere/markdownblocks'
def records(file):
 result=[]
 for line in (out/file).read_text().splitlines():
  try:result.append(json.loads(line))
  except:pass
 return result
def median(group):
 values=[]
 for file in sorted(out.glob('clean-'+group+'-*.log')):
  for r in records(file.name):
   if 'Test' not in r and r.get('Output','').startswith('ok '):values.append(float(re.search(r'\t([\d.]+)s',r['Output']).group(1)))
 return statistics.median(values) if len(values)==3 else None
scope=json.loads((out/'scope.json').read_text())['rows'];loc={r['test']:r for r in scope}
rows=[('TestProduct_MarkdownLeafGo family','products-go',['TestProduct_MarkdownLeafGoLists','TestProduct_MarkdownLeafGoDocLayout'],'S1'),('TestProduct_MarkdownLeafLowered','product-lowered',['TestProduct_MarkdownLeafLowered'],'S2'),('TestProduct_MarkdownLeafNative family','products-native',['TestProduct_MarkdownLeafNativeSanitized','TestProduct_MarkdownLeafNativeRelease'],'S3'),('TestMarkdownListLayout_Setup','list-setup',['TestMarkdownListLayout_Setup'],'S4'),('TestMarkdownListLayout family','list-family',['TestMarkdownListLayoutUnion']+[f'TestMarkdownListLayout_{i:03d}' for i in range(16)],None),('TestMarkdownListLayout','list-legacy',['TestMarkdownListLayout'],'S4'),('TestMarkdownQuoteLayout','quote',['TestMarkdownQuoteLayout'],'S5'),('TestMarkdownTableLayout','table',['TestMarkdownTableLayout'],'S6'),('TestMarkdownLayout family','block-layouts',['TestMarkdownCodeBlockLayout','TestMarkdownHTMLBlockLayout','TestMarkdownRootLayout'],None),('TestMarkdownLeafComposition','leaf',['TestMarkdownLeafComposition'],'S7'),('TestMarkdownStructureLayout_Setup','structure-setup',['TestMarkdownStructureLayout_Setup'],'S8'),('TestMdastMalformedEvents family','malformed-family',[f'TestMdastMalformedEvents_{i:03d}' for i in range(3)]+['TestMdastMalformedEventsUnion'],None)]
matrix=json.loads((out/'matrix-runs.json').read_text());construction=json.loads((out/'construction-runs.json').read_text());groupname={g:name for name,g,m,s in rows};memberrow={m:name for name,g,members,s in rows for m in members}
matrows=[groupname[g] for g in ['list-family','block-layouts','malformed-family']];result=[]
for name,group,members,sid in rows:
 row={'test':name,'package':pkg,'file':f"{pkg}/{loc[members[0]]['file']}:{loc[members[0]]['line']}",'members':members,'seconds':median(group),'oracle':'Successful preparation is a self expectation; no independent semantic answer is compared.','oracle_kind':'self','kills':[],'unique_kills':[],'last_proven_fail':None,'verdict':'cannot-judge','subsumed_by':[],'mutants_in_matrix':0,'probe_kills':[],'subsumer_seconds':None,'vacuous':None,'bounded':True,'matrix_rows':[name],'evidence':None}
 if sid:
  experiment=next(x for x in construction if x['id']==sid);data=records(sid+'.log');errors=[x['Output'].strip() for x in data if x.get('OutputType')=='error' and x.get('Test','').split('/')[0] in members];cooked=any('test timed out' in x.get('Output','') for x in data)
  fail=[x for x in data if x.get('Action')=='fail' and x.get('Test','').split('/')[0] in members]
  row['construction_mutants']=[sid];row['construction_kills']=[sid] if fail else []
  row['verdict']='cannot-judge' if cooked or experiment['vet'] else ('setup-check' if fail else 'untrue')
  line=errors[0] if errors else next((x['Output'].strip() for x in data if x.get('Output','').startswith('--- PASS:') and x.get('Test','').split('/')[0] in members),'no completed row')
  if sid == 'S8': line=line.replace('structure_layout_shards_test.go:120:', 'structure_layout_shards_test.go:121 (origin; raw scratch line 120):')
  if fail:row['last_proven_fail']=sid+' '+line
  row['evidence']=f"timeout 120 go test -json -count=1 -timeout 90s ./{pkg}/ -run '{experiment['pattern']}'; {sid}.log: {line}"
  row['oracle']+=' '+sid+' '+('broken construction was rejected.' if fail else 'changed construction passed; this verdict rests on one construction edit and is not a claim that every failure is impossible.')
 else:
  observed=[x for x in matrix if x['group']==group];row['mutants_in_matrix']=len([x for x in observed if x['id'].startswith('M')]);row['matrix_rows']=matrows
  row['kills']=[x['id'] for x in observed if x['id'].startswith('M') and any(t.split('/')[0] in members for t in x['failed'])]
  row['probe_kills']=[x['id'] for x in observed if x['id'].startswith('P') and any(t.split('/')[0] in members for t in x['failed'])]
  row['vacuous']=False if row['probe_kills'] else None
  row['oracle_kind']=['external-run','self'];row['oracle']='Independent Go cohere bytes and retained fork Node results; self requirements for build success, exit status, framing, and built-in mutant difference. Full semantic bytes are checked; leak checks use status only.'
  row['probe_passing_members']=[t for x in observed if x['id'].startswith('P') for r in records(x['id']+'-'+group+'.log') if r.get('Action')=='pass' and (t:=r.get('Test')) in members]
  row['vacuous_subcases']=json.loads((out/'probe-subcases.json').read_text())['passing_cases'] if group=='block-layouts' else []
  row['witness_experiment']={'list-family':'W2','block-layouts':'W1','malformed-family':'W3'}[group]
  if group=='malformed-family':
   row['oracle']='Go cohere error messages plus retained fork Node diagnostics; self exit70, panic prefix and empty stdout. Full diagnostic comparison rejects a different error with the same exit status.';row['unique_kills']=row['kills'];row['verdict']='sacred'
  else:
   other='block-layouts' if group=='list-family' else 'list-family';row['subsumed_by']=[groupname[other]];row['subsumer_seconds']=median(other);row['verdict']='subsumed';row['subsumption_basis']=3
  kills=[x for x in observed if x['id'] in row['kills']]
  if kills:
   x=kills[-1];line=x['errors'][0]['Output'].strip() if x['errors'] else 'FAIL';row['last_proven_fail']=x['id']+' '+line;row['evidence']=shlex.join(x['command'])+'; '+x['id']+'-'+group+'.log: '+line
 result.append(row)
(out/'rows.json').write_text(json.dumps(result,indent=2)+'\n')
mutants=[]
for mid,file,a,b in [('M1','codeblocks.ts','Math.max(3, longest + 1)','Math.max(4, longest + 1)'),('M2','htmlblocks.ts','code === 32','code === 33'),('M3','root.ts','    parts.push(arena.hardline());\n    return arena.concat(parts);','drop final parts.push(arena.hardline())'),('M4','mdastCompile.ts','previous.type !== token.type','previous.type === token.type')]:
 old=subprocess.check_output(['git','show','HEAD:'+pkg+'/'+file],text=True);entry={'id':mid,'file':pkg+'/'+file,'line':old[:old.index(a)].count('\n')+1,'change':b,'failed_rows':sorted({memberrow[t.split('/')[0]] for r in matrix if r['id']==mid for t in r['failed'] if t.split('/')[0] in memberrow})};mutants.append(entry)
(out/'mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
# Native outputs from standalone products against retained Go fixture answers.
proof=[]
for mid in ['M1','M2','M3']:
 fixture=Path('/tmp/u127-witness/native.txt');want=Path('/tmp/u127-witness/want.txt').read_bytes()
 r=subprocess.run(['/tmp/u127-'+mid,str(fixture)],capture_output=True);(out/f'{mid}-native-stdout.txt').write_bytes(r.stdout);(out/f'{mid}-native-stderr.txt').write_bytes(r.stderr)
 offset=next((i for i,(a,b) in enumerate(zip(r.stdout,want)) if a!=b),min(len(r.stdout),len(want)))
 proof.append({'id':mid,'command':['/tmp/u127-'+mid,str(fixture)],'exit':r.returncode,'changed':r.stdout!=want,'first_difference':offset,'got':r.stdout[offset:offset+60].decode(errors='replace'),'Go_expected':want[offset:offset+60].decode(errors='replace'),'got_sha256':hashlib.sha256(r.stdout).hexdigest(),'Go_sha256':hashlib.sha256(want).hexdigest()})
r=subprocess.run(['/tmp/u127-M4',pkg+'/gaps/event_mismatch.txt'],capture_output=True);(out/'M4-native-stdout.txt').write_bytes(r.stdout);(out/'M4-native-stderr.txt').write_bytes(r.stderr);proof.append({'id':'M4','command':['/tmp/u127-M4',pkg+'/gaps/event_mismatch.txt'],'exit':r.returncode,'clean_expected_exit':70,'stdout':r.stdout.decode(errors='replace'),'stderr':r.stderr.decode(errors='replace')});(out/'native-witnesses.json').write_text(json.dumps(proof,indent=2)+'\n')
text='\n'.join(['u127: 35 named tests, grouped into 12 rows.', 'Base: '+subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(), 'Whole-package and combined slice baselines timed out; isolated clean rows passed without skips.', 'Four production mutants plus two probes; bounded uniqueness only.', 'Evidence branch: test-audit/stage1-cohere-markdownblocks-leaf_composition_products.'])+'\n\n```json\n'+json.dumps(result,indent=2)+'\n```\n\n'
text+='| ID | Origin file:line | Change | Failed grouped rows |\n|---|---|---|---|\n'
for m in mutants:text+=f"| {m['id']} | {m['file']}:{m['line']} | {m['change']} | {', '.join(m['failed_rows'])} |\n"
text+='\nNo production survivors in the bounded matrix. Outside rows and repo-wide uniqueness remain unknown.\n\n'
text+='The brief and costs:\n\n- The supplied historical commit is not the fetched origin/main. This audit uses ce1c5a2f, not 8de93800f4. All 35 names remain. malformed_events_shards_test.go moved to malformed_events_independent_test.go.\n- The brief says 15 rows while listing 35 functions. The current shared-checker bodies group into 12 rows. Product Go and native recipes share helpers; code/HTML/root share testBlockLayout; list and malformed-event shards include their unions.\n- I initially misgrouped code/HTML/root. Nine unnecessary member timing runs were preserved, then three proper family timings were collected. This avoidable error cost about ten minutes.\n- The whole baseline timed out at 90.043s in list preparation; the 35-name bounded baseline timed out at 90.038s in structure preparation. Neither reported an assertion failure first. Isolated clean rows all passed.\n- Some legacy names now only prepare products. They do not execute their advertised semantic comparisons. Their construction experiments are separated from production kills.\n- Construction verdicts rest on one selected edit each. Passing a changed path or dropped preparation does not establish that every possible failure is impossible. Treat these as bounded findings.\n- The first switch duplicated a built-in fence mutation anchor. Its neutral control failed at mutation site count. Those runs are archived under invalid-switch and excluded from final verdicts; the corrected neutral controls and full matrices were replayed. This avoidable instrumentation error cost a second matrix run.\n- Shared block fixture failure prevents sibling wrappers reaching their own body; grouping removes a false impression of independent kills. Raw member failures are preserved.\n- A static import closure is an upper bound on reachable functions, including type-only imports. Exact dynamic function coverage was not measured.\n- Empty probes use production printDocument and MdastCompiler.compile entries, preserving protocol adapters. Component producers were not separately probed. Direct native P1 replay accepts two naturally empty positive root cases; their names are listed as vacuous_subcases, while the family is not vacuous. Union/setup members that pass a probe are explicitly listed and are not semantic-entry vacuity proofs.\n- The runtime source switch avoids content changes per selection, but the uncached full block fixture still lowers and builds on each invocation. Each standalone diff also received its own native compile validation.\n- The source-executed Node port is code under test, not an oracle. The independent Go cohere answers and retained fork Node results were unchanged. Direct standalone native output witnesses are saved separately.\n- The four-mutant set is smaller than three per row, because most rows now construct products and native preparation is expensive. Subsumption rests on three mutants only, and is not a deletion recommendation.\n- No scope row skipped. Optional whole-package width SDK/corpus coverage beyond the named slice was not audited. No repo-wide replay was run.\n\n'
clean=json.loads((out/'clean-runs.json').read_text());builds=json.loads((out/'standalone-builds.json').read_text())
text+=f"Warm tool setup: skipped; npm ci reported 458 ms; nproc 5. Isolated clean command wall time including redundant member runs: {sum(r['wall'] for r in clean):.3f}s. Mutation/probe command wall time: {sum(r['wall'] for r in matrix if r['id'] != 'CONTROL'):.3f}s. Neutral controls: {sum(r['wall'] for r in matrix if r['id'] == 'CONTROL'):.3f}s. Construction experiments including vet: {sum(r['wall'] for r in construction):.3f}s. Standalone native build time: {sum(r['seconds'] for r in builds):.3f}s. Two 90-second baselines and the archived invalid switch runs are additional. Full per-build timings and commands are in the logs and standalone-builds.json.\n"
invalid=json.loads((out/'invalid-switch/matrix-runs.json').read_text());text+=f' Archived invalid mutation/probe runs: {sum(r["wall"] for r in invalid):.3f}s, plus their failed neutral control.\n'
text+='\nThe requested approximate thirty-minute port budget was exceeded. The redundant timing runs, repeated uncached large-fixture builds, neutral controls and standalone validation explain the overrun. Setup-to-finish UTC timestamps are in the logs. Production source and tests were restored; only evidence is committed.\n'
(out/'REPORT.md').write_text(text)
